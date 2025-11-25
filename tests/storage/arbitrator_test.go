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
	"strings"
	"testing"
	"time"

	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	noncommutative "github.com/arcology-network/common-lib/crdt/noncommutative"
	statecell "github.com/arcology-network/common-lib/crdt/statecell"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/slice"
	ethadaptor "github.com/arcology-network/eu/ethadaptor"
	stgcommon "github.com/arcology-network/state-engine/common"

	arbitrator "github.com/arcology-network/scheduler/arbitrator"
	statestore "github.com/arcology-network/state-engine"
	statecommitter "github.com/arcology-network/state-engine/state/committer"
	"github.com/arcology-network/state-engine/storage/proxy"
	stgproxy "github.com/arcology-network/state-engine/storage/proxy"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	meta := commutative.NewPath()
	writeCache.Write(stgcommon.SYSTEM, stgcommon.ETH_ACCOUNT_PREFIX, meta)
	trans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	// sstore:= statestore.NewStateStore(store.(*proxy.StorageProxy))
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(trans).Encode()).(statecell.StateCells))

	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.DebugCommit(10)

	// Create Alice account
	alice := AliceAccount()
	if _, err := ethadaptor.CreateDefaultPaths(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	accesses1 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITAccess{})
	statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	writeCache.Clear()

	// Create Bob account
	bob := BobAccount()
	if _, err := ethadaptor.CreateDefaultPaths(2, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	accesses2 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITAccess{})
	statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	// Initiate the arbitrator and detect conflicts
	arib := arbitrator.NewArbitrator()
	IDVec := append(slice.Fill(make([]uint64, len(accesses1)), 0), slice.Fill(make([]uint64, len(accesses2)), 1)...)
	ids := arib.InsertAndDetect(IDVec, append(accesses1, accesses2...))

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
	writeCache := sstore.StateCache

	meta := commutative.NewPath()
	writeCache.Write(stgcommon.SYSTEM, stgcommon.ETH_ACCOUNT_PREFIX, meta)
	initTrans := writeCache.Export(statecell.Sorter)
	trans := statecell.StateCells(slice.Clone(initTrans)).To(statecell.ITTransition{})

	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(trans).Encode()).(statecell.StateCells))

	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.DebugCommit(10)

	committer.SetStore(store)
	alice := AliceAccount()

	// = committer.StateCache()
	if _, err := ethadaptor.CreateDefaultPaths(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	path1 := commutative.NewPath()                                                // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path1) // create a path
	raw := writeCache.Export(statecell.Sorter)
	tr := slice.Clone(raw)
	accesses1 := statecell.StateCells(tr).To(statecell.IPAccess{})
	tr = tr[:1]
	tr = statecell.StateCells(tr).To(statecell.IPAccess{})
	writeCache.Clear()

	// = committer.StateCache()
	if _, err := ethadaptor.CreateDefaultPaths(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {

	// writeCache = committer.StateCache()
	if _, err := ethadaptor.CreateDefaultPaths(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	path2 := commutative.NewPath() // create a path
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/", path2)
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	// committer.Write(2, "blcc://eth1.0/account/"+alice+"/storage/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))
	// accesses2, _ := committer.StateCache().Export(statecell.Sorter)
	accesses2 := writeCache.Export(statecell.Sorter)
	accesses2 = accesses2[:1]
	accesses2 = statecell.StateCells(accesses2).To(statecell.ITAccess{})
	accesses2 = writeCache.Export(statecell.Sorter)
	accesses2 = slice.Clone(writeCache.Export(statecell.Sorter))
	accesses2 = statecell.StateCells(accesses2).To(statecell.ITAccess{})

	// accesses1.Print()
	// fmt.Print(" ++++++++++++++++++++++++++++++++++++++++++++++++ ")
	// accesses2.Print()

	IDVec := append(slice.Fill(make([]uint64, len(accesses1)), 0), slice.Fill(make([]uint64, len(accesses2)), 1)...)
	ids := arbitrator.NewArbitrator().InsertAndDetect(IDVec, append(accesses1, accesses2...))
	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()

	if len(conflictdict) != 1 {
		t.Error("Error: There shouldn 1 conflict, actual: ", len(conflictdict))
	}
}

func TestArbiTwoTxModifyTheSameAccount(t *testing.T) {
	store := chooseDataStore()

	alice := AliceAccount()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	if _, err := ethadaptor.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// writeCache.Write(stgcommon.SYSTEM, stgcommon.ETH_ACCOUNT_PREFIX, commutative.NewPath())
	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))

	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.DebugCommit(10)
	committer.SetStore(store)

	// committer.NewAccount(1, alice)

	if _, err := ethadaptor.CreateDefaultPaths(1, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", commutative.NewPath()) // create a path
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-1"))
	// accesses1, transitions1 := writeCache.Export(statecell.Sorter)
	accesses1 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITAccess{})
	transitions1 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	// writeCache = committer.StateCache()

	if _, err := ethadaptor.CreateDefaultPaths(2, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	} // NewAccount account structure {
	path2 := commutative.NewPath() // create a path

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/", path2)
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-1-by-tx-2"))
	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("value-2-by-tx-2"))

	// accesses2, transitions2 := committer.StateCache().Export(statecell.Sorter)
	accesses2 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITAccess{})
	transitions2 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	IDVec := append(slice.Fill(make([]uint64, len(accesses1)), 0), slice.Fill(make([]uint64, len(accesses2)), 1)...)
	ids := arbitrator.NewArbitrator().InsertAndDetect(IDVec, append(accesses1, accesses2...))
	conflictDict, _, pairs := arbitrator.Conflicts(ids).ToDict()

	// pairs := arbitrator.Conflicts(ids).ToPairs()

	if len(conflictDict) != 1 || len(pairs) != 1 {
		t.Error("Error: There should be 1 conflict")
	}

	toCommit := slice.Exclude([]uint64{1, 2}, mapi.Keys(conflictDict))

	in := append(transitions1, transitions2...)
	buffer := statecell.StateCells(in).Encode()
	out := statecell.StateCells{}.Decode(buffer).(statecell.StateCells)

	committer = statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(out)
	committer.Precommit(toCommit)
	committer.DebugCommit(10)

	if _, err := writeCache.Write(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("committer-1-by-tx-3")); err != nil {
		t.Error(err)
	}

	// accesses3, transitions3 := committer.Export(statecell.Sorter)
	exports := writeCache.Export(statecell.Sorter)
	accesses3 := statecell.StateCells(slice.Clone(exports)).To(statecell.ITAccess{})
	transitions3 := statecell.StateCells(slice.Clone(exports)).To(statecell.IPTransition{})

	// url4 := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	if _, err := writeCache.Write(4, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", noncommutative.NewString("url4-1-by-tx-3")); err != nil {
		t.Error(err)
	}
	// accesses4, transitions4 := url4.Export(statecell.Sorter)
	accesses4 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITAccess{})
	transitions4 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	IDVec = append(slice.Fill(make([]uint64, len(accesses3)), 0), slice.Fill(make([]uint64, len(accesses4)), 1)...)
	ids = arbitrator.NewArbitrator().InsertAndDetect(IDVec, append(accesses3, accesses4...))
	conflictDict, _, _ = arbitrator.Conflicts(ids).ToDict()

	conflictTx := mapi.Keys(conflictDict)
	if len(conflictDict) != 1 || conflictTx[0] != 4 {
		t.Error("Error: There should be only 1 conflict", "actual:", len(conflictDict))
	}

	toCommit = slice.RemoveIf(&[]uint64{3, 4}, func(_ int, tx uint64) bool {
		// conflictTx := mapi.Keys(*conflictDict)
		_, ok := conflictDict[tx]
		return ok
	})

	buffer = statecell.StateCells(append(transitions3, transitions4...)).Encode()
	out = statecell.StateCells{}.Decode(buffer).(statecell.StateCells)

	acctTrans = append(transitions3, transitions4...)
	committer = statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))

	// committer.Import(committer.Decode(statecell.StateCells(append(transitions3, transitions4...)).Encode()))

	committer.Precommit(toCommit)
	committer.DebugCommit(10)
	committer.SetStore(store)

	v, _, _ := writeCache.Read(3, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-2/elem-1", new(noncommutative.String))
	if v == nil || v.(string) != "committer-1-by-tx-3" {
		t.Error("Error: Wrong value, expecting:", "committer-1-by-tx-3 ", "actual:", v)
	}
}

func TestArbiWildcardConflict(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.StateCache
	alice := AliceAccount()
	ethadaptor.CreateDefaultPaths(1, alice, writeCache)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", commutative.NewBoundedUint64(0, 100))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele1", commutative.NewBoundedUint64(0, 100))

	accesses1 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.IPTransition{})
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(accesses1)
	committer.Precommit([]uint64{1})
	committer.DebugCommit(1)
	writeCache.Clear()

	// this should conflict

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/[:]", nil)
	raws := writeCache.Export(statecell.Sorter)
	accesses2 := statecell.StateCells(slice.Clone(raws)).To(statecell.IPTransition{})

	// accesses2.Print()
	acctTrans1 := []*statecell.StateCell(accesses1)
	slice.RemoveIf(&acctTrans1, func(_ int, v *statecell.StateCell) bool {
		return !strings.Contains(*v.GetPath(), "/container/")
	})

	acctTrans2 := []*statecell.StateCell(accesses2)
	slice.RemoveIf(&acctTrans2, func(_ int, v *statecell.StateCell) bool {
		return !strings.Contains(*v.GetPath(), "/container/")
	})

	arib := arbitrator.NewArbitrator()
	IDVec := append(slice.Fill(make([]uint64, len(acctTrans1)), 0), slice.Fill(make([]uint64, len(acctTrans2)), 1)...)
	ids := arib.InsertAndDetect(IDVec, append(acctTrans1, acctTrans2...))
	conflictdict, _, _ := arbitrator.Conflicts(ids).ToDict()
	if len(conflictdict) != 0 {
		t.Error("Error: There should be one conflict, actual:", len(conflictdict))
		statecell.StateCells(acctTrans1).Print()
		statecell.StateCells(acctTrans2).Print()
	}
}

func BenchmarkSimpleArbitrator(b *testing.B) {
	alice := AliceAccount()
	univalues := make([]*statecell.StateCell, 0, 5*200000)
	groupIDs := make([]uint64, 0, len(univalues))

	v := commutative.NewPath()
	for i := 0; i < len(univalues)/5; i++ {
		univalues = append(univalues, statecell.NewStateCell(uint64(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, statecell.NewStateCell(uint64(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, statecell.NewStateCell(uint64(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, statecell.NewStateCell(uint64(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))
		univalues = append(univalues, statecell.NewStateCell(uint64(i), "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000"+fmt.Sprint(rand.Float32()), 1, 0, 0, v, nil))

		groupIDs = append(groupIDs, uint64(i))
		groupIDs = append(groupIDs, uint64(i))
		groupIDs = append(groupIDs, uint64(i))
		groupIDs = append(groupIDs, uint64(i))
		groupIDs = append(groupIDs, uint64(i))
	}

	t0 := time.Now()
	arbitrator.NewArbitrator().InsertAndDetect(groupIDs, univalues)
	fmt.Println("Detect "+fmt.Sprint(len(univalues)), "path in ", time.Since(t0))
}
