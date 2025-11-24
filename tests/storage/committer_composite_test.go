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
	"reflect"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	ethadaptor "github.com/arcology-network/eu/ethadaptor"

	// "github.com/arcology-network/common-lib/exp/softdeltaset"
	crdtcommon "github.com/arcology-network/common-lib/crdt/common"
	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	noncommutative "github.com/arcology-network/common-lib/crdt/noncommutative"
	statecell "github.com/arcology-network/common-lib/crdt/statecell"
	statestore "github.com/arcology-network/state-engine"
	stgcommon "github.com/arcology-network/state-engine/common"
	statecommitter "github.com/arcology-network/state-engine/state/committer"
	"github.com/arcology-network/state-engine/storage/proxy"
)

func TestAuxTrans(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	writeCache := sstore.StateCache

	alice := AliceAccount()
	if _, err := ethadaptor.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(statecell.Sorter)
	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))
	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.Commit(10) // Commit

	committer.SetStore(store)
	// create a path

	path := commutative.NewPath()

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err)
	}

	// Try to rewrite a path, should fail !
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err == nil {
		t.Error(err)
	}

	// Try to read an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1", new(commutative.Path)); value != nil {
		t.Error("Path shouldn't be not found")
	}

	// Try to read an nonexistent entry from an nonexistent path, should fail !
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-1/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	//try again
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.String)); value != nil {
		t.Error("Shouldn't be not found")
	}

	// Write the entry
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Shouldn't be not found")
	}

	// Read the entry back
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-000", new(noncommutative.Int64)); value == nil || value.(int64) != 1111 {
		t.Error("Shouldn't be not found")
	}

	// Read the path
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error(value)
	} else {
		keys := value.(*softdeltaset.DeltaSet[string]).Elements()
		if !reflect.DeepEqual(keys, []string{"elem-000"}) {
			t.Error("Wrong value ")
		}
	}

	transitions := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	delv, _ := transitions[0].Value().(crdtcommon.Type).Delta()
	if len(transitions) == 0 || !reflect.DeepEqual(delv.(*softdeltaset.DeltaSet[string]).Added().Elements(), []string{"elem-000"}) {
		t.Error("keys don't match")
	}

	value := transitions[1].Value()
	if *(value.(*noncommutative.Int64)) != 1111 {
		t.Error("keys don't match")
	}

	// wrong condition, value should still exists
	if value, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path)); value == nil {
		t.Error("The variable has been cleared")
	}

	in := statecell.StateCells(transitions).Encode()
	out := statecell.StateCells{}.Decode(in).(statecell.StateCells)

	committer = statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(out)

	committer.Precommit([]uint64{1})
	committer.Commit(10)
}

func TestCheckAccessRecords(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	alice := AliceAccount()
	if _, err := ethadaptor.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// _, trans00 := writeCache.Export(statecell.Sorter)
	trans00 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(trans00)
	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.Commit(10) // Commit

	// committer.SetStore(store)
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	// _, trans10 := writeCache.Export(statecell.Sorter)
	trans10 := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})

	committer = statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(trans10).Encode()).(statecell.StateCells))
	committer.Precommit([]uint64{1})
	committer.Commit(10) // Commit

	// committer.SetStore(store)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/1", noncommutative.NewInt64(1111)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/1") // create a path
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/2", noncommutative.NewInt64(2222)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/2") // create a path
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/3", noncommutative.NewInt64(3333)); err != nil {
		t.Error("Error: Failed to write blcc://eth1.0/account/alice/storage/ctrn-0/3") // create a path
	}

	v1, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", new(commutative.Path))
	if v1 == nil {
		t.Error("Error: Failed to read blcc://eth1.0/account/alice/storage/ctrn-0/") // create a path
	}

	keys := v1.(*softdeltaset.DeltaSet[string]).Elements()
	if len(keys) != 3 {
		t.Error("Error: There should be 3 elements only!!! actual = ", len(keys)) // create a path
	}
}
