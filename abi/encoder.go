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

package abi

import (
	"encoding/binary"
	"errors"

	"github.com/holiman/uint256"
)

func Encode(typed interface{}) ([]byte, error) {
	buffer := [32]byte{}

	switch typed.(type) {
	case bool:
		if typed.(bool) {
			buffer[31] = 1
		}
		return buffer[:], nil

	case uint8:
		buffer[31] = typed.(uint8)
		return buffer[:], nil

	case uint16:
		binary.BigEndian.PutUint16(buffer[32-2:], typed.(uint16))
		return buffer[:], nil

	case uint32:
		binary.BigEndian.PutUint32(buffer[32-4:], typed.(uint32))
		return buffer[:], nil

	case uint64:
		binary.BigEndian.PutUint64(buffer[32-8:], typed.(uint64))
		return buffer[:], nil

	case *uint256.Int:
		bytes := typed.(*uint256.Int).Bytes32()
		return bytes[:], nil

	case uint256.Int:
		v := typed.(uint256.Int)
		bytes := (&v).Bytes32()
		return bytes[:], nil

	case string:
		return []byte(typed.(string)), nil

	case [20]uint8:
		bytes := typed.([20]uint8)
		return bytes[:], nil

	case [32]uint8:
		bytes := typed.([32]uint8)
		return bytes[:], nil

	case []uint8:
		binary.BigEndian.PutUint32(buffer[32-4:], uint32(len(typed.([]byte))))
		if len(typed.([]byte))%32 == 0 {
			return append(buffer[:], typed.([]byte)...), nil
		}

		body := make([]byte, len(typed.([]byte))/32*32+32)
		copy(body, typed.([]byte))
		return append(buffer[:], body...), nil
	}
	return []byte{}, errors.New("Error: Unsupported data type")
}

func AddOffset(sections [][]byte) []byte {
	encoded := []byte{}

	// sumLength := 0
	// for i := 0; i < len(sections); i++ {
	// 	sumLength
	// }

	// offset := [32]byte{}
	// offset[len(offset)-1] = uint8(len(offset))
	// encoded = append(offset[:], encoded...)
	return encoded
}
