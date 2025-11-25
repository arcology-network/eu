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
	"testing"
	"time"

	crdtcommon "github.com/arcology-network/common-lib/crdt/common"
	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	noncommutative "github.com/arcology-network/common-lib/crdt/noncommutative"
	statecell "github.com/arcology-network/common-lib/crdt/statecell"
	ethadaptor "github.com/arcology-network/eu/ethadaptor"
	statestore "github.com/arcology-network/state-engine"
	stgcommon "github.com/arcology-network/state-engine/common"
	cache "github.com/arcology-network/state-engine/storage/livecache"
	"github.com/arcology-network/state-engine/storage/proxy"
	"github.com/holiman/uint256"
)

func TestLiveCache(t *testing.T) {
	alice := AliceAccount()

	meta := commutative.NewPath()
	meta.(*commutative.Path).SetSubPaths([]string{"e-01", "e-001", "e-002", "e-002"})
	meta.(*commutative.Path).SetAdded([]string{"+01", "+001", "+002", "+002"})
	meta.(*commutative.Path).InsertRemoved([]string{"-091", "-0092", "-092", "-092", "-097"})
	metaIn := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", 3, 4, 11, meta, nil)
	fmt.Println("Meta size: ", metaIn.Value().(crdtcommon.Type).MemSize())

	u256 := commutative.NewBoundedU256(uint256.NewInt(0), uint256.NewInt(100))
	u256In := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256, nil)
	fmt.Println("u256 size: ", u256In.Value().(crdtcommon.Type).MemSize())

	bytes := noncommutative.NewBytes([]byte{1, 2, 3, 4, 5})
	bytesIn := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/bytes", 3, 4, 5, bytes, nil)
	fmt.Println("bytesIn size: ", bytesIn.Value().(crdtcommon.Type).MemSize())

	// u256In := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", 3, 4, 0, u256, nil)
	// fmt.Println("u256 size: ", u256In.Value().(crdtcommon.Type).MemSize())

	// outV, _, _ := statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
	liveCache := cache.NewLiveCache(151)
	liveCache.DebugCommit([]*statecell.StateCell{metaIn, u256In, bytesIn}, 1)
	liveCache.Print()

	bytes = noncommutative.NewBytes([]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	bytesIn = statecell.NewStateCell(1, "blcc://eth1.0/account/"+alice+"/storage/bytes", 3, 4, 5, bytes, nil)
	fmt.Println("bytesIn size: ", bytesIn.Value().(crdtcommon.Type).MemSize())

	liveCache.DebugCommit([]*statecell.StateCell{bytesIn}, 2)
	liveCache.Print()
}

func BenchmarkLiveCache(b *testing.B) {
	store := chooseDataStore()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache
	alice := AliceAccount()
	if _, err := ethadaptor.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	for i := 0; i < 1000; i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0"+fmt.Sprint(i), noncommutative.NewString(string("test"))); err != nil { /* The first Element */
			b.Error(err)
			break
		}
	}

	trans := writeCache.Export()

	liveCache := cache.NewLiveCache(1280)
	t0 := time.Now()
	liveCache.DebugCommit(trans, 1)
	fmt.Println("liveCache Commit time: ", liveCache.Length(), liveCache.Size(), time.Since(t0))

	// if liveCache.Length() != 11 || liveCache.Size() != 194 {
	// 	b.Error("Length is not 11 or the size is not 194")
	// }

	// fmt.Println("Count, Size in Bytes:", liveCache.Length(), liveCache.Size())
}
