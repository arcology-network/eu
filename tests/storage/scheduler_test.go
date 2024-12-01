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
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/eth"
	scheduler "github.com/arcology-network/scheduler"
	statestore "github.com/arcology-network/storage-committer"
	stgcommon "github.com/arcology-network/storage-committer/common"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

func TestSchedulerDeclaration(t *testing.T) {
	store := stgproxy.NewMemDBStoreProxy()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.WriteCache

	committer := stgcommitter.NewStateCommitter(sstore, sstore.GetWriters())

	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	bob := BobAccount()
	if _, err := eth.CreateNewAccount(stgcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})

	committer.Import(acctTrans).Precommit([]uint64{stgcommon.SYSTEM})
	committer.Commit(10)

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/"+stgcommon.PROPERTY_PATH+"1234/", commutative.NewPath()); err != nil {
		t.Error(err)
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/"+stgcommon.PROPERTY_PATH+"1234/"+stgcommon.DEFERRED_FUNC, noncommutative.NewBytes([]byte{255})); err != nil {
		t.Error(err)
	}

	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	propertyTrans := univalue.Univalues(acctTrans).To(univalue.RuntimeProperty{})
	propertyTrans.Print()

	sch, _ := scheduler.NewScheduler("", true)
	sch.Import(propertyTrans)
}
