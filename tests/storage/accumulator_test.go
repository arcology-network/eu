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
	"strings"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	"github.com/arcology-network/eu/eth"

	arbitrator "github.com/arcology-network/scheduler/arbitrator"
	statestore "github.com/arcology-network/storage-committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/holiman/uint256"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()

	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	itc := univalue.ITTransition{}
	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(itc)
	transV := []*univalue.Univalue(trans)
	balanceDeltas := slice.CopyIf(transV, func(_ int, v *univalue.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	// v := *uint256.NewInt(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin(*uint256.NewInt(0))
	balanceDeltas[0].Value().(*commutative.U256).SetMax(*uint256.NewInt(100))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta(*uint256.NewInt(11))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(21))
	balanceDeltas[2].Value().(*commutative.U256).SetDelta(*uint256.NewInt(5))
	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(63))

	// dict := make(map[string]*[]*univalue.Univalue)
	// dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(64))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}

func TestAccumulatorLowerLimit(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	transV := []*univalue.Univalue(trans)
	balanceDeltas := slice.CopyIf(transV, func(_ int, v *univalue.Univalue) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetMin((*uint256.NewInt(0)))
	balanceDeltas[0].Value().(*commutative.U256).SetMax((*uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((*uint256.NewInt(11)))

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*univalue.Univalue))

	balanceDeltas[1].SetTx(1)
	balanceDeltas[1].Value().(*commutative.U256).SetDelta((*uint256.NewInt(21)))

	balanceDeltas[2].SetTx(2)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((*uint256.NewInt(5)))

	balanceDeltas[3].SetTx(3)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(63)))

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 0 {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(64)))
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if len(conflicts) != 1 {
		t.Error("Error: There is should be a of-limit-error")
	}
}
