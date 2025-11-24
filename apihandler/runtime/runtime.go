/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package runtime

import (
	"fmt"
	"math"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"

	"github.com/arcology-network/eu/abi"
	eucommon "github.com/arcology-network/eu/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	crdtcommon "github.com/arcology-network/common-lib/crdt/common"
	"github.com/arcology-network/common-lib/crdt/commutative"
	"github.com/arcology-network/common-lib/crdt/noncommutative"
	ethadaptor "github.com/arcology-network/eu/ethadaptor"
	intf "github.com/arcology-network/eu/interface"
	statecommon "github.com/arcology-network/state-engine/common"

	workload "github.com/arcology-network/scheduler/workload"
)

type RuntimeHandlers struct {
	api         intf.EthApiRouter
	pathBuilder *ethadaptor.ContainerPathBuilder
}

func NewRuntimeHandlers(ethApiRouter intf.EthApiRouter) *RuntimeHandlers {
	return &RuntimeHandlers{
		api:         ethApiRouter,
		pathBuilder: ethadaptor.NewPathBuilder("/storage", ethApiRouter),
	}
}

// To register the runtime handler to the API router.
func (this *RuntimeHandlers) Address() [20]byte { return eucommon.RUNTIME_HANDLER }

func (this *RuntimeHandlers) writeCache(path string, val crdtcommon.Type, gasMeter *eucommon.GasMeter) error {
	txID, cache := this.api.GetTxContext()
	writeDataSize, err := cache.Write(txID, path, val)
	gasMeter.Use(0, writeDataSize, 0)
	return err // Return the error if any.
}

func (this *RuntimeHandlers) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, isReadOnly bool) ([]byte, bool, int64) {
	selector := codec.Bytes4{}.FromBytes(input[:])

	switch selector {
	case [4]byte{0xf1, 0x06, 0x84, 0x54}: // 79 fc 09 a2
		return this.pid(caller, input[4:])

	case [4]byte{0xbb, 0x07, 0xe8, 0x5d}: // bb 07 e8 5d
		return this.uuid(caller, callee, input[4:])

	case [4]byte{0x0f, 0x0d, 0x97, 0xaa}: //
		return this.setParallelism(caller, callee, input[4:])

	case [4]byte{0xac, 0x8f, 0x58, 0xf3}: // 1c 2f 3b 6d	case [4]byte{0xac, 0x8f, 0x58, 0xf3}: // 19 7f 62 5f
		return this.deferCall(caller, callee, input[4:])

	case [4]byte{0x21, 0xcb, 0x6b, 0xc3}: // bb 07 e8 5d
		return this.isInDeferred(caller, callee, input[4:])

	case [4]byte{0x37, 0x66, 0x82, 0xb5}: // 19 7f 62 5f
		return this.print(caller, callee, input[4:])
	}

	fmt.Println("Function not found !!")
	fmt.Println(input)
	return []byte{}, false, eucommon.GAS_CALL_UNKNOW
}

func (this *RuntimeHandlers) pid(_ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	encoded, err := abi.Encode(this.api.Pid())
	return encoded, err == nil, eucommon.GAS_DECODE + eucommon.GAS_GET_RUNTIME_INFO
}

func (this *RuntimeHandlers) uuid(_, _ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	return this.api.ElementUID(), true, eucommon.GAS_GET_RUNTIME_INFO
}

// Get the number of running instances of a function.
func (this *RuntimeHandlers) isInDeferred(_ evmcommon.Address, _ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	job := this.api.VM().(*vm.EVM).ArcologyAPIs.Job()
	encoded, err := abi.Encode(job.(*workload.Job).StdMsg.IsDeferred)
	return encoded, err == nil, eucommon.GAS_ENCODE + eucommon.GAS_GET_RUNTIME_INFO
}

func (this *RuntimeHandlers) setParallelism(caller, addr evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()
	paraLvl, err := abi.Decode(input, 3, uint32(0), 1, 1)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE) // Gas for decoding the input

	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	result, successful, gas := this.setExecutionParallelism(caller, addr, input, paraLvl.(uint32))
	gasMeter.Use(0, 0, gas) // Add the gas used for setting the execution method.

	return result, successful, gasMeter.TotalGasUsed
}

func (this *RuntimeHandlers) setExecutionParallelism(caller, _ evmcommon.Address, input []byte, degree uint32) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()
	selector, err := abi.DecodeTo(input, 0, [4]byte{}, 1, 4) // Get the target contract address.
	gasMeter.Use(0, 0, eucommon.GAS_GET_RUNTIME_INFO+eucommon.GAS_DECODE)
	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	_, err = abi.DecodeTo(input, 1, [20]byte{}, 1, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)

	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Get the target function selectors
	selectorBytes, err := abi.DecodeTo(input, 2, []byte{}, 1, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)

	if err != nil || len(selectorBytes) <= 32 {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Parse the function selectors.
	_, err = abi.DecodeTo(selectorBytes[32:], 0, [][4]byte{}, 2, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)
	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Check if the property path exists, if not create it.
	funcPath := (&statecommon.PathBuilder{caller, selector, statecommon.ETH_PATH}).ProfileField("")
	if _, cache := this.api.GetTxContext(); !cache.IfExists(funcPath) {
		if err := this.writeCache(funcPath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed // If the property path write fails, return an error.
		}
	}

	// blcc://eth1.0/account/[0x...]/profiles/paraDegree
	path := (&statecommon.PathBuilder{caller, selector, statecommon.ETH_PATH}).ProfileField(statecommon.PARALLELISM_DEGREE)
	v := noncommutative.NewUint32(uint32(degree))
	if err := this.writeCache(path, v, gasMeter); err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}
	return []byte{}, err == nil, gasMeter.TotalGasUsed
}

// This function inform the scheduler to scheduler a defer call for a particular function.
func (this *RuntimeHandlers) deferCall(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()

	// Decode the function selector from the input.
	selectorBytes, err := abi.DecodeTo(input, 0, []uint8{}, 1, 32)
	gasMeter.Use(0, 0, eucommon.GAS_GET_RUNTIME_INFO+eucommon.GAS_DEFER) // Gas for deferring the call.

	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Decode the amount of prepaid gas from the input.
	requiredPrepayment, err := abi.Decode(input, 1, uint64(0), 1, 8)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)
	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}
	// No less than GAS_MIN_PREPAYMENT.
	requiredPrepayment = common.Max(requiredPrepayment.(uint64), eucommon.GAS_MIN_PREPAYMENT)

	// Check if the function path exists, if not create it.
	// It may be created by the developer in setting the parallelism
	// level in the constructor as well.
	selector := new(codec.Bytes4).FromBytes(selectorBytes)
	profilePath := (&statecommon.PathBuilder{caller, selector, statecommon.ETH_PATH}).ProfileField("")
	if _, cache := this.api.GetTxContext(); !cache.IfExists(profilePath) {
		if err := this.writeCache(profilePath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed
		}
	}

	// Create the prepayer path if not existent.
	txID, cache := this.api.GetTxContext()
	prepayerPath := (&statecommon.PathBuilder{caller, selector, statecommon.ETH_PATH}).ProfileField(statecommon.PREPAYERS)
	if v, _, _ := cache.Read(txID, prepayerPath, new(commutative.Path)); v == nil {
		// Create the full function path.
		if err := this.writeCache(prepayerPath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed
		}
	}

	// Write the required prepaid amount to storage
	path := (&statecommon.PathBuilder{caller, selector, statecommon.ETH_PATH}).ProfileField(statecommon.DEFERRED_PAYMENT)
	if err := this.writeCache(path, noncommutative.NewUint64(requiredPrepayment.(uint64)), gasMeter); err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}
	return []byte{}, true, gasMeter.TotalGasUsed
}

func (this *RuntimeHandlers) print(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	msg, err := abi.DecodeTo(input, 2, []uint8{}, 1, math.MaxInt)

	msg = evmcommon.TrimRightZeroes(msg)
	fmt.Println("From=", caller, " Msg=", (msg), " Error=", err)
	return []byte{}, true, eucommon.GAS_SET_RUNTIME_INFO * 10
}
