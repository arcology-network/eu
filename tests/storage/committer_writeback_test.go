/*
 *   Copyright (c) 2023 Arcology Network

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

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	platform "github.com/arcology-network/storage-committer/platform"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
	"github.com/holiman/uint256"
)

func TestEmptyNodeSet(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	// acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{})
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{})
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)
}

func TestRecursiveDeletionSameBatch(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)
	committer.SetStore(store)
	// create a path
	path := commutative.NewPath()
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", path)
	// _, addPath := writeCache.Export(univalue.Sorter)
	addPath := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(addPath).Encode()).(univalue.Univalues))
	// committer.Import(committer.Decode(univalue.Univalues(addPath).Encode()))

	committer.Precommit([]uint64{1})
	committer.Commit(10)
	committer.SetStore(store)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/1", noncommutative.NewInt64(1))
	// _, addTrans := writeCache.Export(univalue.Sorter)
	addTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	// committer.Import(committer.Decode(univalue.Univalues(addTrans).Encode()))
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(addTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint64{1})
	committer.Commit(10)

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.Int64)); v == nil {
		t.Error("Error: Failed to read the key !")
	}

	// url2 := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.Int64)); v.(int64) != 1 {
		t.Error("Error: Failed to read the key !")
	}

	writeCache.Write(2, "blcc://eth1.0/account/"+alice+"/storage/container/1", nil)
	// _, deleteTrans := url2.WriteCache().Export(univalue.Sorter)
	deleteTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to read the key !")
	}

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(append(addTrans, deleteTrans...))

	committer.Precommit([]uint64{1, 2})
	committer.Commit(10)

	if v, _, _ := writeCache.Read(2, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.Int64)); v != nil {
		t.Error("Error: Failed to delete the entry !")
	}
}

func TestApplyingTransitionsFromMulitpleBatches(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans)

	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	committer.SetStore(store)
	// path := commutative.NewPath()
	// _, err := writeCache.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path)

	// if err != nil {
	// 	t.Error("error")
	// }

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint64{1})
	committer.Commit(10)

	committer.SetStore(store)

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/4", nil)

	if acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{}); len(acctTrans) != 0 {
		t.Error("Error: Wrong number of transitions")
	}
}

func TestRecursiveDeletionDifferentBatch(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	in := univalue.Univalues(acctTrans).Encode()
	out := univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// committer.Import(committer.Decode(univalue.Univalues(out).Encode()))

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))

	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	committer.SetStore(store)
	// create a path
	path := commutative.NewPath()

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", path)
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/1", noncommutative.NewString("1"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/2", noncommutative.NewString("2"))

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	in = univalue.Univalues(acctTrans).Encode()
	out = univalue.Univalues{}.Decode(in).(univalue.Univalues)
	// committer.Import(committer.Decode(univalue.Univalues(out).Encode()))
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(out).Encode()).(univalue.Univalues))

	committer.Precommit([]uint64{1})
	committer.Commit(10)

	writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	_1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.String))
	if _1 != "1" {
		t.Error("Error: Not match")
	}

	committer.SetStore(store)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/1", noncommutative.NewString("3")); err != nil {
		t.Error(err)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/3", noncommutative.NewString("13"))
	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/4", noncommutative.NewString("14"))

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/2", noncommutative.NewString("4"))

	outpath, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/", &commutative.Path{})
	keys := outpath.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"1", "2", "3", "4"}) {
		t.Error("Error: Not match")
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/", nil) // delete the path
	if acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{}); len(acctTrans) != 3 {
		t.Error("Error: Wrong number of transitions")
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/1", new(noncommutative.String)); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

func TestStateUpdate(t *testing.T) {
	store := chooseDataStore()
	// store := storage.NewDataStore( nil, nil, platform.Codec{}.Encode, platform.Codec{}.Decode)

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	// _, initTrans := writeCache.Export(univalue.Sorter)
	initTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	// committer.Import(committer.Decode(univalue.Univalues(initTrans).Encode()))
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(initTrans).Encode()).(univalue.Univalues))

	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)
	committer.SetStore(store)

	tx0bytes, trans, err := Create_Ctrn_0(alice, store)
	if err != nil {
		t.Error(err)
	}
	tx0Out := univalue.Univalues{}.Decode(tx0bytes).(univalue.Univalues)
	tx0Out = trans
	tx1bytes, err := Create_Ctrn_1(alice, store)
	if err != nil {
		t.Error(err)
	}

	tx1Out := univalue.Univalues{}.Decode(tx1bytes).(univalue.Univalues)

	// committer.Import(committer.Decode(univalue.Univalues(append(tx0Out, tx1Out...)).Encode()))
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import((append(tx0Out, tx1Out...)))

	committer.Precommit([]uint64{0, 1})
	committer.Commit(10)
	//need to encode delta only now it encodes everything

	if err := CheckPaths(alice, writeCache); err != nil {
		t.Error(err)
	}

	v, _, _ := writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/", &commutative.Path{}) //system doesn't generate sub paths for /storage/
	// if v.(*commutative.Path).CommittedLength() != 2 {
	// 	t.Error("Error: Wrong sub paths")
	// }

	// if !reflect.DeepEqual(v.([]string), []string{"ctrn-0/", "ctrn-1/"}) {
	// 	t.Error("Error: Didn't find the subpath!")
	// }

	v, _, _ = writeCache.Read(9, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{})
	keys := v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		t.Error("Error: Keys don't match !")
	}

	// Delete the container-0
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", nil); err != nil {
		t.Error("Error: Cann't delete a path twice !")
	}

	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: The path should be gone already !")
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	out := univalue.Univalues{}.Decode(univalue.Univalues(transitions).Encode()).(univalue.Univalues)

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(out)

	committer.Precommit([]uint64{1})
	committer.Commit(10)

	if v, _, _ := writeCache.Read(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", &commutative.Path{}); v != nil {
		t.Error("Error: Should be gone already !")
	}
}

func TestCommitUint256(b *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	writeCache := sstore.WriteCache
	WriteNewAcountsToCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	basev := commutative.NewBoundedU256FromU64(0, 999)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), basev); err != nil {
		b.Error(err)
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(trans)
	committer.Precommit([]uint64{0})
	committer.Commit(10)

	writeCache = NewWriteCacheWithAcounts(store)
	delta := commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta); err != nil {
		b.Error(err)
	}

	trans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(trans)
	committer.Precommit([]uint64{0})
	committer.Commit(10)

	writeCache = NewWriteCacheWithAcounts(store)
	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(commutative.U256)); v == nil ||
		v.(uint256.Int) != *uint256.NewInt(11) {
		b.Error(err)
	}

	delta = commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta); err != nil {
		b.Error(err)
	}

	delta2 := commutative.NewU256DeltaFromU64(uint64(11), true)
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), delta2); err != nil {
		b.Error(err)
	}

	trans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(trans)
	committer.Precommit([]uint64{0})
	committer.Commit(10)

	writeCache = NewWriteCacheWithAcounts(store)
	if v, _, err := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+RandomKey(0), new(commutative.U256)); v == nil ||
		v.(uint256.Int) != *uint256.NewInt(33) {
		b.Error(err)
	}
}
