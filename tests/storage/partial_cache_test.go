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
// 	stgcomm "github.com/arcology-network/state-engine"
// 	stgcommon "github.com/arcology-network/state-engine/common"
// 	importer "github.com/arcology-network/state-engine/state/committer"
// 	noncommutative "github.com/arcology-network/state-engine/type/noncommutative"
// 	storage "github.com/arcology-network/state-engine/storage/proxy"
// )

// func TestPartialCache(t *testing.T) {
// 	memDB := storage.NewMemoryDB()
// 	policy := storage.NewCachePolicy(10000000, 1.0)
// 	store := storage.NewDataStore( policy, memDB, stgtypecommonCodec{}.Encode, stgtypecommonCodec{}.Decode)
// 		committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.StateCache()
// 	alice := AliceAccount()
// 	if _, err := eucommon.CreateNewAccount(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
// 	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))
//
// 	committer.Precommit([]uint32{stgcommon.SYSTEM})
// committer.Commit(10)

// 	/* Filter persistent data source */
// 	excludeMemDB := func(db storage.PersistentStorageInterface) bool { // Do not access MemDB
// 		name := reflect.TypeOf(db).String()
// 		return name != "*storage.MemDB"
// 	}

// 	committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999"))
// 	acctTrans = statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
// 	committer.Importer().Store().(*storage.DataStore).Cache().Clear()
// 	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells), true, excludeMemDB) // The changes will be discarded.
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
// 	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells)) // This should take effect
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

// 	store := storage.NewDataStore( policy, memDB, stgtypecommonCodec{}.Encode, stgtypecommonCodec{}.Decode, excludeMemDB)
// 		committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.StateCache()
// 	alice := AliceAccount()
// 	if _, err := eucommon.CreateNewAccount(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	committer.Write(stgcommon.SYSTEM, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("1234"))
// 	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
// 	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))
//
// 	committer.Precommit([]uint32{stgcommon.SYSTEM})
// committer.Commit(10)

// 	if _, err := committer.Write(1, "blcc://eth1.0/account/"+alice+"/storage/1234", noncommutative.NewString("9999")); err != nil {
// 		t.Error(err)
// 	}

// 	acctTrans = statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

// 	// 	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
// writeCache := committer.StateCache()

// 	committer.StateCache().Clear()

// 	// ccmap2 := committer.Importer().Store().(*storage.DataStore).Cache()
// 	// fmt.Print(ccmap2)
// 	out := statecell.StateCells{}.Decode(statecell.StateCells(slice.Clone(acctTrans)).Encode()).(statecell.StateCells)
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
