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

	abi "github.com/arcology-network/eu/abi"
	evmcommon "github.com/ethereum/go-ethereum/common"
)

// getByIndex the element by its index
func (this *BaseHandlers) getByIndex(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	if len(path) == 0 {
		return []byte{}, false, 0
	}

	idx, err := abi.DecodeTo(input, 0, uint64(0), 1, 32)
	if err != nil {
		return []byte{}, false, 0
	}

	values, successful, _ := this.GetByIndex(path, idx)
	if len(values) > 0 && successful {
		return values, true, 0
	}
	return []byte{}, false, 0
}

// Set the element by its index. This function is different from the SetByKey function in that if
// the index is out of range, the function will return false.
func (this *BaseHandlers) setByIndex(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path

	idx, bytes, err := abi.Parse2(input,
		uint64(0), 1, 32,
		[]byte{}, 2, math.MaxInt,
	)

	if err != nil {
		return []byte{}, false, 0
	}

	if successful, fee := this.SetByIndex(path, idx, bytes); successful {
		return []byte{}, true, fee
	}
	return []byte{}, false, 0
}

func (this *BaseHandlers) delByIndex(caller evmcommon.Address, input []byte) ([]byte, bool, int64) {
	path := this.pathBuilder.Key(caller) // Build container path
	idx, err := abi.DecodeTo(input, 0, uint64(0), 1, 32)
	if err == nil {
		if successful, fee := this.SetByIndex(path, idx, nil); successful {
			return []byte{}, true, fee
		}
	}
	return []byte{}, false, 0
}
