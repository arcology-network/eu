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
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/abi"
	eucommon "github.com/arcology-network/eu/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

	adaptorcommon "github.com/arcology-network/eu/common"
	eth "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
	schtype "github.com/arcology-network/scheduler"
	stgcommon "github.com/arcology-network/storage-committer/common"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
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

func (this *RuntimeHandlers) Address() [20]byte {
	return adaptorcommon.RUNTIME_HANDLER
}

func (this *RuntimeHandlers) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, isReadOnly bool) ([]byte, bool, int64) {
	signature := [4]byte{}
	copy(signature[:], input)

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
	if encoded, err := abi.Encode(this.api.Pid()); err == nil {
		return encoded, true, eucommon.GAS_DECODE
	}
	return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO
}

// This function rolls back the storage to the previous generation. It should be used with extreme caution.
// func (this *RuntimeHandlers) rollback(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	tempcache.NewWriteCacheFilter(this.api.WriteCache()).RemoveByAddress(codec.Bytes20(caller).Hex())
// 	return []byte{}, true, 0
// }

func (this *RuntimeHandlers) uuid(_, _ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	return this.api.ElementUID(), true, eucommon.GAS_GET_RUNTIME_INFO
}

// Get the number of running instances of a function.
func (this *RuntimeHandlers) isInDeferred(_ evmcommon.Address, _ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	job := this.api.VM().(*vm.EVM).ArcologyAPIs.Job()
	if encoded, err := abi.Encode(job.(*eucommon.Job).IsDeferred); err == nil {
		return encoded, true, eucommon.GAS_DECODE + eucommon.GAS_GET_RUNTIME_INFO
	}

	return []byte{}, false, eucommon.GAS_GET_RUNTIME_INFO
}

func (this *RuntimeHandlers) setParallelism(caller, addr evmcommon.Address, input []byte) ([]byte, bool, int64) {
	paraLvl, err := abi.Decode(input, 3, uint64(0), 1, 1)
	if err != nil {
		return []byte{}, false, eucommon.GAS_DECODE
	}

	executionMethod := stgcommon.PARALLEL_EXECUTION
	if paraLvl == 1 {
		executionMethod = stgcommon.SEQUENTIAL_EXECUTION // If the parallelism level is 1, set the execution method to sequential.
	}
	return this.setExecutionMethod(caller, addr, input, uint8(executionMethod))
}

func (this *RuntimeHandlers) setExecutionMethod(caller, _ evmcommon.Address, input []byte, executionMethod uint8) ([]byte, bool, int64) {
	totalFee := eucommon.GAS_GET_RUNTIME_INFO
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, totalFee // Can only be called from a constructor.
	}

	// Get the target contract address.
	sourceFunc, err := abi.DecodeTo(input, 0, [4]byte{}, 1, 4)
	if err != nil {
		return []byte{}, false, eucommon.GAS_READ + eucommon.GAS_DECODE
	}

	targetAddr, err := abi.DecodeTo(input, 1, [20]byte{}, 1, math.MaxInt)
	if err != nil {
		return []byte{}, false, eucommon.GAS_READ + eucommon.GAS_DECODE*2
	}

	// Get the target function signatures
	signBytes, err := abi.DecodeTo(input, 2, []byte{}, 1, math.MaxInt)
	if err != nil || len(signBytes) <= 32 {
		return []byte{}, false, eucommon.GAS_READ + eucommon.GAS_DECODE*3
	}

	// Parse the function signatures.
	signatures, err := abi.DecodeTo(signBytes[32:], 0, [][4]byte{}, 2, math.MaxInt)
	if err != nil {
		return []byte{}, false, eucommon.GAS_READ + eucommon.GAS_DECODE*4
	}

	tempcache := this.api.WriteCache().(*tempcache.WriteCache)
	propertyPath := stgcommon.FuncPropertyPath(caller, sourceFunc)
	txID := this.api.GetEU().(interface{ ID() uint64 }).ID()

	// Check if the property path exists, if not create it.
	if path, _, _ := tempcache.Read(txID, propertyPath, commutative.NewPath()); path == nil {
		_, err := tempcache.Write(txID, propertyPath, commutative.NewPath()) // Create the property path only when needed.
		totalFee += eucommon.GAS_WRITE
		if err != nil {
			return []byte{}, false, totalFee // If the property path write fails, return an error.
		}
	}

	// Create the parent path for the properties.
	// writePathFee, err := tempcache.Write(txID, propertyPath, commutative.NewPath())
	// if err != nil {
	// 	return []byte{}, err == nil, eucommon.GAS_READ + eucommon.GAS_DECODE*4 + writePathFee
	// }

	// Either the function is parallel or sequential.
	path := stgcommon.ExecutionMethodPath(caller, sourceFunc)

	// If local method is parallel, global method is sequential and vice versa.
	// How the scheduler all the function under the contract should be executed in parallel or sequentially by DEFAULT.
	globalMethod := stgcommon.PARALLEL_EXECUTION
	if executionMethod == stgcommon.PARALLEL_EXECUTION {
		globalMethod = stgcommon.SEQUENTIAL_EXECUTION
	}

	_, err = tempcache.Write(txID, path, noncommutative.NewBytes([]byte{globalMethod}))
	if err != nil { //
		return []byte{}, false, totalFee
	}

	// Users can add some excepted callees so they can be handled differently.
	callees := slice.Transform(signatures, func(i int, signature [4]byte) string { // Get the excepted callees.
		return hex.EncodeToString(schtype.Compact(targetAddr[:], signature[:]))
	})

	path = stgcommon.ExceptPaths(caller, sourceFunc)
	_, err = tempcache.Write(txID, path, commutative.NewPath(callees...)) // Write the excepted callees regardless of its existence.
	return []byte{}, err == nil, totalFee
}

// This function needs to schedule a defer call to the next generation.

func (this *RuntimeHandlers) deferCall(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	totalFee := eucommon.GAS_DEFER + eucommon.GAS_GET_RUNTIME_INFO
	if !this.api.VM().(*vm.EVM).ArcologyAPIs.IsInConstructor() {
		return []byte{}, false, totalFee // Can only be called from a constructor.
	}

	// Decode the function signature from the input.
	funSignBytes, err := abi.DecodeTo(input, 0, []uint8{}, 1, 32)
	totalFee += eucommon.GAS_DECODE
	if err != nil {
		return []byte{}, false, totalFee
	}

	// Decode the amount of prepaid gas from the input.
	prepaidGas, err := abi.Decode(input, 1, uint64(0), 1, 8)
	totalFee += eucommon.GAS_DECODE
	if err != nil {
		return []byte{}, false, totalFee
	}

	funSign := new(codec.Bytes4).FromBytes(funSignBytes)
	txID := this.api.GetEU().(interface{ ID() uint64 }).ID()

	// Get the function signature.
	propertyPath := stgcommon.FuncPropertyPath(caller, funSign)
	tempcache := this.api.WriteCache().(*tempcache.WriteCache)

	// Check if the property path exists, if not create it.
	if path, _, _ := tempcache.Read(txID, propertyPath, commutative.NewPath()); path == nil {
		_, err := tempcache.Write(txID, propertyPath, commutative.NewPath()) // Create the property path only when needed.
		totalFee += eucommon.GAS_WRITE
		if err != nil {
			return []byte{}, false, totalFee // If the property path write fails, return an error.
		}
	}

	// Write deferrable information to the property path.
	deferPath := stgcommon.DeferrablePath(caller, funSign)                          // Generate the sub path for the deferrable.
	_, err = tempcache.Write(txID, deferPath, noncommutative.NewBytes([]byte{255})) // Set the function deferrable
	totalFee += eucommon.GAS_WRITE
	if err != nil {
		return []byte{}, false, totalFee
	}

	job := this.api.VM().(*vm.EVM).ArcologyAPIs.Job()
	totalFee += eucommon.GAS_GET_RUNTIME_INFO
	job.(*eucommon.Job).PrepaidGas = prepaidGas.(uint64) // Set the prepaid gas for the deferred call.
	return []byte{}, err == nil, totalFee
}

func (this *RuntimeHandlers) print(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	msg, err := abi.DecodeTo(input, 2, []uint8{}, 1, math.MaxInt)

	msg = evmcommon.TrimRightZeroes(msg)
	fmt.Println("From=", caller, " Msg=", (msg), " Error=", err)
	return []byte{}, true, eucommon.GAS_SET_RUNTIME_INFO
}
