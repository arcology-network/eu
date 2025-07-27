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
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

func TestCascadeDeletesSingleAccount(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	bob := BobAccount()
	eth.CreateDefaultPaths(1, alice, writeCache)
	eth.CreateDefaultPaths(1, bob, writeCache)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn/", commutative.NewPath())
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn/ele-00", noncommutative.NewString("ele-00"))

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", commutative.NewBoundedUint64(0, 100))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele1", commutative.NewBoundedUint64(0, 100))

	writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/storage/container/", commutative.NewPath())

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit([]uint64{1})
	committer.Commit(1)

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", new(commutative.Uint64)); v.(uint64) != 0 {
		t.Errorf("Expected 0, got %d", v.(uint64))
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele1", new(commutative.Uint64)); v.(uint64) != 0 {
		t.Errorf("Expected 0, got %d", v.(uint64))
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", new(commutative.Path)); v == nil {
		t.Errorf("Expected 0, got %d", v.(uint64))
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn/ele-00", new(noncommutative.String)); v.(string) != "ele-00" {
		t.Errorf("Expected 'ele-00', got %d", v)
	}

	// Use the wildcard path to delete all elements in the container
	wildcards := []*univalue.Univalue{univalue.NewUnivalue(1, "blcc://eth1.0/account/"+alice+"/storage/container/*", 0, 1, 0, nil, nil)}
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(wildcards)
	committer.Precommit([]uint64{1})

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele0", new(commutative.Uint64)); v != nil {
		t.Errorf("Expected nil, got %d", v)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ele1", new(commutative.Uint64)); v != nil {
		t.Errorf("Expected nil, got %d", v)
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn/ele-00", new(noncommutative.String)); v != nil {
		t.Errorf("Expected 'ele-00', got %d", v)
	}
}

func TestAddThenDeletePathAfterCommit(t *testing.T) {
	store := chooseDataStore().(*stgproxy.StorageProxy).DisableCache()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
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

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
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
	// v, univ, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/ctrn-0-0/elem-0-0:2", new(noncommutative.Int64))
	// if univ != nil {
	// 	t.Errorf("The element should have been deleted, got %v", v)
	// }
}
