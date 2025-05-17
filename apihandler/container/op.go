/*
 *   Copyright (c) 2025 Arcology Network

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
	"bytes"
	"math/big"

	// "github.com/arcology-network/common-lib/exp/deltaset"

	"github.com/arcology-network/common-lib/exp/slice"

	abi "github.com/arcology-network/eu/abi"
	evmcommon "github.com/ethereum/go-ethereum/common"

	"github.com/holiman/uint256"
)

// The function returns the minimum value in the container sorted by numerical order by
// converting the byte array to a big integer and comparing the two values.
func (this *BaseHandlers) min(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller)
	entries, _, _ := this.ReadAll(path)

	lhv, rhv := new(big.Int), new(big.Int)
	idx, v := slice.Extreme(entries, func(lhvBytes, rhvBytes []byte) bool {
		lhv.SetBytes(lhvBytes) // Convert the byte array to a big integer
		rhv.SetBytes(rhvBytes)
		return lhv.Cmp(rhv) < 0
	})

	// This leaves a read access for the minimum value in the container. It will be used for the conflict detection
	if val, _, _ := this.GetByIndex(path, uint64(idx)); !bytes.Equal(v, val) {
		panic("The value is not equal to the value in the container.")
	}

	idxBytes, _ := abi.Encode(uint256.NewInt(uint64(idx)))
	return append(idxBytes, v...), true, 0
}

// The function max returns the maximum value in the container sorted by numerical order by
// converting the byte array to a big integer and comparing the two values.
func (this *BaseHandlers) max(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller)
	entries, _, _ := this.ReadAll(path)

	lhv, rhv := new(big.Int), new(big.Int)
	idx, v := slice.Extreme(entries, func(lhvBytes, rhvBytes []byte) bool {
		lhv.SetBytes(lhvBytes) // Convert the byte array to a big integer
		rhv.SetBytes(rhvBytes)
		return lhv.Cmp(rhv) > 0
	})

	// This leaves a read access for the maxmium value in the container. It will be used for the conflict detection
	if val, _, _ := this.GetByIndex(path, uint64(idx)); !bytes.Equal(v, val) {
		panic("The value is not equal to the value in the container.")
	}

	idxBytes, _ := abi.Encode(uint256.NewInt(uint64(idx)))
	return append(idxBytes, v...), true, 0
}

// func (this *BaseHandlers) minString(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path := this.pathBuilder.Key(caller)
// 	entries, _, _ := this.ReadAll(path)

// 	idx, v := slice.Extreme(entries, func(lhv, rhv []byte) bool {
// 		return string(lhv) < string(rhv)
// 	})

// 	// This leaves a read access for the maxmium string in the container. It will be used for the conflict detection
// 	if val, _, _ := this.GetByIndex(path, uint64(idx)); !bytes.Equal(v, val) {
// 		panic("The value is not equal to the value in the container.")
// 	}

// 	idxBytes, _ := abi.Encode(uint256.NewInt(uint64(idx)))
// 	return append(idxBytes, v...), true, 0
// }

// func (this *BaseHandlers) maxString(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
// 	path := this.pathBuilder.Key(caller)
// 	entries, _, _ := this.ReadAll(path)

// 	idx, v := slice.Extreme(entries, func(lhv, rhv []byte) bool {
// 		return string(lhv) > string(rhv)
// 	})

// 	// This leaves a read access for the maxmium string in the container. It will be used for the conflict detection
// 	if val, _, _ := this.GetByIndex(path, uint64(idx)); !bytes.Equal(v, val) {
// 		panic("The value is not equal to the value in the container.")
// 	}

// 	idxBytes, _ := abi.Encode(uint256.NewInt(uint64(idx)))
// 	return append(idxBytes, v...), true, 0
// }
