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
	"encoding/hex"
	"fmt"
	"math"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"

	"github.com/arcology-network/eu/abi"
	eucommon "github.com/arcology-network/eu/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	eth "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
	schtype "github.com/arcology-network/scheduler"
	stgcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/noncommutative"
)

type RuntimeHandlers struct {
	api         intf.EthApiRouter
	pathBuilder *eth.PathBuilder
}

func NewRuntimeHandlers(ethApiRouter intf.EthApiRouter) *RuntimeHandlers {
	return &RuntimeHandlers{
		api:         ethApiRouter,
		pathBuilder: eth.NewPathBuilder("/storage", ethApiRouter),
	}
}

// To register the runtime handler to the API router.
func (this *RuntimeHandlers) Address() [20]byte { return eucommon.RUNTIME_HANDLER }

func (this *RuntimeHandlers) writeCache(path string, val stgcommon.Type, gasMeter *eucommon.GasMeter) error {
	txID, cache := this.api.GetTxContext()
	writeDataSize, err := cache.Write(txID, path, val)
	gasMeter.Use(0, writeDataSize, 0)
	return err // Return the error if any.
}

func (this *RuntimeHandlers) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, isReadOnly bool) ([]byte, bool, int64) {
	signature := codec.Bytes4{}.FromBytes(input[:])

	switch signature {
	case [4]byte{0xf1, 0x06, 0x84, 0x54}: // 79 fc 09 a2
		return this.pid(caller, input[4:])

	// case [4]byte{0x64, 0x23, 0xdb, 0x34}: // d3 01 e8 fe
	// return this.rollback(caller, input[4:])

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
	encoded, err := abi.Encode(job.(*eucommon.Job).StdMsg.IsDeferred)
	return encoded, err == nil, eucommon.GAS_ENCODE + eucommon.GAS_GET_RUNTIME_INFO
}

func (this *RuntimeHandlers) setParallelism(caller, addr evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()
	paraLvl, err := abi.Decode(input, 3, uint64(0), 1, 1)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE) // Gas for decoding the input

	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	executionMethod := stgcommon.PARALLEL_EXECUTION
	if paraLvl == 1 {
		executionMethod = stgcommon.SEQUENTIAL_EXECUTION // If the parallelism level is 1, set the execution method to sequential.
	}

	result, successful, gas := this.setExecutionParallelism(caller, addr, input, executionMethod)
	gasMeter.Use(0, 0, gas) // Add the gas used for setting the execution method.

	return result, successful, gasMeter.TotalGasUsed
}

func (this *RuntimeHandlers) setExecutionParallelism(caller, _ evmcommon.Address, input []byte, executionMethod uint8) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()
	funcSign, err := abi.DecodeTo(input, 0, [4]byte{}, 1, 4) // Get the target contract address.
	gasMeter.Use(0, 0, eucommon.GAS_GET_RUNTIME_INFO+eucommon.GAS_DECODE)
	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	targetAddr, err := abi.DecodeTo(input, 1, [20]byte{}, 1, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)

	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Get the target function signatures
	signBytes, err := abi.DecodeTo(input, 2, []byte{}, 1, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)

	if err != nil || len(signBytes) <= 32 {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Parse the function signatures.
	signatures, err := abi.DecodeTo(signBytes[32:], 0, [][4]byte{}, 2, math.MaxInt)
	gasMeter.Use(0, 0, eucommon.GAS_DECODE)
	if err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Check if the property path exists, if not create it.
	funcPath := stgcommon.FuncPath(caller, funcSign)
	if _, cache := this.api.GetTxContext(); !cache.IfExists(funcPath) {
		if err := this.writeCache(funcPath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed // If the property path write fails, return an error.
		}
	}

	// If local method is parallel, global method is sequential and vice versa.
	// How the scheduler all the function under the contract should be executed in parallel or sequentially by DEFAULT.
	globalMethod := stgcommon.PARALLEL_EXECUTION
	if executionMethod == stgcommon.PARALLEL_EXECUTION {
		globalMethod = stgcommon.SEQUENTIAL_EXECUTION
	}

	if err := this.writeCache(stgcommon.ExecutionParallelism(caller, funcSign), noncommutative.NewBytes([]byte{globalMethod}), gasMeter); err != nil {
		return []byte{}, false, gasMeter.TotalGasUsed
	}

	// Users can add some excepted callees so they can be handled differently.
	callees := slice.Transform(signatures, func(i int, signature [4]byte) string { // Get the excepted callees.
		return hex.EncodeToString(schtype.Compact(targetAddr[:], signature[:]))
	})

	err = this.writeCache(stgcommon.ExceptPaths(caller, funcSign), commutative.NewPath(callees...), gasMeter)
	return []byte{}, err == nil, gasMeter.TotalGasUsed
}

// This function inform the scheduler to scheduler a defer call for a particular function.
func (this *RuntimeHandlers) deferCall(caller, callee evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO // Can only be called from a constructor.
	}

	gasMeter := eucommon.NewGasMeter()

	// Decode the function signature from the input.
	funSignBytes, err := abi.DecodeTo(input, 0, []uint8{}, 1, 32)
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
	funSign := new(codec.Bytes4).FromBytes(funSignBytes)
	propertyPath := stgcommon.FuncPath(caller, funSign)
	if _, cache := this.api.GetTxContext(); !cache.IfExists(propertyPath) {
		if err := this.writeCache(propertyPath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed
		}
	}

	// Create the prepayer path if not existent.
	txID, cache := this.api.GetTxContext()
	prepayerPath := stgcommon.PrepayersPath(caller, funSign)
	if v, _, _ := cache.Read(txID, prepayerPath, new(commutative.Path)); v == nil {
		// Create the full function path.
		if err := this.writeCache(prepayerPath, commutative.NewPath(), gasMeter); err != nil {
			return []byte{}, false, gasMeter.TotalGasUsed
		}
	}

	// Write the required prepaid amount to storage
	RequiredPrepaymentPath := stgcommon.RequiredPrepaymentPath(caller, funSign) // Generate the sub path for the prepaid gas amount.
	if err := this.writeCache(RequiredPrepaymentPath, noncommutative.NewUint64(requiredPrepayment.(uint64)), gasMeter); err != nil {
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
