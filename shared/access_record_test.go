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
	"fmt"
	"math/rand"
	"testing"
	"time"

	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	"github.com/arcology-network/common-lib/crdt/statecell"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/holiman/uint256"

	"encoding/json"
)

func RandomAccount() string {
	var letters = []byte("abcdef0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	addr := hexutil.Encode(b)
	return addr
}

func TestAccessRecordEncoding(t *testing.T) {

	alice := RandomAccount()

	u64 := commutative.NewBoundedUint64(0, 100)

	in0 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)

	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u256-000", 3, 4, 0, u256, nil)

	records := &TxAccessRecords{
		Hash: "0x1234567",
		ID:   99,
		Accesses: []*statecell.StateCell{
			in0, in1,
		},
	}

	buffer := records.Encode()
	out := (&TxAccessRecords{}).Decode(buffer)
	aj, _ := json.Marshal(records)
	bj, _ := json.Marshal(out)
	if !bytes.Equal(aj, bj) {
		t.Error("Error")
	}
}

func TestAccessRecordSetEncoding(t *testing.T) {
	alice := RandomAccount()
	u64 := commutative.NewBoundedUint64(0, 100)
	in0 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/u64-000", 3, 4, 0, u64, nil)
	_1 := &TxAccessRecords{
		Hash: "0x1234567",
		ID:   99,
		Accesses: []*statecell.StateCell{
			in0,
		},
	}

	bob := RandomAccount()
	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	in1 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+bob+"/storage/ctrn-0/u256-000", 3, 4, 0, u256, nil)
	_2 := &TxAccessRecords{
		Hash: "0xabcde",
		ID:   88,
		Accesses: []*statecell.StateCell{
			in1,
		},
	}

	comn := RandomAccount()
	u641 := commutative.NewBoundedUint64(0, 200)
	in2 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+comn+"/storage/ctrn-0/u64-000", 3, 4, 0, u641, nil)
	_3 := &TxAccessRecords{
		Hash: "0x8976542",
		ID:   77,
		Accesses: []*statecell.StateCell{
			in2,
		},
	}

	accessSet := TxAccessRecordSet{_1, _2, _3}

	buffer, _ := accessSet.GobEncode()
	out := TxAccessRecordSet{}
	out.GobDecode(buffer)
	for i := range accessSet {
		aj, _ := json.Marshal(accessSet[i])
		bj, _ := json.Marshal(out[i])
		if !bytes.Equal(aj, bj) {
			t.Error("Error")
		}
	}
}

func BenchmarkAccessRecordSetEncoding(b *testing.B) {
	recordVec := make([]*TxAccessRecords, 1000000)
	for i := 0; i < len(recordVec); i++ {
		comn := RandomAccount()
		u641 := commutative.NewBoundedUint64(0, uint64(200+i))
		in2 := statecell.NewStateCell(1, "blcc://eth1.0/account/"+comn+"/storage/ctrn-0/u64-000", 3, 4, 0, u641, nil)
		recordVec[i] = &TxAccessRecords{
			Hash: "0x1234567",
			ID:   uint64(i),
			Accesses: []*statecell.StateCell{
				in2,
			},
		}
	}

	t0 := time.Now()
	set := TxAccessRecordSet(recordVec)
	buffer, _ := (&set).GobEncode()
	fmt.Println("GobEncode():", time.Now().Sub(t0))

	out := new(TxAccessRecordSet)
	t0 = time.Now()
	out.GobDecode(buffer)
	fmt.Println("GobDecode():", time.Now().Sub(t0))
}
