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

package scheduler

import (
	"sort"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/matrix"
	eucommon "github.com/arcology-network/eu/common"
)

const (
	ADDRESS_LENGTH     = 8
	ID_LENGTH          = ADDRESS_LENGTH + 4 // 8 bytes for address and 4 bytes for signature
	MAX_CONFLICT_RATIO = 0.5
	MAX_NUM_CONFLICTS  = 256
)

type Scheduler struct {
	fildb        string
	calleeLookup map[string]uint32 // A calleeLookup table to find the index of a calleeLookup by its address + signature.
	callees      []*Callee
	bitmat       *matrix.BitMatrix
}

func (this *Scheduler) Find(addr [20]byte, sig [4]byte) (uint32, bool) {
	lftKey := string(append(addr[:ADDRESS_LENGTH], sig[:]...))
	idx, ok := this.calleeLookup[lftKey]
	if !ok {
		idx = uint32(len(this.callees))
		this.callees = append(this.callees, &Callee{
			Index:     uint32(len(this.callees)),
			Address:   [8]byte(addr[:ADDRESS_LENGTH]),
			Signature: sig,
		})
		this.calleeLookup[lftKey] = idx
	}
	return idx, ok
}

// Add a conflict pair to the scheduler
func (this *Scheduler) Add(lftAddr [20]byte, lftSig [4]byte, rgtAddr [20]byte, rgtSig [4]byte) bool {
	lftIdx, lftExist := this.Find(lftAddr, lftSig)
	rgtIdx, rgtExist := this.Find(rgtAddr, rgtSig)

	if lftExist && rgtExist {
		return false // The conflict pair is already recorded.
	}

	this.callees[lftIdx].Indices = append(this.callees[lftIdx].Indices, rgtIdx)
	this.callees[rgtIdx].Indices = append(this.callees[rgtIdx].Indices, lftIdx)

	return true
}

// The scheduler will optimize the given transactions and return a schedule.
// The schedule will contain the transactions that can be executed in parallel and the ones that have to
// be executed sequentially.
func (this *Scheduler) New(stdMsgs []*eucommon.StandardMessage) *Schedule {
	// Generate the keys for the given addresses and signatures.
	msgPairs := array.ParallelAppend(stdMsgs, 4, func(i int, msg *eucommon.StandardMessage) *common.Pair[uint32, *eucommon.StandardMessage] {
		if idx, ok := this.calleeLookup[string(append((*msg.Native.To)[:ADDRESS_LENGTH], msg.Native.Data[:4]...))]; ok {
			return &common.Pair[uint32, *eucommon.StandardMessage]{idx, stdMsgs[i]}
		}
		return nil
	})

	sch := this.ScheduleStatic(&msgPairs)

	sort.Slice(msgPairs, func(i, j int) bool {
		if len(this.callees[msgPairs[i].First].Indices) != len(this.callees[msgPairs[j].First].Indices) {
			return len(this.callees[msgPairs[i].First].Indices) < len(this.callees[msgPairs[j].First].Indices)
		}
		return msgPairs[i].Second.ID < msgPairs[j].Second.ID
	})

	minIdx, _ := array.Min(msgPairs, func(lft, rgt *common.Pair[uint32, *eucommon.StandardMessage]) bool {
		return len(this.callees[lft.First].Indices) < len(this.callees[rgt.First].Indices)
	})

	// The conflict dictionary of all indices of the current transaction set.
	calleeDict := mapi.FromArrayBy(msgPairs, func(_ int, v *common.Pair[uint32, *eucommon.StandardMessage]) (uint32, bool) {
		return v.First, true
	})

	// The conflict dictionary of all the known conflict indices of the current transaction set.
	conflictDict := mapi.FromArray(this.callees[minIdx].Indices, func(k uint32) bool { return true })

	// The msg to include in the parallel transaction set must not have any conflicts with the current callee set.
	for i, msgToInclude := range msgPairs {
		if i == minIdx &&
			!conflictDict[msgToInclude.First] && // The current callee isn't in the conflict idx set or other callees.
			!mapi.ContainsAny(calleeDict, this.callees[msgToInclude.First].Indices) { // None of the callees is in the conflict idx set of the current callee.
			mapi.Insert(conflictDict, this.callees[msgToInclude.First].Indices, func(_ int, k uint32) (uint32, bool) { return k, true })
			calleeDict[msgToInclude.First] = true
		}
	}

	// // The conflict dictionary of all the known conflict indices of the current transaction set.
	// dict := make(map[uint32]*int)
	// for i := range msgPairs {
	// 	for _, idx := range this.callees[msgPairs[i].First].Indices {
	// 		dict[idx] = new(int) // idx is the original index of the callee in the callee list.
	// 	}
	// }

	// // The callee has conflicts but none of these known conflicts is in the current callee set.
	// paraTrans := array.MoveIf(&msgPairs, func(i int, _ *common.Pair[uint32, *eucommon.StandardMessage]) bool {
	// 	_, ok := dict[msgPairs[i].First]
	// 	return !ok
	// })

	// // Remap the indices to a new set of contiguous indices for the label matrix.
	// i := 0
	// for _, v := range dict {
	// 	*v = i
	// 	i++
	// }

	// // Create a label matrix for all the known conflicts of the current transaction set.
	// labelMat := matrix.NewBitMatrix(len(msgPairs), len(msgPairs), false)
	// for i, info := range msgPairs {
	// 	row := dict[this.callees[info.First].Indices[i]]
	// 	for j := 0; j < len(this.callees[info.First].Indices); j++ {
	// 		col := dict[this.callees[info.First].Indices[j]]
	// 		labelMat.Set(*col, *row, true)
	// 	}
	// }

	// sums := make([]int, labelMat.Width())
	// for i := range labelMat.Width() {
	// 	sums[i] = labelMat.CountInRow(i, true)
	// }

	// withConflict := []*eucommon.StandardMessage{}

	// // The transactions that have conflicts with other transactions that should be executed sequentially.
	// // Whatever left are the transactions are conflict free.
	// withConflict = array.MoveIf(&msgPairs, func(i int, _ *common.Pair[uint32, *eucommon.StandardMessage]) bool {
	// 	labelMat.FillCol(i, false) // clear the row of the label matrix.
	// 	return sums[i] > 0
	// })

	// sch.Generations = append(sch.Generations, array.PairSeconds[uint32, *eucommon.StandardMessage](paraTrans))
	// sch.WithConflict = array.PairSeconds[uint32, *eucommon.StandardMessage](withConflict)

	return sch
}

func (this *Scheduler) ScheduleStatic(msgInfo *[]*common.Pair[uint32, *eucommon.StandardMessage]) *Schedule {
	// Transfers won't have any conflicts, as long as they have enough balances.
	transfers := array.MoveIf(msgInfo, func(i int, msg *common.Pair[uint32, *eucommon.StandardMessage]) bool {
		return len(msg.Second.Native.Data) == 0
	})

	// Deployments are less likely to have conflicts, but it's not guaranteed.
	deployments := array.MoveIf(msgInfo, func(i int, msg *common.Pair[uint32, *eucommon.StandardMessage]) bool {
		return len(msg.Second.Native.To) == 0
	})

	// Move the transactions that have no known conflicts to the parallel trasaction array first.
	// If a callee has no known conflicts with anyone else, it is either a conflict-free
	// implementation or has been fortunate enough to avoid conflicts so far.
	unknows := array.MoveIf(msgInfo, func(_ int, v *common.Pair[uint32, *eucommon.StandardMessage]) bool {
		return v == nil
	})

	// Deployments are less likely to have conflicts, but it's not guaranteed.
	sequentialOnly := array.MoveIf(msgInfo, func(_ int, v *common.Pair[uint32, *eucommon.StandardMessage]) bool {
		return this.callees[v.First].SequentialOnly
	})

	return &Schedule{
		Transfers:   array.PairSeconds[uint32, *eucommon.StandardMessage](transfers),
		Deployments: array.PairSeconds[uint32, *eucommon.StandardMessage](deployments),
		Unknows:     array.PairSeconds[uint32, *eucommon.StandardMessage](unknows),
		Sequentials: array.PairSeconds[uint32, *eucommon.StandardMessage](sequentialOnly),
	}
}
