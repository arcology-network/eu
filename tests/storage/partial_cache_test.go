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

// import (
// 	"reflect"
// 	"testing"

// 	storage "github.com/arcology-network/common-lib/storage"
// 	"github.com/arcology-network/common-lib/common"
// 	stgcomm "github.com/arcology-network/storage-committer"
// 	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
// 	importer "github.com/arcology-network/storage-committer/storage/committer"
// 	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
// 	storage "github.com/arcology-network/storage-committer/storage/proxy"
// )

// func TestPartialCache(t *testing.T) {
// 	memDB := storage.NewMemoryDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)
// 	store := storage.NewDataStore( policy, memDB, platform.Codec{}.Encode, platform.Codec{}.Decode)
// 		committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
//
// 	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
// committer.Commit(10)

// 	/* Filter persistent data source */
// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { // Do not access MemDB
// 		name := reflect.TypeOf(db).String()
// 		return name != "*storage.MemDB"
// 	}

// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues), true, excludeMemDB) // The changes will be discarded.
//
// 	committer.Precommit([]uint32{1})
// committer.Commit(10)

// 	// if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 	// 	t.Error("Error: The entry shouldn't be in the DB !")
// 	// } else {
// 	// 	if string(*(v.(*noncommutative.String))) != "1234" {
// 	// 		t.Error("Error: The entry shouldn't changed !")
// 	// 	}
// 	// }

// 	/* Don't filter persistent data source	*/
// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()                                 // Make sure only the persistent storage has the data.
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues)) // This should take effect
//
// 	committer.Precommit([]uint32{1})
// committer.Commit(10)

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry should have been changed !")
// 		}
// 	}
// }

// func TestPartialCacheWithFilter(t *testing.T) {
// 	memDB := storage.NewMemoryDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)

// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { /* Filter persistent data source */
// 		name := reflect.TypeOf(db).String()
// 		return name == "*storage.MemDB"
// 	}

// 	store := storage.NewDataStore( policy, memDB, platform.Codec{}.Encode, platform.Codec{}.Decode, excludeMemDB)
// 		committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.WriteCache()
// 	alice := AliceAccount()
// 	if _, err := adaptorcommon.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
//
// 	committer.Precommit([]uint32{stgcommcommon.SYSTEM})
// committer.Commit(10)

// 	if _, err := committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
// 		t.Error(err)
// 	}

// 	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

// 	// 	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.WriteCache()

// 	committer.WriteCache().Clear()

// 	// ccmap2 := committer.Importer().Store().(*storage.DataStore).Cache()
// 	// fmt.Print(ccmap2)
// 	out := univalue.Univalues{}.Decode(univalue.Univalues(slice.Clone(acctTrans)).Encode()).(univalue.Univalues)
// 	committer.Import(out, true, excludeMemDB) // The changes will be discarded.
//
// 	committer.Precommit([]uint32{1})
// committer.Commit(10)

// 	if v, _ := committer.Read(2, "blcc://eth1.0/account/"+alice+"/storage/1234"); v == nil {
// 		t.Error("Error: The entry shouldn't be in the DB as the persistent DB has been excluded !")
// 	} else {
// 		if v.(string) != "9999" {
// 			t.Error("Error: The entry shouldn't changed !")
// 		}
// 	}
// }
