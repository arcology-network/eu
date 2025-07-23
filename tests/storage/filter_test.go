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
	"strings"
	"testing"

	deltaset "github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/orderedset"
	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/arcology-network/storage-committer/type/commutative"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
	"github.com/holiman/uint256"
)

/* Commutative Int64 Test */
func TestTransitionFilters(t *testing.T) {
	store := chooseDataStore()

	alice := RandomAccount()
	bob := RandomAccount()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	// writeCache = cache.NewWriteCache(store, 1, 1, platform.NewPlatform())

	eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache)
	// committer.NewAccount(stgcommcommon.SYSTEM, bob)

	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(univalue.Sorter)

	acctTrans := univalue.Univalues(slice.Clone(raw)).To(univalue.IPTransition{})

	if !acctTrans[1].Value().(*commutative.U256).Equal(raw[1].Value()) {
		t.Error("Error: Non-path commutative should have the values!!")
	}

	acctTrans[0].Value().(*commutative.Path).SetSubPaths([]string{"k0", "k1"})
	acctTrans[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	acctTrans[0].Value().(*commutative.Path).InsertRemoved([]string{"789", "116"})

	acctTrans[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	acctTrans[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999), true)
	acctTrans[1].Value().(*commutative.U256).SetLimits(*uint256.NewInt(1), *uint256.NewInt(2222222))

	deltav, _ := raw[0].Value().(*commutative.Path).Delta()
	if v := deltav.(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Added().Elements(), []string{}) {
		t.Error("Error: Value altered")
	}

	if v := deltav.(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Removed().Elements(), []string{}) {
		t.Error("Error: Delta altered")
	}

	deltav, _ = raw[1].Value().(*commutative.U256).Delta()
	if v := deltav.(uint256.Int); !v.Eq(uint256.NewInt(0)) {
		t.Error("Error: Value altered")
	}

	if v := deltav.(uint256.Int); !v.Eq(uint256.NewInt(0)) {
		t.Error("Error: Delta altered")
	}

	minv, maxv := raw[1].Value().(*commutative.U256).Limits()
	if v := minv.(uint256.Int); !v.Eq(&commutative.U256_MIN) {
		t.Error("Error: Min Value altered")
	}

	if v := maxv.(uint256.Int); !v.Eq(&commutative.U256_MAX) {
		t.Error("Error: Max altered")
	}

	copied := univalue.Univalues(slice.Clone(acctTrans)).To(univalue.IPTransition{})

	// Test Path
	v := copied[0].Value().(*commutative.Path).Value() // Committed
	if v.(*orderedset.OrderedSet[string]).Length() != 0 {
		t.Error("Error: A path commutative variable shouldn't have the initial value")
	}

	deltav, _ = copied[0].Value().(*commutative.Path).Delta() // Delta
	if v := deltav.(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Added().Elements(), []string{"123", "456"}) {
		t.Error("Error: Delta altered")
	}

	if v := deltav.(*deltaset.DeltaSet[string]); !reflect.DeepEqual(v.Removed().Elements(), []string{"789", "116"}) {
		t.Error("Error: Delta altered")
	}

	// Test Commutative 256
	if v := copied[1].Value().(*commutative.U256).Value().(uint256.Int); !(&v).Eq(uint256.NewInt(111)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	deltav, _ = copied[1].Value().(*commutative.U256).Delta() // Delta
	if v := deltav.(uint256.Int); !(&v).Eq(uint256.NewInt(999)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	minv, maxv = copied[1].Value().(*commutative.U256).Limits() // Min/Max
	if v := minv.(uint256.Int); !(&v).Eq(uint256.NewInt(1)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := maxv.(uint256.Int); !(&v).Eq(uint256.NewInt(2222222)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

}

func TestAccessFilters(t *testing.T) {
	store := chooseDataStore()

	alice := RandomAccount()
	bob := RandomAccount()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	eth.CreateDefaultPaths(stgcommcommon.SYSTEM, alice, writeCache)
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	raw := writeCache.Export(univalue.Sorter)

	raw[0].Value().(*commutative.Path).SetSubPaths([]string{"k0", "k1"})
	raw[0].Value().(*commutative.Path).SetAdded([]string{"123", "456"})
	raw[0].Value().(*commutative.Path).InsertRemoved([]string{"789", "116"})

	raw[1].Value().(*commutative.U256).SetValue(*uint256.NewInt(111))
	raw[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(999), true)
	raw[1].Value().(*commutative.U256).SetLimits(*uint256.NewInt(1), *uint256.NewInt(2222222))

	acctTrans := univalue.Univalues(slice.Clone(raw)).To(univalue.IPAccess{})

	if acctTrans[0].Value() != nil {
		t.Error("Error: Value altered")
	}

	// Test Commutative 256
	if v := acctTrans[1].Value().(*commutative.U256).Value().(uint256.Int); !(&v).Eq(uint256.NewInt(111)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	deltav, _ := acctTrans[1].Value().(*commutative.U256).Delta() // Delta
	if v := deltav.(uint256.Int); !(&v).Eq(uint256.NewInt(999)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	minv, maxv := acctTrans[1].Value().(*commutative.U256).Limits() // Min/Max
	if v := minv.(uint256.Int); !(&v).Eq(uint256.NewInt(1)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	if v := maxv.(uint256.Int); !(&v).Eq(uint256.NewInt(2222222)) {
		t.Error("Error: A non-path commutative variable should have the initial value")
	}

	idx, v := slice.FindFirstIf(acctTrans, func(_ int, v *univalue.Univalue) bool {
		// If balance is nil or nonce is nil, which shoudn't happen
		return (strings.Contains(*v.GetPath(), "/balance") && v.Value() == nil) || (strings.Contains(*v.GetPath(), "/nonce") && v.Value() == nil)
	})

	if idx != -1 {
		t.Error("Error: Nonce non-path commutative variables may keep their initial values", v)
	}
}
