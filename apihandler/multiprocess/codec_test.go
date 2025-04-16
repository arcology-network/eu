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

package multiprocessor

import (
	"bytes"
	"fmt"
	"log"
	"testing"
)

func TestEncoder(t *testing.T) {
	returnData := [][]byte{
		[]byte{0x01},
		[]byte{0x02},
		[]byte{0x03},
	}
	flags := []bool{true, false, true}

	encoded, err := EncodeCallReturns(returnData, flags)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Encoded output: 0x%x\n", encoded)

	results, _ := DecodeCallReturns(encoded)
	fmt.Println(results)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	if results[0].Success != true || results[1].Success != false || results[2].Success != true {
		t.Errorf("Expected success flags to be true, false, true but got %v", results)
	}

	if !bytes.Equal(results[0].ReturnData, []byte{1}) || !bytes.Equal(results[1].ReturnData, []byte{2}) || !bytes.Equal(results[2].ReturnData, []byte{3}) {
		t.Errorf("Expected return data to be 0x01, 0x02, 0x03 but got %v", results)
	}
}
