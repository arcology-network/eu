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
	"math/rand"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	platform "github.com/arcology-network/common-lib/types/storage/platform"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	cache "github.com/arcology-network/common-lib/types/storage/writecache"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
	hexutil "github.com/ethereum/go-ethereum/common/hexutil"
	"golang.org/x/crypto/sha3"
	// "github.com/google/btree"
	// ehtrlp "github.com/elliotchance/orderedmap"
)

func TestWriteWithNewWriteCacheSlowWrite(b *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	NewAcountsInCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 5)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i)))
		// writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys))))
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(10)
	fmt.Println("Commit time:", time.Since(t0))

	writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	k := RandomKey(9999999)
	v := noncommutative.NewInt64(int64(9999999))
	t0 = time.Now()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+k, v); err != nil {
		b.Error(err)
	}
	fmt.Println("Second Write time:", 1, "keys in", time.Since(t0))

	keys = RandomKeys(11, 12)
	writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	t0 = time.Now()
	for i := 0; i < len(keys); i++ {
		writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i)))
		// writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys))))
	}
	fmt.Println("2nd Write time:", len(keys), "keys in", time.Since(t0))
}

func TestWriteWithNewWriteCache(b *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	NewAcountsInCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100000)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i)))
		writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys))))
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	fmt.Println("New committer + Import:", time.Since(t0))

	t0 = time.Now()
	committer.Precommit([]uint32{0})
	fmt.Println("Precommit time:", time.Since(t0))

	t0 = time.Now()
	committer.Commit(10)
	fmt.Println("Commit time:", time.Since(t0))

	t0 = time.Now()
	writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	keys = RandomKeys(len(keys), len(keys)+1)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		// if _, err := cache.NewWriteCache(store, 1, 1, platform.NewPlatform()).Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
		// 	b.Error(err)
		// }
	}
	fmt.Println("Second Write time:", len(keys), "keys in", time.Since(t0))
}

func BenchmarkWriteAfterLargeCommitUint64(b *testing.B) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	NewAcountsInCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	keys := RandomKeys(0, 100000)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i)))
		writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys))))
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(10)
	fmt.Println("Commit time:", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(len(keys), len(keys)+10)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(100000, 100000*2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(10)
	fmt.Println("2.Commit time:", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(200000, 200000*2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], noncommutative.NewInt64(int64(i))); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], noncommutative.NewInt64(int64(i+len(keys)))); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Third Write time:", len(keys)*2, "keys in", time.Since(t0))
}

func BenchmarkWriteAfterLargeCommitUint256(b *testing.B) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	NewAcountsInCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0-0/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	bob := BobAccount()
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/", commutative.NewPath()); err != nil {
		b.Error(err)
	}

	batchSize := 100

	v := noncommutative.NewBytes(big.NewInt(100).Bytes())
	keys := RandomKeys(0, batchSize)
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], v)
		writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], v)
	}
	fmt.Println("First Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(10)
	fmt.Println("Commit time:", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(len(keys), len(keys)+10)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(batchSize, batchSize*2)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Second Write time:", len(keys)*2, "keys in", time.Since(t0))

	t0 = time.Now()
	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{}))
	committer.Precommit([]uint32{0})
	committer.Commit(10)
	fmt.Println("2.Commit time:", time.Since(t0))

	t0 = time.Now()
	keys = RandomKeys(batchSize*2, batchSize*3)
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/alice-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}

		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/storage/container/ctrn-1/bob-elem-"+keys[i], v); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Third Write time:", len(keys)*2, "keys in", time.Since(t0))
}

func BenchmarkPathReadAndWrites(b *testing.B) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		fmt.Println(err)
	}

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/", path); err != nil {
		b.Error(err)
	}

	keys := make([]string, 2000)
	values := make([]interface{}, len(keys))

	// Generate random keys and values
	for i := 0; i < len(keys); i++ {
		buf := sha3.Sum256([]byte(fmt.Sprint(rand.Int())))
		keys[i] = hexutil.Encode(buf[:20])
		values[i] = commutative.NewUnboundedU256()
	}

	// Insert keys and values
	t0 := time.Now()
	for i := 0; i < len(keys); i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys[i], values[i]); err != nil {
			b.Error(err)
		}
	}
	fmt.Println("Inserted ", len(keys), "Keys in:", time.Since(t0))

	t0 = time.Now()
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters()).Import(writeCache.Export())
	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(10)
	fmt.Println("committer ", len(keys), "Keys in:", time.Since(t0))

	// Read keys and values
	t0 = time.Now()
	for i := 0; i < len(keys)-1; i++ {
		writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys[i], values[i])
	}
	fmt.Println("Read ", len(keys), "Keys in:", time.Since(t0))

	// // writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform()) // a new write cache
	// // t0 = time.Now()
	for i := 0; i < 1; i++ {
		if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/storage/container/ctrn-0/"+keys[i], values[i]); err != nil {
			b.Error(err)
		}
	}
	// fmt.Println("Rewrite ", len(keys), "Keys in:", time.Since(t0))

	// sets := make([]*orderedset.OrderedSet[string], 50)
	// t0 = time.Now()
	// for i := 0; i < len(sets); i++ {
	// 	sets[i] = orderedset.NewOrderedSet[string]("", 1)
	// }

	// t0 = time.Now()
	// slice.ParallelTransform(sets, 16, func(i int, _ *orderedset.OrderedSet[string]) *orderedset.OrderedSet[string] {
	// 	return orderedset.NewOrderedSet[string]("", 1000)
	// })
	// fmt.Println("NewOrderedSet:  ", len(sets), "Keys in:", time.Since(t0))
}
