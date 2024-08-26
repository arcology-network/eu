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
	"encoding/hex"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/types"
	"github.com/holiman/uint256"

	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	cache "github.com/arcology-network/common-lib/types/storage/writecache"
	abi "github.com/arcology-network/eu/abi"
	"github.com/arcology-network/eu/common"
	evmcommon "github.com/ethereum/go-ethereum/common"

	// intf "github.com/arcology-network/common-lib/types/execution"
	path "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
)

// U256CumulativeHandlers handles the U256Cumulative APIs that can be called by concurrent API called.
type U256CumHandler struct {
	api       intf.EthApiRouter
	connector *path.PathBuilder
	key       string
}

func NewU256CumulativeHandler(api intf.EthApiRouter) *U256CumHandler {
	k := [20]byte{}
	return &U256CumHandler{
		api:       api,
		connector: path.NewPathBuilder("/storage/container", api),
		key:       hex.EncodeToString(k[:]),
	}
}

func (this *U256CumHandler) Address() [20]byte {
	return common.CUMULATIVE_U256_HANDLER
}

func (this *U256CumHandler) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64) ([]byte, bool, int64) {
	signature := [4]byte{}
	copy(signature[:], input)

	switch signature {
	case [4]byte{0x1c, 0x64, 0x49, 0x9c}:
		return this.new(caller, input[4:])

	case [4]byte{0x59, 0xe0, 0x2d, 0xd7}: // 59 e0 2d d7
		return this.peek(caller, input[4:])

	case [4]byte{0x6d, 0x4c, 0xe6, 0x3c}:
		return this.get(caller, input[4:])

	case [4]byte{0xf8, 0x89, 0x79, 0x45}: // f8 89 79 45
		return this.min(caller, input[4:])

	case [4]byte{0x6a, 0xc5, 0xdb, 0x19}:
		return this.max(caller, input[4:])

	case [4]byte{0x10, 0x03, 0xe2, 0xd2}: // 10 03 e2 d2
		return this.add(caller, input[4:])

	case [4]byte{0x27, 0xee, 0x58, 0xa6}:
		return this.sub(caller, input[4:]) //27 ee 58 a6
	}
	return []byte{}, false, 0
}

func (this *U256CumHandler) new(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	txIndex := this.api.GetEU().(interface{ ID() uint32 }).ID()
	if !this.connector.New(txIndex, types.Address(codec.Bytes20(caller).Hex())) { // A new container
		return []byte{}, false, 0
	}

	min, minErr := abi.Decode(input, 0, &uint256.Int{}, 1, 32)
	max, maxErr := abi.Decode(input, 1, &uint256.Int{}, 1, 32)
	if minErr != nil || maxErr != nil {
		return []byte{}, false, 0
	}

	keyPath := this.connector.Key(caller) + string(this.key) // Element ID
	newU256 := commutative.NewBoundedU256(min.(*uint256.Int), max.(*uint256.Int))
	if _, err := this.api.WriteCache().(*cache.WriteCache).Write(txIndex, keyPath, newU256); err != nil {
		return []byte{}, false, 0
	}
	return []byte{}, true, 0
}

func (this *U256CumHandler) get(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.connector.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	keyPath := path + string(this.key) // Element ID
	if value, _, _ := this.api.WriteCache().(*cache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint32 }).ID(), keyPath, new(commutative.U256)); value == nil {
		return []byte{}, false, 0
	} else {
		updated := value.(uint256.Int)
		if encoded, err := abi.Encode(updated); err == nil { // Encode the result
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

// Peek reads the initial value from the WriteCache. It assumes that the initial value
// is always in the cache by the time it is called.
func (this *U256CumHandler) peek(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.connector.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	keyPath := path + string(this.key) // Element ID
	if value, _ := this.api.WriteCache().(*cache.WriteCache).PeekCommitted(keyPath, new(commutative.U256)); value != nil {
		initv := value.(*commutative.U256).Value().(uint256.Int)
		if encoded, err := abi.Encode((*uint256.Int)(&initv)); err == nil { // Encode the result
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

// Add adds a positive delta to the variable's delta.
func (this *U256CumHandler) add(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	return this.set(caller, input, true)
}

// Add adds a negative delta to the variable's delta.
func (this *U256CumHandler) sub(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	return this.set(caller, input, false)
}

func (this *U256CumHandler) set(caller evmcommon.Address, input []byte, isPositive bool) ([]byte, bool, int64) {
	path := this.connector.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	delta, err := abi.Decode(input, 0, &uint256.Int{}, 1, 32)
	if err != nil {
		return []byte{}, false, 0
	}

	value := commutative.NewU256Delta(delta.(*uint256.Int), isPositive)

	txIndex := this.api.GetEU().(interface{ ID() uint32 }).ID()
	keyPath := path + string(this.key) // Element ID
	_, err = this.api.WriteCache().(*cache.WriteCache).Write(txIndex, keyPath, value)
	return []byte{}, err == nil, 0
}

func (this *U256CumHandler) min(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.connector.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	// Min and Max are read only variable
	txIndex := this.api.GetEU().(interface{ ID() uint32 }).ID()
	keyPath := path + string(this.key) // Element ID
	if value, _ := this.api.WriteCache().(*cache.WriteCache).Find(txIndex, keyPath, new(commutative.U256)); value != nil {
		minv := value.(*commutative.U256).Min().(uint256.Int)
		if encoded, err := abi.Encode((*uint256.Int)(&minv)); err == nil { // Encode the result
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}

func (this *U256CumHandler) max(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.connector.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	txIndex := this.api.GetEU().(interface{ ID() uint32 }).ID()
	keyPath := path + string(this.key) // Element ID
	if value, _ := this.api.WriteCache().(*cache.WriteCache).Find(txIndex, keyPath, new(commutative.U256)); value != nil {
		maxv := value.(*commutative.U256).Max().(uint256.Int)
		if encoded, err := abi.Encode((*uint256.Int)(&maxv)); err == nil { // Encode the result
			return encoded, true, 0
		}
	}
	return []byte{}, false, 0
}
