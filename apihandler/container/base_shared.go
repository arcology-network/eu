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
	"strings"

	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	noncommutative "github.com/arcology-network/common-lib/crdt/noncommutative"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	abi "github.com/arcology-network/eu/abi"
	eucommon "github.com/arcology-network/eu/common"
	cache "github.com/arcology-network/state-engine/state/cache"
	"github.com/holiman/uint256"
)

// Get the number of elements in the container, EXCLUDING the nil elements.
func (this *BaseHandlers) NonNilLength(path string) (uint64, bool, int64) {
	if path, _, dataSize := this.api.StateCache().(*cache.StateCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
		return path.(*softdeltaset.DeltaSet[string]).NonNilCount(), true, int64(dataSize)
	}
	return 0, false, int64(eucommon.DATA_MIN_READ_SIZE)
}

// Get the number of elements in the container, INCLUDING the nil elements.
func (this *BaseHandlers) FullLength(path string) (uint64, bool, int64) {
	if path, _, dataSize := this.api.StateCache().(*cache.StateCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
		return path.(*softdeltaset.DeltaSet[string]).Length(), true, int64(dataSize)
	}
	return 0, false, int64(eucommon.DATA_MIN_READ_SIZE) // Return 0 if the path does not exist, but return a data size of 32 bytes to avoid errors in the client code.
}

// Export all the elements in the container to a two-dimensional slice.
// This function will read all the elements in the container.
func (this *BaseHandlers) ReadAll(path string) ([][]byte, []bool, int64) {
	length, _, dataSize := this.NonNilLength(path)
	entries := make([][]byte, length)
	flags := make([]bool, length)
	dataSizes := make([]int64, length)

	slice.NewDo(int(length), func(i int) []byte {
		entries[i], flags[i], dataSizes[i] = this.GetByIndex(path, uint64(i))
		return []byte{}
	})

	return entries, flags, int64(dataSize) + slice.Sum[int64, int64](dataSizes)
}

// Get the element by its key
func (this *BaseHandlers) GetByKey(path string, T any) (any, bool, int64) {
	value, _, dataSize := this.api.StateCache().(*cache.StateCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, T)
	return value, true, int64(dataSize)
}

// Get the index of the element by its key
func (this *BaseHandlers) GetByIndex(path string, idx uint64) ([]byte, bool, int64) {
	keyidx := strings.LastIndex(path, "/")
	elemTypeID := this.pathBuilder.PathElemTypeIDs(path[:keyidx] + "/") // Get the type of the container

	var typedV any
	switch elemTypeID {
	case commutative.UINT256: // Commutative container
		typedV = new(commutative.U256)
	case noncommutative.BYTES: // Noncommutative container
		typedV = new(noncommutative.Bytes)
	case noncommutative.INT64:
		typedV = new(noncommutative.Int64)
	}

	value, dataSize, err := this.api.StateCache().(*cache.StateCache).ReadAt(
		this.api.GetEU().(interface{ ID() uint64 }).ID(), path, idx, typedV)

	if err == nil && value != nil {
		encoder := func(v any) ([]byte, bool, error) {
			b, ok := v.([]byte)
			return b, ok, nil
		}

		if encoded, err := abi.Encode(value, encoder); err == nil {
			return encoded, true, int64(dataSize)
		}
	}
	return []byte{}, false, int64(dataSize)
}

// Set the element by its index
func (this *BaseHandlers) SetByIndex(path string, idx uint64, value any) (bool, int64) {
	dataSize, err := this.api.StateCache().(*cache.StateCache).WriteAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, idx, value)
	return err == nil, dataSize
}

// Set the element by its key
func (this *BaseHandlers) SetByKey(path string, value any) (bool, int64) {
	dataSize, err := this.api.StateCache().(*cache.StateCache).Write(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, value)
	return err == nil, dataSize
}

// Get the index of a key
func (this *BaseHandlers) KeyAt(path string, index uint64) (string, uint64) {
	key, dataSize := this.api.StateCache().(*cache.StateCache).KeyAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, index, new(noncommutative.Bytes))
	return key, dataSize
}

// Get the index of a key
func (this *BaseHandlers) IndexOf(path string, key string) (uint64, int64) {
	index, dataSize := this.api.StateCache().(*cache.StateCache).IndexOf(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, key, new(noncommutative.Bytes))
	return index, int64(dataSize)
}

func (this *BaseHandlers) ResetByKey(path string, key string) ([]byte, bool, int64) {
	var typedV any
	elemTypeID := this.pathBuilder.PathElemTypeIDs(path) // Get the type of the container
	readDataSize := uint64(0)
	var v any
	switch elemTypeID {
	case commutative.UINT256: // Commutative container
		_, v, readDataSize = this.api.StateCache().(*cache.StateCache).Peek(path+key, new(commutative.U256))
		if v == nil {
			return []byte{}, false, int64(readDataSize) // Not found
		}

		delta, _, _ := v.(*commutative.U256).Get()
		absDelta := delta.(uint256.Int)

		rawminv, rawmaxv := v.(*commutative.U256).Limits()
		minv := rawminv.(uint256.Int)
		maxv := rawmaxv.(uint256.Int)

		_, sign := v.(*commutative.U256).Delta()
		typedV = commutative.NewBoundedU256Delta(&minv, &maxv, &absDelta, !sign) // Set the delta to the opposite of the current delta to set it to zero.

	case noncommutative.BYTES:
		typedV = noncommutative.NewBytes([]byte{}) // Non-commutative container by default

	case noncommutative.INT64:
		typedV = noncommutative.NewInt64(0) // Non-commutative container by default
	}

	// Bytes by default
	writeDataSize, err := this.api.StateCache().(*cache.StateCache).Write(
		this.api.GetEU().(interface{ ID() uint64 }).ID(), path+key, typedV)

	return []byte{}, err == nil, int64(readDataSize) + writeDataSize
}

// PopAt removes the element at the given index and returns it.
func (this *BaseHandlers) ExtractAt(path string, idx uint64) ([]byte, bool, int64) {
	funCall, getSuccessful, readDataSize := this.GetByIndex(path, idx) // Get the function call data and the fee.
	setSuccessful, writeDataSize := this.SetByIndex(path, idx, nil)
	return funCall, getSuccessful && setSuccessful, readDataSize + writeDataSize
}
