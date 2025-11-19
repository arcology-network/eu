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
	"fmt"
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/state-engine"
	stgcommon "github.com/arcology-network/state-engine/common"
	cache "github.com/arcology-network/state-engine/state/cache"
	statecommitter "github.com/arcology-network/state-engine/state/committer"
	stgproxy "github.com/arcology-network/state-engine/storage/proxy"
	"github.com/arcology-network/state-engine/type/commutative"
	"github.com/arcology-network/state-engine/type/noncommutative"
	statecell "github.com/arcology-network/state-engine/type/statecell"
)

func TestAddThenDeletePathAfterCommit(t *testing.T) {
	store := chooseDataStore().(*stgproxy.StorageProxy).DisableCache()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.StateCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/elem-0-0:2", noncommutative.NewInt64(99))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem:0", noncommutative.NewInt64(33))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem:1", noncommutative.NewInt64(11))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/elem:2", noncommutative.NewInt64(22))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-1/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-1/ctrn-1-0/", commutative.NewPath())

	/*
		├── ctrn-0/
		│   ├── ctrn-0-0/
		│   │   └── elem-0-0:2 = 88
		│   ├── elem-0-0 = 33
		│   ├── elem:1    = 11
		│   └── elem:2    = 22
		└── ctrn-1/
		    └── ctrn-1-0/
	*/

	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.IPTransition{})
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(1)

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	elems := v.(*softdeltaset.DeltaSet[string]).Elements()
	if v == nil || !reflect.DeepEqual(elems, []string{"ctrn-0-0/", "elem:0", "elem:1", "elem:2"}) {
		fmt.Println(elems)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-1/", commutative.NewPath())
	if v == nil || !reflect.DeepEqual(v.(*softdeltaset.DeltaSet[string]).Elements(), []string{"ctrn-1-0/"}) {
		t.Errorf("Expected nil, got %d", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/", commutative.NewPath())
	if v == nil || !reflect.DeepEqual(v.(*softdeltaset.DeltaSet[string]).Elements(), []string{"elem-0-0:2"}) {
		t.Errorf("Expected nil, got %d", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	elems = v.(*softdeltaset.DeltaSet[string]).Elements()
	if v == nil || !reflect.DeepEqual(elems, []string{"ctrn-0-0/", "elem:0", "elem:1", "elem:2"}) {
		fmt.Println(elems)
	}

	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/", nil)
	if err != nil {
		t.Error(err)
	}

	// Check the deleted the path
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/elem-0-0:2", new(noncommutative.Int64))
	if v != nil {
		t.Errorf("The element should have been deleted, got %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	elems = v.(*softdeltaset.DeltaSet[string]).Elements()
	if v == nil || !reflect.DeepEqual(elems, []string{"elem:0", "elem:1", "elem:2"}) {
		t.Errorf("Wrong elements after delete: %v", elems)
	}

	// Delete the path
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/elem-0-0:2", new(noncommutative.Int64))
	if v != nil {
		t.Errorf("The element should have been deleted, got %v", v)
	}

	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", nil)
	if err != nil {
		t.Error(err)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath())
	if v != nil {
		t.Errorf("Wrong elements after delete: %v", elems)
	}

	acctTrans = statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.IPTransition{})
	committer = statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(1)

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	elems = v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(elems, []string{"ctrn-1/"}) {
		t.Errorf("Wrong elements after delete: %v", elems)
	}
}

func CommitToDBs(writeCache *cache.StateCache, store *stgproxy.StorageProxy, sstore *statestore.StateStore, filter any) []*statecell.StateCell {
	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(filter)
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(1)
	writeCache.Clear()
	return acctTrans
}

func TestAllUnderGrantParentPathWildcardSimplest(t *testing.T) {
	store := chooseDataStore().(*stgproxy.StorageProxy).DisableCache()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.StateCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", noncommutative.NewInt64(33))

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	// Delete all elements under the path with wildcard
	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/*", nil)
	if err != nil {
		t.Error(err)
	}

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	elems := v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(elems, []string{}) {
		t.Errorf("The path should be empty: %v", elems)
	}
}

func TestAllUnderGrantParentPathWildcardSimple(t *testing.T) {
	store := chooseDataStore().(*stgproxy.StorageProxy).DisableCache()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.StateCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", noncommutative.NewInt64(33))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", noncommutative.NewInt64(11))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", noncommutative.NewInt64(22))

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	// Delete all elements under the path with wildcard
	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/*", nil)
	if err != nil {
		t.Error(err)
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", new(noncommutative.Int64))
	if v != nil {
		t.Errorf("The element should have been deleted already: %v", v)
	}

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	elems := v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(elems, []string{}) {
		t.Errorf("The path should be empty: %v", elems)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", new(noncommutative.Int64))
	if v != nil {
		t.Errorf("The element should have been deleted already: %v", v)
	}
}

func TestAllUnderGrantParentPathWildcard(t *testing.T) {
	store := chooseDataStore().(*stgproxy.StorageProxy).DisableCache()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.StateCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", noncommutative.NewInt64(33))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", noncommutative.NewInt64(11))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", noncommutative.NewInt64(22))

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	// Delete all elements under the path with wildcard
	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/*", nil)
	if err != nil {
		t.Error(err)
	}

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	// Check the elements are back
	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	elems := v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(elems, []string{}) {
		t.Errorf("Wrong elements after delete: %v", elems)
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", new(noncommutative.Int64)); v != nil {
		t.Errorf("The element should have been deleted: %v", v)
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", new(noncommutative.Int64)); v != nil {
		t.Errorf("The element should have been deleted: %v", v)
	}

	if v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", new(noncommutative.Int64)); v != nil {
		t.Errorf("The element should have been deleted: %v", v)
	}

	// Write the elements back.
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", noncommutative.NewInt64(133)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", noncommutative.NewInt64(111)); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", noncommutative.NewInt64(122)); err != nil {
		t.Error(err)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", new(noncommutative.Int64))
	if v == nil || v.(int64) != 133 {
		t.Errorf("Wrong elements after delete: %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", new(noncommutative.Int64))
	if v == nil || v.(int64) != 111 {
		t.Errorf("Wrong elements after delete: %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", new(noncommutative.Int64))
	if v == nil || v.(int64) != 122 {
		t.Errorf("Wrong elements after delete: %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	elems = v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(elems, []string{"elem:0", "elem:1", "elem:2"}) {
		t.Errorf("Wrong elements after delete: %v", elems)
	}

	CommitToDBs(writeCache, store, sstore, statecell.IPTransition{})

	// 	It is in the execCache, meaning committion was successful.
	// but the line below cannot load the right meta, containing the elements.

	// 1. Get the value from the execStrong despite the fact it exists in the cache.
	// 2. The stagedRemoval did not seem to be cleared properly before committing to the storage.

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:0", new(noncommutative.Int64))
	if v == nil || v.(int64) != 133 {
		t.Errorf("Wrong elements after delete: %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:1", new(noncommutative.Int64))
	if v == nil || v.(int64) != 111 {
		t.Errorf("Wrong elements after delete: %v", v)
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/elem:2", new(noncommutative.Int64))
	if v == nil || v.(int64) != 122 {
		t.Errorf("Wrong elements after delete: %v", v)
	}
}
