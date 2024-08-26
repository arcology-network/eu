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
	"bytes"
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/holiman/uint256"
)

func TestEncoder(t *testing.T) {
	var err error

	buffer, _ := Encode(uint64(99))
	// fmt.Println(buffer)
	if buffer[31] != 99 {
		t.Error("Error: Should be 99!")
	}

	buffer, _ = Encode(uint64(256))
	// fmt.Println(buffer)
	if buffer[30] != 1 || buffer[31] != 0 || err != nil {
		t.Error("Error: Wrong value!")
	}

	data := [31]byte{12, 13, 14, 1, 5, 16, 17, 18}
	buffer, err = Encode(data[:])
	// fmt.Println(buffer)
	if len(buffer)%32 != 0 || err != nil {
		t.Error("Error: Should be 32-byte!")
	}

	data2 := [66]byte{12, 13, 14, 1, 5, 16, 17, 18, 19, 21}
	buffer, err = Encode(data2[:])
	// fmt.Println(buffer)

	if len(buffer)%32 != 0 || err != nil {
		t.Error("Error: Should be 32-byte!")
	}

	u256 := uint256.NewInt(999)
	buffer, err = Encode(u256)
	if len(buffer)%32 != 0 || err != nil {
		t.Error("Error: Should be 32-byte!")
	}

	if d, _ := Decode(buffer, 0, new(uint256.Int), 1, math.MaxInt); !u256.Eq(d.(*uint256.Int)) {
		t.Error("Error: Should be equal!")
	}

	bs, err := Encode([]byte{1, 2, 3})
	buffer = append(buffer, bs...)
	// fmt.Println(buffer)

	// 0 for fixed size tyes.
	v0, _ := Decode(buffer, 0, new(uint256.Int), 1, math.MaxInt)
	fmt.Println(v0)

	if v0.(*uint256.Int).Cmp(u256) != 0 {
		t.Error("Error: Should be equal!")
	}

	// 1 for dynamic size types.
	outBytes, _ := Decode(buffer, 1, []byte{}, 1, math.MaxInt)
	if !bytes.Equal(outBytes.([]byte), bs) {
		t.Error("Error: Should be equal!")
	}
}

func TestDecoder(t *testing.T) {
	raw := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 100, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 160, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 32, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 96, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 99, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 75, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 65, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 25, 234, 68, 190, 137, 238, 206, 15, 212, 236, 116, 130, 4, 159, 71, 42, 17, 175, 25, 56, 75, 255, 179, 138, 136, 231, 123, 59, 29, 213, 76, 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 25, 234, 68, 190, 137, 238, 206, 15, 212, 236, 116, 130, 4, 159, 71, 42, 17, 175, 25, 56, 75, 255, 179, 138, 136, 231, 123, 59, 29, 213, 76, 25}
	// buffer, _ := decorder.At(decorder.raw, 0, uint32(0))
	Fields := codec.Bytes32s{}.Decode(raw).(codec.Bytes32s)
	for i := 0; i < len(Fields); i++ {
		fmt.Println(Fields[i])
	}

	buffer, _ := Decode(raw, 1, []byte{}, 2, math.MaxInt)
	fmt.Println(buffer)

	subbytes := buffer.([]byte)
	idx, _ := slice.FindFirstIf(subbytes, func(_ int, v byte) bool { return v != 65 })
	if len(buffer.([]byte)) != 75 || idx != -1 {
		t.Error("Error; The array should be 75 byte long!")
	}

	buffer, _ = Decode(raw, 0, uint32(0), 1, math.MaxInt) //need to indicate here !!
	if buffer.(uint32) != 100 {
		t.Error("Error: Should be 100!")
	}
	fmt.Println(reflect.TypeOf(uint256.NewInt(0)).Kind())

	raw = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 64, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 25, 234, 68, 190, 137, 238, 206, 15, 212, 236, 116, 130, 4, 159, 71, 42, 17, 175, 25, 56, 75, 255, 179, 138, 136, 231, 123, 59, 29, 213, 76, 25}
	buffer, _ = Decode(raw, 1, []byte{}, 1, 32)
	fmt.Println()
	fmt.Println(buffer)

	// field 0
	raw = []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 96, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 160, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 32, 25, 234, 68, 190, 137, 238, 206, 15, 212, 236, 116, 130, 4, 159, 71, 42, 17, 175, 25, 56, 75, 255, 179, 138, 136, 231, 123, 59, 29, 213, 76, 25, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 5, 170, 170, 170, 170, 170, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	buffer, _ = Decode(raw, 0, []byte{}, 2, 32)
	fmt.Println()
	fmt.Println(buffer)

	if !bytes.Equal(buffer.([]byte), []byte{25, 234, 68, 190, 137, 238, 206, 15, 212, 236, 116, 130, 4, 159, 71, 42, 17, 175, 25, 56, 75, 255, 179, 138, 136, 231, 123, 59, 29, 213, 76, 25}) {
		t.Error("Error: Should be equal!")
	}

	// field 1
	buffer, _ = Decode(raw, 1, []byte{}, 1, 32)
	if len(buffer.([]byte)) != 32 || int(buffer.([]byte)[31]) != 1 {
		t.Error("Error: Wrong length")
	}

	// field 2
	buffer, _ = Decode(raw, 2, []byte{}, 2, 32)
	if len(buffer.([]byte)) != 5 {
		t.Error("Error: Wrong length")
	}

	buffer32, _ := Decode(raw, 0, [32]byte{}, 2, math.MaxInt)
	if buffer32.([32]byte)[31] != 96 {
		t.Error("Error: Wrong [32]byte length")
	}
}
