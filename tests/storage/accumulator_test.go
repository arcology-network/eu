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
	"github.com/arcology-network/eu/eth"
	stgcommon "github.com/arcology-network/state-engine/common"
	commutative "github.com/arcology-network/state-engine/type/commutative"
	statecell "github.com/arcology-network/state-engine/type/statecell"

	arbitrator "github.com/arcology-network/scheduler/arbitrator"
	statestore "github.com/arcology-network/state-engine"
	"github.com/arcology-network/state-engine/storage/proxy"
	"github.com/holiman/uint256"
)

func TestAccumulatorUpperLimit(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	itc := statecell.ITTransition{}
	trans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(itc)
	transV := []*statecell.StateCell(trans)
	balanceDeltas := slice.CopyIf(transV, func(_ int, v *statecell.StateCell) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	// v := *uint256.NewInt(0)
	balanceDeltas[0].Value().(*commutative.U256).SetLimits(*uint256.NewInt(0), *uint256.NewInt(100))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta(*uint256.NewInt(11), true)

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))

	balanceDeltas[1].Value().(*commutative.U256).SetDelta(*uint256.NewInt(21), true)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta(*uint256.NewInt(5), true)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(63), true)

	// dict := make(map[string]*[]*statecell.StateCell)
	// dict[*(balanceDeltas[0]).GetPath()] = &balanceDeltas

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if (conflicts) != nil {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta(*uint256.NewInt(64), true)
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if (conflicts) == nil {
		t.Error("Error: There should be a of-limit-error")
	}
}

func TestAccumulatorLowerLimit(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	alice := AliceAccount()
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	trans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	transV := []*statecell.StateCell(trans)
	balanceDeltas := slice.CopyIf(transV, func(_ int, v *statecell.StateCell) bool { return strings.LastIndex(*v.GetPath(), "/balance") > 0 })

	balanceDeltas[0].SetTx(0)
	balanceDeltas[0].Value().(*commutative.U256).SetLimits((*uint256.NewInt(0)), (*uint256.NewInt(100)))
	balanceDeltas[0].Value().(*commutative.U256).SetDelta((*uint256.NewInt(11)), true)

	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))
	balanceDeltas = append(balanceDeltas, balanceDeltas[0].Clone().(*statecell.StateCell))

	balanceDeltas[1].SetTx(1)
	balanceDeltas[1].Value().(*commutative.U256).SetDelta((*uint256.NewInt(21)), true)

	balanceDeltas[2].SetTx(2)
	balanceDeltas[2].Value().(*commutative.U256).SetDelta((*uint256.NewInt(5)), true)

	balanceDeltas[3].SetTx(3)
	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(63)), true)

	conflicts := (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if (conflicts) != nil {
		t.Error("Error: There is no conflict")
	}

	balanceDeltas[3].Value().(*commutative.U256).SetDelta((*uint256.NewInt(64)), true)
	conflicts = (&arbitrator.Accumulator{}).CheckMinMax(balanceDeltas)
	if (conflicts) == nil {
		t.Error("Error: There should be a of-limit-error")
	}
}
