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

package shared

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestEuresultEncodingWithDefer(t *testing.T) {
	// in := &DeferredCall{
	// 	DeferID:         "7777999",
	// 	ContractAddress: "45678abc",
	// 	Signature:       "xxxx%123",
	// }

	eu := &EuResult{
		H:  "0x1234567",
		ID: 99,
		//Transitions:  []byte{byte(1), byte(2)},
		TransitTypes: []byte{8, 7},
		// DC:           in,
		Status:  12,
		GasUsed: 34,
	}

	// if eu.DC == nil {
	// 	fmt.Println()
	// }

	buffer := eu.Encode()
	out := (&EuResult{}).Decode(buffer)
	if !reflect.DeepEqual(*eu, *out) {
		t.Error("Error")
	}
}

// func TestDeferEncoding(t *testing.T) {
// 	// in := &DeferredCall{
// 	// 	DeferID:         "7777999",
// 	// 	ContractAddress: "45678abc",
// 	// 	Signature:       "xxxx%123",
// 	// }

// 	buffer := in.Encode()
// 	out := (&DeferredCall{}).Decode(buffer)

// 	if !reflect.DeepEqual(in, out) {
// 		t.Error("Error")
// 	}
// }

func TestEuResultEncodingWithDefer(t *testing.T) {
	// dc := &DeferredCall{
	// 	DeferID:         "7777",
	// 	ContractAddress: "45678",
	// 	Signature:       "xxxx",
	// }

	euresult := EuResult{
		H:  "1234",
		ID: uint32(99),
		//Transitions:  []byte{byte(2), byte(8)},
		TransitTypes: []byte{1, 2},
		// DC:           dc,
		Status:  0,
		GasUsed: 199,
	}

	buffer := euresult.Encode()
	out := (&EuResult{}).Decode(buffer)

	if !reflect.DeepEqual(euresult, *out) {
		t.Error("Error")
	}
}

func TestEuResultsEncoding(t *testing.T) {
	// dc := &DeferredCall{
	// 	DeferID:         "7777",
	// 	ContractAddress: "45678",
	// 	Signature:       "xxxx",
	// }

	euresults := make([]*EuResult, 10)
	for i := 0; i < len(euresults); i++ {
		euresults[i] = &EuResult{
			H:  "0x1234567",
			ID: uint32(99),
			//Transitions:  []byte{byte(9), byte(11)},
			TransitTypes: []byte{1, 2},
			// DC:           dc,
			Status:  11,
			GasUsed: 99,
		}
	}

	t0 := time.Now()
	buffer, _ := Euresults(euresults).GobEncode()
	fmt.Println("EuResults GobEncode():", time.Now().Sub(t0))

	out := new(Euresults)
	out.GobDecode(buffer)
	fmt.Println("EuResults GobDecode():", time.Now().Sub(t0))

	for i := 0; i < len(euresults); i++ {
		if !reflect.DeepEqual(euresults[i], (*out)[i]) {
			t.Error("Error")
		}
	}
}

func BenchmarkEuResultsEncoding(b *testing.B) {
	// dc := &DeferredCall{
	// 	DeferID:         "7777",
	// 	ContractAddress: "45678",
	// 	Signature:       "xxxx",
	// }

	euresults := make([]*EuResult, 1000000)
	for i := 0; i < len(euresults); i++ {
		euresults[i] = &EuResult{
			H:  "0x1234567",
			ID: uint32(99),
			//Transitions: []byte{byte(90), byte(110)},
			// DC:          dc,
			Status:  11,
			GasUsed: 99,
		}
	}

	t0 := time.Now()
	buffer, _ := Euresults(euresults).GobEncode()
	fmt.Println("EuResults GobEncode():", time.Now().Sub(t0))

	out := new(Euresults)
	out.GobDecode(buffer)
	fmt.Println("EuResults GobDecode():", time.Now().Sub(t0))
}
