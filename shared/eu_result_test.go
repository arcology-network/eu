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
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	"github.com/arcology-network/common-lib/crdt/statecell"
	"github.com/holiman/uint256"
)

func TestEuresultEncodingWithDefer(t *testing.T) {

	alice := RandomAccount()

	u64 := commutative.NewBoundedUint64(0, 100)
	in0 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)

	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256, nil)

	eu := &EuResult{
		H:            "0x1234567",
		ID:           99,
		TransitTypes: []byte{8, 7},
		Trans:        []*statecell.StateCell{in0, in1},
		Status:       12,
		GasUsed:      34,
	}

	buffer := eu.Encode()
	out := (&EuResult{}).Decode(buffer)
	aj, _ := json.Marshal(eu)
	bj, _ := json.Marshal(out)
	if !bytes.Equal(aj, bj) {
		t.Error("Error")
	}
}

func TestEuResultsEncoding(t *testing.T) {
	euresults := make([]*EuResult, 10)
	for i := 0; i < len(euresults); i++ {
		alice := RandomAccount()
		val := 1001 + i
		u64 := commutative.NewBoundedUint64(0, uint64(val))
		in0 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)

		euresults[i] = &EuResult{
			H:            "0x1234567",
			ID:           uint64(99),
			TransitTypes: []byte{1, 2},
			Trans:        []*statecell.StateCell{in0},
			Status:       11,
			GasUsed:      99,
		}
	}

	t0 := time.Now()
	buffer, _ := Euresults(euresults).GobEncode()
	fmt.Println("EuResults GobEncode():", time.Now().Sub(t0))

	out := new(Euresults)
	out.GobDecode(buffer)
	fmt.Println("EuResults GobDecode():", time.Now().Sub(t0))

	for i := range euresults {
		aj, _ := json.Marshal(euresults[i])
		bj, _ := json.Marshal((*out)[i])
		if !bytes.Equal(aj, bj) {
			t.Error("Error")
		}
	}
}

func BenchmarkEuResultsEncoding(b *testing.B) {

	euresults := make([]*EuResult, 1000000)
	for i := 0; i < len(euresults); i++ {
		alice := RandomAccount()
		val := 1001 + i
		u64 := commutative.NewBoundedUint64(0, uint64(val))
		in0 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)
		euresults[i] = &EuResult{
			H:            "0x1234567",
			ID:           uint64(99),
			TransitTypes: []byte{1, 2},
			Trans:        []*statecell.StateCell{in0},
			Status:       11,
			GasUsed:      99,
		}
	}

	t0 := time.Now()
	buffer, _ := Euresults(euresults).GobEncode()
	fmt.Println("EuResults GobEncode():", time.Now().Sub(t0))

	out := new(Euresults)
	out.GobDecode(buffer)
	fmt.Println("EuResults GobDecode():", time.Now().Sub(t0))
}
