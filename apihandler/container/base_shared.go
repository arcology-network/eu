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
	"math"
	"strings"

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	abi "github.com/arcology-network/eu/abi"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/holiman/uint256"
)

// Get the number of elements in the container, EXCLUDING the nil elements.
func (this *BaseHandlers) Length(path string) (uint64, bool, int64) {
	if len(path) == 0 {
		return 0, false, 0
	}

	if path, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
		return path.(*deltaset.DeltaSet[string]).NonNilCount(), true, 0
	}
	return 0, false, 0
}

// Get the number of elements in the container, INCLUDING the nil elements.
func (this *BaseHandlers) FullLength(path string) (uint64, bool, int64) {
	if len(path) == 0 {
		return 0, false, 0
	}

	if path, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
		return path.(*deltaset.DeltaSet[string]).Length(), true, 0
	}
	return 0, false, 0
}

// Export all the elements in the container to a two-dimensional slice.
// This function will read all the elements in the container.
func (this *BaseHandlers) ReadAll(path string) ([][]byte, []bool, []int64) {
	length, _, _ := this.Length(path)
	entries := make([][]byte, length)
	flags := make([]bool, length)
	fees := make([]int64, length)

	slice.NewDo(int(length), func(i int) []byte {
		entries[i], flags[i], fees[i] = this.GetByIndex(path, uint64(i))
		return []byte{}
	})
	return entries, flags, fees
}

// Get the element by its key
func (this *BaseHandlers) GetByKey(path string, T any) (any, bool, int64) {
	value, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, T)
	return value, true, 0
}

// Get the index of the element by its key
func (this *BaseHandlers) GetByIndex(path string, idx uint64) ([]byte, bool, int64) {
	keyidx := strings.LastIndex(path, "/")
	typeID := this.pathBuilder.PathTypeID(path[:keyidx] + "/") // Get the type of the container

	var typedV any
	switch typeID {
	case commutative.UINT256: // Commutative container
		typedV = new(commutative.U256)
	case noncommutative.BYTES: // Noncommutative container
		typedV = new(noncommutative.Bytes)
	case noncommutative.INT64:
		typedV = new(noncommutative.Int64)
	}

	if value, _, err := this.api.WriteCache().(*tempcache.WriteCache).ReadAt(
		this.api.GetEU().(interface{ ID() uint64 }).ID(), path, idx, typedV); err == nil && value != nil {

		encoder := func(v any) ([]byte, bool, error) {
			b, ok := v.([]byte)
			return b, ok, nil
		}

		if encoded, err := abi.Encode(value, encoder); err == nil {
			return encoded, true, int64(0)
		}
		// return value.([]byte), true, 0
	}
	return []byte{}, false, 0
}

// Set the element by its index
func (this *BaseHandlers) SetByIndex(path string, idx uint64, value any) (bool, int64) {
	if len(path) == 0 {
		return false, 0
	}

	if _, err := this.api.WriteCache().(*tempcache.WriteCache).WriteAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, idx, value); err == nil {
		return true, 0
	}
	return false, 0
}

// Set the element by its key
func (this *BaseHandlers) SetByKey(path string, value any) (bool, int64) {
	if len(path) > 0 {
		if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, value); err == nil {
			return true, 0
		}
	}
	return false, 0
}

// Get the index of a key
func (this *BaseHandlers) KeyAt(path string, index uint64) (string, int64) {
	if len(path) > 0 {
		key, _ := this.api.WriteCache().(*tempcache.WriteCache).KeyAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, index, new(noncommutative.Bytes))
		return key, 0
	}
	return "", 0
}

// Get the index of a key
func (this *BaseHandlers) IndexOf(path string, key string) (uint64, int64) {
	if len(path) > 0 {
		index, _ := this.api.WriteCache().(*tempcache.WriteCache).IndexOf(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, key, new(noncommutative.Bytes))
		return index, 0
	}
	return math.MaxUint64, 0
}

func (this *BaseHandlers) ResetByKey(path string, key string) ([]byte, bool, int64) {
	var typedV any
	typeID := this.pathBuilder.PathTypeID(path) // Get the type of the container
	switch typeID {
	case commutative.UINT256: // Commutative container
		v, _ := this.api.WriteCache().(*tempcache.WriteCache).PeekRaw(path+key, new(commutative.U256))
		if v == nil {
			return []byte{}, false, int64(0)
		}

		absDelta := v.(*commutative.U256).Delta().(uint256.Int)
		typedV = commutative.NewU256Delta(&absDelta, !v.(*commutative.U256).DeltaSign()) // Set the delta to the opposite of the current delta to set it to zero.
	case noncommutative.BYTES:
		typedV = noncommutative.NewBytes([]byte{}) // Non-commutative container by default

	case noncommutative.INT64:
		typedV = noncommutative.NewInt64(0) // Non-commutative container by default
	}

	// Bytes by default
	fee, err := this.api.WriteCache().(*tempcache.WriteCache).Write(
		this.api.GetEU().(interface{ ID() uint64 }).ID(), path+key, typedV)

	return []byte{}, err == nil, int64(fee)
}
