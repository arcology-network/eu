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

package stgtest

import (
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	codec "github.com/arcology-network/common-lib/codec"

	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	rlp "github.com/ethereum/go-ethereum/rlp"
)

func TestNoncommutativeCodec(t *testing.T) {
	/* Noncommutative Path Test*/
	inPath := commutative.NewPath()

	pathBytes := inPath.(*commutative.Path).Encode()
	outPath := (&commutative.Path{}).Decode(pathBytes)
	inPath.(*commutative.Path).Equal(outPath.(*commutative.Path))

	/* Noncommutative String Test */
	inStr := noncommutative.NewString("ctrn-0")
	strBytes := inStr.(*noncommutative.String).Encode()
	stringer := noncommutative.String("")
	outStr := stringer.Decode(strBytes)
	if !reflect.DeepEqual(inStr, outStr) {
		t.Error("Error: String Encoding/decoding error, strings don't match")
	}

	/*  []byte Test */
	inBytes := noncommutative.NewBytes([]byte("test bytes"))
	encoded := inBytes.(*noncommutative.Bytes).Encode()

	result := (&noncommutative.Bytes{}).Decode(encoded)
	outBytes := noncommutative.Bytes(*(result.(*noncommutative.Bytes)))

	if !reflect.DeepEqual(inBytes, &outBytes) {
		t.Error("Error: Bytes Encoding/decoding error, bytes don't match")
	}

	/* Int64 Test */
	inInt64 := noncommutative.NewInt64(12345)
	int64Bytes := inInt64.Encode()
	int64cdc := noncommutative.Int64(0)
	outInt64 := (&int64cdc).Decode(int64Bytes).(*noncommutative.Int64)
	if !reflect.DeepEqual(inInt64, outInt64) {
		t.Error("Error: Int64 Encoding/decoding error, numbers don't match")
	}

	/* Bigint Test */
	inBig := noncommutative.NewBigint(789456)
	bigBytes := inBig.(*noncommutative.Bigint).Encode()
	outBig := (&noncommutative.Bigint{}).Decode(bigBytes).(*noncommutative.Bigint)
	if !reflect.DeepEqual(inBig, outBig) {
		t.Error("Error: Bigint Encoding/decoding error, numbers don't match")
	}
}

// func TestUnivalueCodec(t *testing.T) {
// 	store := ccstorage.NewDataStore(nil, nil, stgtypecommonCodec{}.Encode, stgtypecommonCodec{}.Decode)
// 	transitions := []*statecell.StateCell{}

// 	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
// 	writeCache := sstore.StateCache
// 	// committer.NewAccount(stgcommon.SYSTEM, fmt.Sprint("rand.Int()"))
// 	eucommon.CreateNewAccount(stgcommon.SYSTEM, fmt.Sprint("rand.Int()"), writeCache)

// 	transVec := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.IPTransition{})
// 	transitions = append(transitions, transVec...)

// 	for i := 0; i < len(transitions); i++ {
// 		buffer := transitions[i].Encode()
// 		in := transitions[i]
// 		out := (&statecell.StateCell{}).Decode(buffer).(*statecell.StateCell)
// 		// out.(*statecell.StateCell).ClearReserve()

// 		if !in.Equal(out) {
// 			t.Error("Error: Missmatched")
// 		}
// 	}
// }

// func TestUnivaluesCodec(t *testing.T) {
// 	store := ccstorage.NewDataStore(nil, nil, stgtypecommonCodec{}.Encode, stgtypecommonCodec{}.Decode)
// 	transitions := []*statecell.StateCell{}
// 	for i := 0; i < 10; i++ {
// 		acct := RandomAccount()

// 		sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
// 		writeCache := sstore.StateCache
// 		// committer.NewAccount(stgcommon.SYSTEM, acct)

// 		eucommon.CreateNewAccount(stgcommon.SYSTEM, acct, writeCache)

// 		transVec := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
// 		transitions = append(transitions, transVec...)
// 	}
// 	t0 := time.Now()
// 	buffer := statecell.StateCells(transitions).Encode()
// 	fmt.Println("Encode() ", len(transitions), " univalue in :", time.Since(t0))

// 	t0 = time.Now()
// 	out := (statecell.StateCells([]*statecell.StateCell{})).Decode(buffer).(statecell.StateCells)
// 	fmt.Println("Decode() ", len(transitions), " univalue in :", time.Since(t0))

// 	for i := 0; i < len(transitions); i++ {
// 		tran := transitions[i]
// 		out[i].ClearCache()
// 		if !reflect.DeepEqual(tran, out[i]) {
// 			fmt.Println("Error: Missmatched")
// 		}
// 	}
// }

func BenchmarkRlpComparePerformance(t *testing.B) {
	num := big.NewInt(100)

	expected, err := rlp.EncodeToBytes(num)
	if err != nil {
		t.Error(expected, err)
	}

	var decoded big.Int
	if err := rlp.DecodeBytes(expected, &decoded); err != nil {
		t.Error(expected, err)
	}

	if num.Cmp(&decoded) != 0 {
		t.Error("Mismatch")
	}

	t0 := time.Now()
	for i := 0; i < 1000000; i++ {
		num = big.NewInt(100)
	}
	fmt.Println("big NewInt RLP Encode:            "+fmt.Sprint(1000000), time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 1000000; i++ {
		v := codec.Bigint(*num)
		v.Encode()
	}
	fmt.Println("big NewInt Codec Encode:            "+fmt.Sprint(1000000), time.Since(t0))
}
