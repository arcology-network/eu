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
	"bytes"
	"fmt"
	"testing"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	"github.com/arcology-network/eu/eth"

	statestore "github.com/arcology-network/storage-committer"
	stgcomm "github.com/arcology-network/storage-committer/common"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	statecell "github.com/arcology-network/storage-committer/type/statecell"
)

func TestRandomOrderImport(t *testing.T) {
	alice := AliceAccount()
	store := stgproxy.NewMemDBStoreProxy().EnableCache()
	sstore := statestore.NewStateStore(store)
	StateCache := sstore.StateCache

	if _, err := eth.CreateDefaultPaths(stgcomm.SYSTEM, alice, StateCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := statecell.StateCells(slice.Clone(StateCache.Export(statecell.Sorter))).To(statecell.IPTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{stgcomm.SYSTEM})
	committer.Commit(stgcomm.SYSTEM)

	fmt.Println(" ================================================= ")

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	// committer.Import(acctTrans)
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(2)

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(1), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}

	if _, err := sstore.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(2), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}

	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1, 2})
	committer.Commit(2)

	outV, _, _ := sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if outV == nil || len(outV.(*softdeltaset.DeltaSet[string]).Elements()) != 3 {
		t.Error("Error: The path should exist", outV)
	}

	sstore = statestore.NewStateStore(store)

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(3), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}

	if _, err := sstore.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(4), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}

	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})
	common.Swap(&acctTrans[0], &acctTrans[1])
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1, 2})
	committer.Commit(2)

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if outV == nil || len(outV.(*softdeltaset.DeltaSet[string]).Elements()) != 5 {
		t.Error("Error: The path should exist", outV)
	}
}

func commitToStateStore(sstore *statestore.StateStore, t *testing.T) {
	alice := AliceAccount()
	// sstore:= statestore.NewStateStore(store)

	if _, err := eth.CreateDefaultPaths(stgcomm.SYSTEM, alice, sstore.StateCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	committer := stgcommitter.NewStateCommitter(sstore.ReadOnlyStore(), sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{stgcomm.SYSTEM})
	committer.Commit(stgcomm.SYSTEM)

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	// committer.Import(acctTrans)
	committer = stgcommitter.NewStateCommitter(sstore.ReadOnlyStore(), sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(2)

	outV, _, _ := sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 199, 199})); err != nil {
		t.Error(err)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 199, 199}) {
		t.Error("Error: The path should exist", outV)
	}

	// Delete the entry
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), nil); err != nil {
		t.Error(err)
	}

	// Delete the entry
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), noncommutative.NewBytes([]byte{77, 77})); err != nil {
		t.Error(err)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(2)

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{77, 77}) {
		t.Error("Error: The path should not exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path))
	if outV == nil || len(outV.(*softdeltaset.DeltaSet[string]).Elements()) != 1 {
		t.Error("Error: The path should exist", outV)
	}

}

func TestCommitToStatStore(t *testing.T) {
	// commitToStateStore(stgproxy.NewMemDBStoreProxy().EnableCache(), t) // Use cache

	sstore := statestore.NewStateStore(stgproxy.NewMemDBStoreProxy().EnableCache())
	// store := statestore.NewStateStore(Proxy)
	commitToStateStore(sstore, t)
}

func TestAsyncCommitToStateStore(t *testing.T) {
	alice := AliceAccount()
	store := stgproxy.NewMemDBStoreProxy().EnableCache()
	sstore := statestore.NewStateStore(store)
	StateCache := sstore.StateCache

	if _, err := eth.CreateDefaultPaths(stgcomm.SYSTEM, alice, StateCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := statecell.StateCells(slice.Clone(StateCache.Export(statecell.Sorter))).To(statecell.IPTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{stgcomm.SYSTEM})
	committer.Commit(stgcomm.SYSTEM)

	fmt.Println(" ================================================= ")

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
		t.Error(err)
	}
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
		t.Error(err)
	}
	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	// committer.Import(acctTrans)
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(2)

	fmt.Println(" ================================================= ")
	outV, _, _ := sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
		t.Error("Error: The path should exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
		t.Error("Error: The path should exist", outV)
	}

	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 199, 199})); err != nil {
		t.Error(err)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 199, 199}) {
		t.Error("Error: The path should exist", outV)
	}

	// Delete the entry
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), nil); err != nil {
		t.Error(err)
	}

	// Delete the entry
	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), noncommutative.NewBytes([]byte{77, 77})); err != nil {
		t.Error(err)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	acctTrans = statecell.StateCells(slice.Clone(sstore.Export(statecell.Sorter))).To(statecell.IPTransition{})

	// committer = statestore.NewStateStore(store)
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(2) //.Clear()
	fmt.Println(" ================================================= ")
	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
	if outV != nil {
		t.Error("Error: The path should not exist", outV)
	}

	outV, _, _ = sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(77), new(noncommutative.Bytes))
	if outV == nil || !bytes.Equal(outV.([]byte), []byte{77, 77}) {
		t.Error("Error: The path should not exist", outV)
	}
}
