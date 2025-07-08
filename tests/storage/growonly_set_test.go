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
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

func TestGrowOnlySet(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlySet([]byte("0x1")))
	if err != nil {
		t.Error("Failed to write prepay function", err)
	}

	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", noncommutative.NewBytes([]byte("0x2")))
	if err != nil {
		t.Error("Failed to write prepay function", err)
	}

	// _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/prepayer", commutative.NewGrowOnlySet())
	// if err != nil {
	// 	t.Error("Failed to write prepay function", err)
	// }

	// if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/0x0000000000000000000000000000000000000000000000000000000000000000", new(noncommutative.String)); v != "124" {
	// 	t.Error("Error: Wrong return value")
	// }
}
