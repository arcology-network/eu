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
package stgtest

import (
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

func TestGrowOnlySet(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet([]byte("11")))
	if err != nil {
		t.Error("Failed to write prepay function", err)
	}

	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet([]byte("22")))
	if err != nil {
		t.Error("Failed to write prepay function", err)
	}

	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet([]byte("33")))
	if err != nil {
		t.Error("Failed to write prepay function", err)
	}

	val, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", new(commutative.GrowOnlySet[[]byte]))
	if !reflect.DeepEqual(val, [][]byte{[]byte("11"), []byte("22"), []byte("33")}) {
		t.Error("Error: Wrong value in grow-only set")
	}
}

// func GenInfo() (*gas.PrepayerInfo, *gas.PrepayerInfo) {
// 	Alice := &gas.PrepayerInfo{}
// 	Alice.Hash = [32]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
// 		0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
// 		0x10, 0x11, 0x12, 0x13, 0x14,
// 		0x15, 0x16, 0x17, 0x18, 0x19, 0x1a,
// 		0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

// 	Alice.TX = 12345
// 	Alice.From = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
// 		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
// 		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}

// 	Alice.To = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
// 		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
// 		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}
// 	Alice.Signature = [4]byte{0x12, 0x34, 0x56, 0x78}
// 	Alice.GasUsed = 100000
// 	Alice.InitialGas = 200000
// 	Alice.PrepayedAmount = 50000
// 	Alice.GasRemaining = 150000
// 	Alice.Successful = true

// 	Alice2 := &gas.PrepayerInfo{}
// 	Alice2.Hash = [32]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26,
// 		0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
// 		0x30,
// 		0x31, 0x32, 0x33, 0x34, 0x35,
// 		0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b,
// 		0x3c, 0x3d, 0x3e,
// 		0x3f, 0x40}

// 	Alice2.TX = 54321

// 	Alice2.From = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
// 		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
// 		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

// 	Alice2.To = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
// 		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
// 		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

// 	Alice2.Signature = [4]byte{0x56, 0x78, 0x9a, 0xbc}
// 	Alice2.GasUsed = 200000
// 	Alice2.InitialGas = 300000
// 	Alice2.PrepayedAmount = 100000
// 	Alice2.GasRemaining = 200000
// 	Alice2.Successful = false

// 	return Alice, Alice2
// }

// func TestGrowOnlySetWithPayerInfo(t *testing.T) {
// 	alice, AliceWriteCache, _, AliceErr := GenerateDB([20]byte(new(codec.Bytes20).Fill(10))) // Generate a database with an account filled with 10 bytes of data
// 	if AliceErr != nil {
// 		t.Error("Failed to generate Alice's database", AliceErr)
// 	}

// 	bob, BobWriteCache, _, BobErr := GenerateDB([20]byte(new(codec.Bytes20).Fill(10))) // Generate a database with an account filled with 10 bytes of data {
// 	if BobErr != nil {
// 		t.Error("Failed to generate Alice2's database", BobErr)
// 	}

// 	Alice, Alice2 := GenInfo()

// 	_, err := AliceWriteCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet(Alice.Encode()))
// 	a := (stgcommon.Type)(commutative.NewGrowOnlyByteSet(Alice.Encode()))
// 	if err != nil {
// 		fmt.Println(a)
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	_, err = BobWriteCache.Write(2, "blcc://eth1.0/account/"+bob+"/func/prepayer", commutative.NewGrowOnlyByteSet(Alice2.Encode()))
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	store := chooseDataStore()
// 	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
// 	// writeCache := sstore.WriteCache

// 	alice1Trans := slice.Clone(AliceWriteCache.Export(univalue.Sorter))
// 	bobTrans := slice.Clone(BobWriteCache.Export(univalue.Sorter))

// 	acctTrans := univalue.Univalues(append(alice1Trans, bobTrans...)).To(univalue.ITTransition{})
// 	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Precommit([]uint64{1, 2, stgcommcommon.SYSTEM})
// 	committer.Commit(10)

// }

// func TestGrowOnlySetWithPayerInfoSavePath(t *testing.T) {
// 	alice, AliceWriteCache, _, AliceErr := GenerateDB([20]byte(new(codec.Bytes20).Fill(10))) // Generate a database with an account filled with 10 bytes of data
// 	if AliceErr != nil {
// 		t.Error("Failed to generate Alice's database", AliceErr)
// 	}

// 	Alice, Alice2 := GenInfo()

// 	_, err := AliceWriteCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet(Alice.Encode()))
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	_, BobWriteCache, _, BobErr := GenerateDB([20]byte(new(codec.Bytes20).Fill(10))) // Generate a database with an account filled with 10 bytes of data {
// 	if BobErr != nil {
// 		t.Error("Failed to generate Alice2's database", BobErr)
// 	}

// 	_, err = BobWriteCache.Write(2, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlyByteSet(Alice2.Encode()))
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	store := chooseDataStore()
// 	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
// 	// writeCache := sstore.WriteCache

// 	alice1Trans := slice.Clone(AliceWriteCache.Export(univalue.Sorter))
// 	bobTrans := slice.Clone(BobWriteCache.Export(univalue.Sorter))

// 	acctTrans := univalue.Univalues(append(alice1Trans, bobTrans...)).To(univalue.ITTransition{})
// 	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Precommit([]uint64{1, 2, stgcommcommon.SYSTEM})
// 	committer.Commit(10)

// }
