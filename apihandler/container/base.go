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

package api

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"math/big"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/types"

	// "github.com/arcology-network/common-lib/exp/deltaset"

	abi "github.com/arcology-network/eu/abi"
	"github.com/arcology-network/eu/common"
	eth "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
	stgtype "github.com/arcology-network/storage-committer/common"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	evmcommon "github.com/ethereum/go-ethereum/common"

	"github.com/holiman/uint256"
)

// APIs under the concurrency namespace
type BaseHandlers struct {
	api         intf.EthApiRouter
	pathBuilder *eth.PathBuilder
	args        []any
}

func NewBaseHandlers(api intf.EthApiRouter, args ...any) *BaseHandlers {
	return &BaseHandlers{
		api:         api,
		pathBuilder: eth.NewPathBuilder("/storage/container", api),
		args:        args,
	}
}

func (this *BaseHandlers) Address() [20]byte           { return common.BYTES_HANDLER }
func (this *BaseHandlers) Connector() *eth.PathBuilder { return this.pathBuilder }

func (this *BaseHandlers) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, isReadOnly bool) ([]byte, bool, int64) {
	signature := [4]byte{}
	copy(signature[:], input)

	// Read-only functions that won't change the state of the container
	if isReadOnly {
		switch signature {
		case [4]byte{0x59, 0xe0, 0x2d, 0xd7}:
			return this.committedLength(caller, input[4:]) // Get the initial length of the container, it remains the same in the same block.

		case [4]byte{0x86, 0x03, 0x9d, 0x78}:
			return this.fullLength(caller, input[4:]) // Get the current number of elements in the container.

		case [4]byte{0x1f, 0x7b, 0x6d, 0x32}:
			return this.length(caller, input[4:]) // Get the current number of elements in the container, excluding the nil elements.

		case [4]byte{0x91, 0x1f, 0x6f, 0xe0}:
			return this.keyToInd(caller, input[4:]) // Get the index of the element by its key.

		case [4]byte{0x06, 0xed, 0x32, 0x3c}:
			return this.indToKey(caller, input[4:]) // Get the key of the element by its index.

		case [4]byte{0x6b, 0x19, 0xdf, 0x9b}:
			return this.getByKey(caller, input[4:]) // Get the element by its key.

		case [4]byte{0x2d, 0x88, 0x3a, 0x73}:
			return this.getByIndex(caller, input[4:]) // Get the element by its key.

		case [4]byte{0xd3, 0x32, 0x51, 0x6f}:
			return this.minNumerical(caller, input[4:]) // Delete the element by its key.

		case [4]byte{0xd6, 0x99, 0x5f, 0x76}:
			return this.maxNumerical(caller, input[4:]) // Delete the element by its key.
		}
	} else {
		// Write functions will change the state of the container
		switch signature {

		case [4]byte{0x66, 0x54, 0x85, 0x21}:
			return this.new(caller, input[4:]) // Create a new container

		case [4]byte{0xe5, 0xe2, 0x14, 0xb5}:
			return this.init(caller, input[4:]) // Set the bounds of the elements in the container.

		case [4]byte{0xf1, 0x06, 0x84, 0x54}:
			return this.pid(caller, input[4:]) // Get the pesudo process ID.

		case [4]byte{0x8a, 0xd4, 0xeb, 0xf6}:
			return this.setByKey(caller, input[4:]) // Set the element by its key.

		case [4]byte{0x55, 0x46, 0x09, 0xea}:
			return this.delByKey(caller, input[4:]) // Delete the element by its key.
		//
		case [4]byte{0x94, 0x42, 0x8e, 0x6a}:
			return this.resetByKey(caller, input[4:]) // Delete the element by its key.

		case [4]byte{0x02, 0x07, 0x83, 0x86}:
			return this.resetByInd(caller, input[4:]) // Delete the element by its index.

		case [4]byte{0x0b, 0xe0, 0xc6, 0xd5}:
			return this.delLast(caller, input[4:]) // shrink the size of the container by one

		case [4]byte{0x52, 0xef, 0xea, 0x6e}:
			return this.clear(caller, input[4:]) // Clear the container.
		}
	}

	// Custom function call. The base handler may have a custom function to call..
	if len(this.args) > 0 {
		customFun := this.args[0].(func([20]byte, [20]byte, []byte, ...any) ([]byte, bool, int64))
		return customFun(caller, callee, input[4:], this.args[1:]...)
	}

	return []byte{}, false, 0 // unknown
}

func (this *BaseHandlers) Api() intf.EthApiRouter { return this.api }

func (this *BaseHandlers) new(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	addr := codec.Bytes20(caller).Hex()
	connected, pathStr := this.pathBuilder.New(
		this.api.GetEU().(interface{ ID() uint64 }).ID(), // Tx ID for conflict detection
		types.Address(addr), // Main contract address
	)

	// Add the type info to the container here.
	path, _ := this.api.WriteCache().(*tempcache.WriteCache).PeekRaw(pathStr, commutative.Path{})

	if typeID, err := abi.Decode(input, 0, uint8(0), 1, 32); len(input) != 0 && err == nil {
		path.(*commutative.Path).Type = typeID.(uint8)
	}

	this.api.SetDeployer(caller)   // Store the MP address to the API
	return caller[:], connected, 0 // Create a new container
}

func (this *BaseHandlers) init(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	// Check if the container exists
	path := this.pathBuilder.Key(caller)
	if !this.api.WriteCache().(*tempcache.WriteCache).IfExists(path) {
		return []byte{}, false, int64(0) // Doesn't exist, cannot initialize in a non-existent container.
	}

	// If the key already exists
	if _, ok, fee := this.getByKey(caller, []byte{}); ok {
		return []byte{}, false, fee
	}

	//Get the type info here
	pathData, fee := this.api.WriteCache().(*tempcache.WriteCache).PeekRaw(path, commutative.Path{})
	if pathData == nil {
		return []byte{}, false, int64(fee)
	}

	// Check if it is a cumulative container. Only cumulative elements can be initialized.
	// If it is, decode the lower and upper bounds
	if pathData.(*commutative.Path).Type == commutative.UINT256 {
		abiDef := `[{
			"name": "init",
			"inputs": [
				{"name": "", "type": "bytes"},
				{"name": "", "type": "bytes"},
				{"name": "", "type": "bytes"}
			],
			"stateMutability": "nonpayable",
			"type": "function"
		}]`

		// Pass pointers to variables so they can be decoded into
		var key, min, max []byte
		abi.DecodeEth(abiDef, "0x"+hex.EncodeToString(input), "init", []any{&key, &min, &max})

		// Initialize the element with the lower and upper bounds
		minv, maxv := uint256.NewInt(0).SetBytes(min), uint256.NewInt(0).SetBytes(max)
		if minv.Cmp(maxv) > 0 {
			return []byte{}, false, 0 // The lower bound is greater than the upper bound
		}
		v := commutative.NewBoundedU256(minv, maxv)

		str := hex.EncodeToString(key)
		fee, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.api.GetEU().(interface{ ID() uint64 }).ID(), path+str, v)
		return []byte{}, err == nil, int64(fee)
	}

	return []byte{}, false, 0 // unknown type
}

func (this *BaseHandlers) pid(_ evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	pidNum := this.api.Pid()
	return []byte(hex.EncodeToString(pidNum[:])), true, 0
}

// getByIndex the number of elements in the container
func (this *BaseHandlers) fullLength(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller)
	if length, successful, _ := this.FullLength(path); successful {
		if encoded, err := abi.Encode(uint256.NewInt(length)); err == nil {
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

// getByIndex the number of elements in the container
func (this *BaseHandlers) length(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller)
	if length, successful, _ := this.Length(path); successful {
		if encoded, err := abi.Encode(uint256.NewInt(length)); err == nil {
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

// committedLength the initial length of the container, which would remain the same in the same block.
func (this *BaseHandlers) committedLength(caller evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // BaseHandlers path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	typedv, fees := this.api.WriteCache().(*tempcache.WriteCache).PeekCommitted(path, new(commutative.Path))
	if typedv != nil {
		type measurable interface{ Length() int }
		numKeys := uint64(typedv.(stgtype.Type).Value().(measurable).Length())
		if encoded, err := abi.Encode(uint256.NewInt(numKeys)); err == nil {
			return encoded, true, int64(fees)
		}
	}
	return []byte{}, false, int64(fees)
}

func (this *BaseHandlers) getByKey(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	// Get the key of the element
	key, err := abi.DecodeTo(input, 0, []byte{}, 2, math.MaxInt)
	if err != nil || len(key) == 0 {
		return []byte{}, false, 0
	}

	// Get the type of the container info
	str := hex.EncodeToString(key)

	// Non-commutative bytes container by default
	typeID := this.pathBuilder.GetPathType(caller) // Get the type of the container

	var typedV any
	switch typeID {
	case noncommutative.BYTES: // Commutative container
		typedV = new(noncommutative.Bytes)
	case commutative.UINT256: // Commutative container
		typedV = new(commutative.U256)
	case noncommutative.INT64:
		typedV = new(noncommutative.Int64)
	}

	v, _, fee := this.GetByKey(path+str, typedV)
	if v != nil {
		// special decoder for byte array
		fun := func(v any) ([]byte, bool, error) {
			b, ok := v.([]byte)
			return b, ok, nil
		}

		if encoded, err := abi.Encode(v, fun); err == nil {
			return encoded, true, int64(fee)
		}
	}
	return []byte{}, false, int64(fee)
}

func (this *BaseHandlers) getByIndex(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	index, err := abi.DecodeTo(input, 0, uint64(0), 1, 32)
	if err != nil {
		return []byte{}, false, 0
	}
	return this.GetByIndex(path, index) // Get the value by its key.
}

// Push a new element into the container. If the key does not exist, it will be created and the value will be set.
func (this *BaseHandlers) setByKey(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}
	fee := int64(0)

	// Decode the input value
	key, valueBytes, err := abi.Parse2(input,
		[]byte{}, 2, math.MaxInt,
		[]byte{}, 2, math.MaxInt,
	)

	if err != nil {
		return []byte{}, false, fee
	}

	// Get the type of the container info
	typeID := this.pathBuilder.GetPathType(caller) // Get the type of the container

	// Other types
	var val any
	switch typeID {
	case commutative.UINT256: // Commutative container
		// Decode the input delta value, could be negative or positive.
		var v *big.Int
		if v, err = abi.DecodeInt256(valueBytes); err != nil {
			return []byte{}, false, fee
		}

		// Get the bytes from the delta bytes and Create a new delta value for the element
		delta := new(uint256.Int).SetBytes(new(big.Int).Abs(v).Bytes())
		val = commutative.NewU256Delta(delta, v.Sign() >= 0)

	case noncommutative.INT64:
		bytes := make([]byte, 8)
		copy(bytes, valueBytes) // valueBytes could be less than 8 bytes, so we need to copy it to the bytes array
		val = noncommutative.NewInt64(int64(binary.LittleEndian.Uint64(bytes)))

	case noncommutative.BYTES:
		val = noncommutative.NewBytes(valueBytes) // Non-commutative container by default
	}

	successful, fee := this.SetByKey(path+hex.EncodeToString(key), val)
	return []byte{}, successful, fee
}

func (this *BaseHandlers) delByKey(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	if key, err := abi.DecodeTo(input, 0, []byte{}, 2, math.MaxInt); err == nil {
		if successful, fee := this.SetByKey(path+hex.EncodeToString(key), nil); successful {
			return []byte{}, true, fee
		}
	}
	return []byte{}, false, 0
}

func (this *BaseHandlers) keyToInd(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // BaseHandlers path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	if key, err := abi.DecodeTo(input, 0, []byte{}, 2, math.MaxInt); err == nil {
		index, _ := this.IndexOf(path, hex.EncodeToString(key))
		if encoded, err := abi.Encode(index); index != math.MaxUint64 && err == nil { // Encode the result
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

func (this *BaseHandlers) indToKey(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // BaseHandlers path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	if index, err := abi.DecodeTo(input, 0, uint64(0), 1, 32); err == nil {
		key, _ := this.KeyAt(path, index)
		v, _ := hex.DecodeString(key)
		return v, true, 0
	}
	return []byte{}, false, 0
}

// Get the last element in the container and remove it from the container.
// The size will remain the same, but the last element will be nil.
func (this *BaseHandlers) delLast(caller evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // BaseHandlers path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	length, successful, fee := this.Length(path)
	if !successful || length == 0 {
		return []byte{}, successful, fee
	}

	// Get the last element in the container first before
	// deleting it from the container.
	values, successful, _ := this.GetByIndex(path, length-1)
	if len(values) == 0 || !successful {
		return values, false, 0 // Failed to get the last element
	}

	// Delete the last element in the container.
	successful, fee = this.SetByIndex(path, length-1, nil)
	return values, successful, fee
}

// Delete all elements in the container.
func (this *BaseHandlers) clear(caller evmcommon.Address, _ []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	tx := this.api.GetEU().(interface{ ID() uint64 }).ID()
	if _, _, err := this.api.WriteCache().(*tempcache.WriteCache).EraseAll(tx, path, nil); err != nil {
		return []byte{}, false, 0
	}
	return []byte{}, true, 0
}

// Set all elements in the container to their default value.
func (this *BaseHandlers) resetByKey(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	// Get the key of the element
	key, err := abi.DecodeTo(input, 0, []byte{}, 2, math.MaxInt)
	if err != nil || len(key) == 0 {
		return []byte{}, false, 0
	}

	// Get the type of the container info
	return this.ResetByKey(path, hex.EncodeToString(key)) // Reset the element by its key.
}

func (this *BaseHandlers) resetByInd(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // BaseHandlers path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	var key string
	if index, err := abi.DecodeTo(input, 0, uint64(0), 1, 32); err == nil {
		if key, _ = this.KeyAt(path, index); len(key) == 0 {
			return []byte{}, false, 0 // Key not found
		}
	}
	return this.ResetByKey(path, key)
}
