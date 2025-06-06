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
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

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

	case [4]byte{0x64, 0x23, 0xdb, 0x34}: // d3 01 e8 fe
		return this.rollback(caller, input[4:])

	case [4]byte{0xbb, 0x07, 0xe8, 0x5d}: // bb 07 e8 5d
		return this.uuid(caller, callee, input[4:])

	case [4]byte{0x68, 0x7b, 0x09, 0xb7}: //
		return this.setExecutionMethod(caller, callee, input[4:], stgcommon.SEQUENTIAL_EXECUTION)

	case [4]byte{0xc4, 0xdf, 0xfe, 0x6e}: //
		return this.setExecutionMethod(caller, callee, input[4:], stgcommon.PARALLEL_EXECUTION)

	case [4]byte{0xa8, 0x7a, 0xe4, 0x81}: // bb 07 e8 5d
		return this.instances(caller, callee, input[4:])

	case [4]byte{0x19, 0x7f, 0x62, 0x5f}: // 19 7f 62 5f
		return this.deferCall(caller, callee, input[4:])

	case [4]byte{0x37, 0x66, 0x82, 0xb5}: // 19 7f 62 5f
		return this.print(caller, callee, input[4:])
	}

	fmt.Println(input)
	return []byte{}, false, 0
}

func (this *RuntimeHandlers) pid(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if encoded, err := abi.Encode(this.api.Pid()); err == nil {
		return encoded, true, 0
	}
	return []byte{}, false, 0
}

// This function rolls back the storage to the previous generation. It should be used with extreme caution.
func (this *RuntimeHandlers) rollback(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	tempcache.NewWriteCacheFilter(this.api.WriteCache()).RemoveByAddress(codec.Bytes20(caller).Hex())
	return []byte{}, true, 0
}

func (this *RuntimeHandlers) uuid(caller, callee evmcommon.Address, input []byte) ([]byte, bool, int64) {
	return this.api.ElementUID(), true, 0
}

// Get the number of running instances of a function.
func (this *RuntimeHandlers) instances(caller evmcommon.Address, callee evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if this.api.GetSchedule() == nil {
		return []byte{}, false, 0
	}

	address, err := abi.DecodeTo(input, 0, [20]byte{}, 1, 4)
	if err != nil {
		return []byte{}, false, 0
	}

	funSign, err := abi.DecodeTo(input, 1, []byte{}, 1, 4)
	if err != nil {
		return []byte{}, false, 0
	}

	dict := this.api.GetSchedule().(*map[string]int)
	key := schtype.CallToKey(address[:], funSign)

	// Encode the total number of instances and return
	if encoded, err := abi.Encode(uint256.NewInt(uint64((*dict)[key]))); err == nil {
		// encoded, _ := abi.Encode(uint256.NewInt(2))
		// if !bytes.Equal(encoded, encoded2) {
		// 	panic("")
		// }

		return encoded, true, 0
	}
	return []byte{}, false, 0
}

func (this *RuntimeHandlers) setExecutionMethod(caller, _ evmcommon.Address, input []byte, executionMethod uint8) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyNetworkAPIs.IsInConstructor() {
		return []byte{}, false, 0 // Can only be called from a constructor.
	}

	// Get the target contract address.
	sourceFunc, err := abi.DecodeTo(input, 0, [4]byte{}, 1, 4)
	if err != nil {
		return []byte{}, false, 0
	}

	targetAddr, err := abi.DecodeTo(input, 1, [20]byte{}, 1, math.MaxInt)
	if err != nil {
		return []byte{}, false, 0
	}

	// Get the target function signatures
	signBytes, err := abi.DecodeTo(input, 2, []byte{}, 1, math.MaxInt)
	if err != nil || len(signBytes) <= 32 {
		return []byte{}, false, 0
	}

	// Parse the function signatures.
	signatures, err := abi.DecodeTo(signBytes[32:], 0, [][4]byte{}, 2, math.MaxInt)
	if err != nil {
		return []byte{}, false, 0
	}

	tempcache := this.api.WriteCache().(*tempcache.WriteCache)
	txID := this.api.GetEU().(interface{ ID() uint64 }).ID()

	// Create the parent path for the properties.
	propertyPath := stgcommon.FuncPropertyPath(caller, sourceFunc)
	if _, err = tempcache.Write(txID, propertyPath, commutative.NewPath()); err != nil {
		return []byte{}, err == nil, 0
	}

	// Either the function is parallel or sequential.
	path := stgcommon.ExecutionMethodPath(caller, sourceFunc)

	// If local method is parallel, global method is sequential and vice versa.
	// How the scheduler all the function under the contract should be executed in parallel or sequentially by DEFAULT.
	globalMethod := stgcommon.PARALLEL_EXECUTION
	if executionMethod == stgcommon.PARALLEL_EXECUTION {
		globalMethod = stgcommon.SEQUENTIAL_EXECUTION
	}

	if _, err = tempcache.Write(txID, path, noncommutative.NewBytes([]byte{globalMethod})); err != nil { //
		return []byte{}, false, 0
	}

	// Users can add some excepted callees so they can be handled differently.
	callees := slice.Transform(signatures, func(i int, signature [4]byte) string { // Get the excepted callees.
		return hex.EncodeToString(schtype.Compact(targetAddr[:], signature[:]))
	})

	path = stgcommon.ExceptPaths(caller, sourceFunc)
	_, err = tempcache.Write(txID, path, commutative.NewPath(callees...)) // Write the excepted callees regardless of its existence.
	return []byte{}, err == nil, 0
}

// This function needs to schedule a defer call to the next generation.
func (this *RuntimeHandlers) deferCall(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	if !this.api.VM().(*vm.EVM).ArcologyNetworkAPIs.IsInConstructor() {
		return []byte{}, false, 0 // Can only be called from a constructor.
	}

	if len(input) < 4 {
		return []byte{}, false, 0
	}

	tempcache := this.api.WriteCache().(*tempcache.WriteCache)

	funSign := new(codec.Bytes4).FromBytes(input[:4])
	txID := this.api.GetEU().(interface{ ID() uint64 }).ID()

	// Get the function signature.
	propertyPath := stgcommon.FuncPropertyPath(caller, funSign)
	tempcache.Write(txID, propertyPath, commutative.NewPath()) // Create the property path only when needed.

	deferPath := stgcommon.DeferrablePath(caller, funSign)                           // Generate the sub path for the deferrable.
	_, err := tempcache.Write(txID, deferPath, noncommutative.NewBytes([]byte{255})) // Set the function deferrable
	return []byte{}, err == nil, 0
}

func (this *RuntimeHandlers) print(caller, _ evmcommon.Address, input []byte) ([]byte, bool, int64) {
	msg, err := abi.DecodeTo(input, 2, []uint8{}, 1, math.MaxInt)

	msg = evmcommon.TrimRightZeroes(msg)
	fmt.Println("From=", caller, " Msg=", (msg), " Error=", err)
	return []byte{}, true, 0
}
