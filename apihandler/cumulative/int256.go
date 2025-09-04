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

// import (
// 	"encoding/hex"

// 	"github.com/arcology-network/common-lib/codec"
// 	"github.com/arcology-network/storage-committer"
// 	"github.com/holiman/uint256"

// 	"github.com/arcology-network/storage-committer/type/commutative"
// 	evmcommon "github.com/ethereum/go-ethereum/common"
// 	abi "github.com/arcology-network/eu/abi"
// 	"github.com/arcology-network/eu/common"
// 	eucommon "github.com/arcology-network/eu/common"

//
// )

// // APIs under the concurrency namespace
// type Int256CumulativeHandler struct {
// 	api       intf.EthApiRouter,
// 	connector *BuiltinPathMaker
// }

// func NewInt256CumulativeHandler(api intf.EthApiRouter) *Int256CumulativeHandler {
// 	return &Int256CumulativeHandler{
// 		api:       api,
// 		connector: NewBuiltinPathMaker("/container", api, api.WriteCache()),
// 	}
// }

// func (this *Int256CumulativeHandler) Address() [20]byte {
// 	return common.CUMULATIVE_I256_HANDLER
// }

// func (this *Int256CumulativeHandler) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64) ([]byte, bool, int64) {
// 	signature := [4]byte{}
// 	copy(signature[:], input)

// 	switch signature {
// 	case [4]byte{0x90, 0x54, 0xce, 0x5f}:
// 		return this.new(caller, input[4:])

// 	case [4]byte{0x6d, 0x4c, 0xe6, 0x3c}:
// 		return this.get(caller, input[4:])

// 	case [4]byte{0xa4, 0xc6, 0xa7, 0x68}:
// 		return this.add(caller, input[4:])

// 	case [4]byte{0xc8, 0xda, 0xaa, 0xab}:
// 		return this.sub(caller, input[4:])
// 	}
// 	return this.Unknow(caller, input)
// }

// func (this *Int256CumulativeHandler) Unknow(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	this.api.AddLog("Unhandled function call in cumulative handler router", hex.EncodeToString(input))
// 	return []byte{}, false, 0
// }

// func (this *Int256CumulativeHandler) new(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	id := this.api.UUID()
// 	if !this.connector.New(types.Address(codec.Bytes20(caller).Hex()), hex.EncodeToString(id)) { // A new container
// 		return []byte{}, false, 0
// 	}

// 	path := this.connector.Key(types.Address(codec.Bytes20(caller).Hex()), hex.EncodeToString(id))
// 	key := path + string(this.api.ElementUID()) // Element ID

// 	// val, valErr := abi.Decode(input, 0, &uint256.Int{}, 1, 32)
// 	min, minErr := abi.Decode(input, 0, &uint256.Int{}, 1, 32)
// 	max, maxErr := abi.Decode(input, 1, &uint256.Int{}, 1, 32)
// 	if minErr != nil || maxErr != nil {
// 		return []byte{}, false, 0
// 	}

// 	newU256 := commutative.NewU256(min.(*uint256.Int), max.(*uint256.Int))
// 	if _, err := this.api.WriteCache().(*cache.WriteCache).Write(uint32(this.api.StdMessage().(*execution.StandardMessage).ID), key, newU256, true); err != nil {
// 		return []byte{}, false, 0
// 	}
// 	return id, true, 0
// }

// func (this *Int256CumulativeHandler) get(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path :=  this.connector.Key(caller)// Build container path
// 	if len(path) == 0 || err != nil {
// 		return []byte{}, false, 0
// 	}

// 	if value, _, err := this.api.WriteCache().(*cache.WriteCache).ReadAt(uint32(this.api.StdMessage().(*execution.StandardMessage).ID), path, 0); value == nil || err != nil {
// 		return []byte{}, false, 0
// 	} else {

// 		updated := value.(*uint256.Int)
// 		if encoded, err := abi.Encode(updated); err == nil { // Encode the result
// 			return encoded, true, 0
// 		}
// 	}
// 	return []byte{}, false, 0
// }

// func (this *Int256CumulativeHandler) add(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path :=  this.connector.Key(caller)// Build container path
// 	if len(path) == 0 || err != nil {
// 		return []byte{}, false, 0
// 	}

// 	delta, err := abi.Decode(input, 1, &uint256.Int{}, 1, 32)
// 	if err != nil {
// 		return []byte{}, false, 0
// 	}

// 	value := commutative.NewU256Delta(delta.(*uint256.Int), true)
// 	_, err = this.api.WriteCache().(*cache.WriteCache).WriteAt(uint32(this.api.StdMessage().(*execution.StandardMessage).ID), path, 0, value, true)
// 	return []byte{}, err == nil, 0
// }

// func (this *Int256CumulativeHandler) sub(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path :=  this.connector.Key(caller)// Build container path
// 	if len(path) == 0 || err != nil {
// 		return []byte{}, false, 0
// 	}

// 	delta, err := abi.Decode(input, 1, &uint256.Int{}, 1, 32)
// 	if err != nil {
// 		return []byte{}, false, 0
// 	}

// 	value := commutative.NewU256Delta(delta.(*uint256.Int), false)
// 	_, err = this.api.WriteCache().(*cache.WriteCache).WriteAt(uint32(this.api.StdMessage().(*execution.StandardMessage).ID), path, 0, value, true)
// 	return []byte{}, err == nil, 0
// }

// func (this *Int256CumulativeHandler) set(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path :=  this.connector.Key(caller)// Build container path
// 	if len(path) == 0 || err != nil {
// 		return []byte{}, false, 0
// 	}

// 	delta, err := abi.DecodeTo(input, 1, &uint256.Int{}, 1, 32)
// 	if err != nil {
// 		return []byte{}, false, 0
// 	}

// 	sign, err := abi.DecodeTo(input, 1, bool(true), 1, 32)
// 	if err != nil {
// 		return []byte{}, false, 0
// 	}

// 	value := commutative.NewU256Delta(delta, sign)
// 	_, err = this.api.WriteCache().(*cache.WriteCache).WriteAt(uint32(this.api.StdMessage().(*execution.StandardMessage).ID), path, 0, value, true)
// 	return []byte{}, err == nil, 0
// }

// // Build the container path
// func (this *Int256CumulativeHandler) buildPath(caller evmcommon.Address, input []byte) (string, error) {
// 	id, err := abi.Decode(input, 0, []byte{}, 2, 32) // max 32 bytes
// 	if err != nil {
// 		return "", nil
// 	} // container ID
// 	return this.connector.Key(types.Address(codec.Bytes20(caller).Hex()), hex.EncodeToString(id.([]byte))), nil // unique ID
// }
