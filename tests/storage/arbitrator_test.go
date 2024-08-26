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
	"math/rand"
	"testing"
	"time"

	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	"github.com/arcology-network/eu/eth"

	arbitrator "github.com/arcology-network/scheduler/arbitrator"
	statestore "github.com/arcology-network/storage-committer"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	// sstore:= statestore.NewStateStore(store.(*proxy.StorageProxy))
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)

	time.Sleep(2 * time.Second)

	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// accesses1, transitions1 := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	writeCache.Clear()

	bob := BobAccount()
	if _, err := eth.CreateNewAccount(2, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	arib := (&arbitrator.Arbitrator{})

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := arib.Detect(IDVec, append(accesses1, accesses2...))

	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()
	if len(conflictdict) != 0 {
		t.Error("Error: There should be NO conflict")
		accesses1.Print()
		accesses2.Print()
	}
}

func TestArbiCreateTwoAccounts1Conflict(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	meta := commutative.NewPath()
	writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, meta)
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(trans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)

	committer.SetStore(store)
	alice := AliceAccount()

	// = committer.WriteCache()
	if _, err := eth.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	path1 := commutative.NewPath()                                                // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	raw := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(raw)).To(univalue.IPTransition{})
	writeCache.Clear()

	// = committer.WriteCache()
	if _, err := eth.CreateNewAccount(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {

	// writeCache = committer.WriteCache()
	if _, err := eth.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	path2 := commutative.NewPath() // create a path
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := committer.WriteCache().Export(univalue.Sorter)
	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})

	// accesses1.Print()
	// fmt.Print(" ++++++++++++++++++++++++++++++++++++++++++++++++ ")
	// accesses2.Print()

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()

	if len(conflictdict) != 1 {
		t.Error("Error: There shouldn 1 conflict")
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// writeCache.Write(stgcommcommon.SYSTEM, stgcommcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)
	committer.SetStore(store)

	// committer.NewAccount(1, alice)

	if _, err := eth.CreateNewAccount(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", commutative.NewPath()) // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := writeCache.Export(univalue.Sorter)
	accesses1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	transitions1 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	// writeCache = committer.WriteCache()

	if _, err := eth.CreateNewAccount(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {
	path2 := commutative.NewPath() // create a path

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", path2)
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := committer.WriteCache().Export(univalue.Sorter)
	accesses2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	transitions2 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	IDVec := append(slice.Fill(make([]uint32, len(accesses1)), 0), slice.Fill(make([]uint32, len(accesses2)), 1)...)
	ids := (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses1, accesses2...))
	conflictDict, _, pairs := arbitrator.Conflicts(ids).ToDict()

	// pairs := arbitrator.Conflicts(ids).ToPairs()

	if len(conflictDict) != 1 || len(pairs) != 1 {
		t.Error("Error: There should be 1 conflict")
	}

	toCommit := slice.Exclude([]uint32{1, 2}, mapi.Keys(conflictDict))

	in := append(transitions1, transitions2...)
	buffer := univalue.Univalues(in).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(out)
	committer.Precommit(toCommit)
	committer.Commit(10)

	if _, err := writeCache.Write(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("committer-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := committer.Export(univalue.Sorter)
	accesses3 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	transitions3 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})

	// url4 := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	if _, err := writeCache.Write(4, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(univalue.Sorter)
	accesses4 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})
	transitions4 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	IDVec = append(slice.Fill(make([]uint32, len(accesses3)), 0), slice.Fill(make([]uint32, len(accesses4)), 1)...)
	ids = (&arbitrator.Arbitrator{}).Detect(IDVec, append(accesses3, accesses4...))
	conflictDict, _, _ = arbitrator.Conflicts(ids).ToDict()

	conflictTx := mapi.Keys(conflictDict)
	if len(conflictDict) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There should be only 1 conflict")
	}

	toCommit = slice.RemoveIf(&[]uint32{3, 4}, func(_ int, tx uint32) bool {
		// conflictTx := mapi.Keys(*conflictDict)

		_, ok := conflictDict[tx]
		return ok
	})

	buffer = univalue.Univalues(append(transitions3, transitions4...)).Encode()
	out = univalue.Univalues{}.Decode(buffer).(univalue.Univalues)

	acctTrans = append(transitions3, transitions4...)
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	// committer.Import(committer.Decode(univalue.Univalues(append(transitions3, transitions4...)).Encode()))

	committer.Precommit(toCommit)
	committer.Commit(10)
	committer.SetStore(store)

	v, _, _ := writeCache.Read(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", new(noncommutative.String))
	if v == nil || v.(string) != "committer-1-by-tx-3" {
		t.Error("Error: Wrong value, expecting:", "committer-1-by-tx-3 ", "actual:", v)
	}

	// have to mark balance and nonce persistent first !!!!!

	// v, _ = committer.Read(3, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	// if v == nil || v.(uint64) != 2 {
	// 	t.Error("Error: Wrong value, expecting:", "2", "actual:", v)
	// }
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	alice := AliceAccount()
	univalues := make([]*univalue.Univalue, 0, 5*200000)
	groupIDs := make([]uint32, 0, len(univalues))

	v := commutative.NewPath()
	for i := 0; i < len(univalues)/5; i++ {
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, univalue.NewUnivalue(uint32(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))

		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
		groupIDs = append(groupIDs, uint32(i))
	}

	t0 := time.Now()
	(&arbitrator.Arbitrator{}).Detect(groupIDs, univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
