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
	"math"
	"sort"

	mapi "github.com/arcology-network/common-lib/exp/map"
	"github.com/arcology-network/common-lib/exp/product"
	slice "github.com/arcology-network/common-lib/exp/slice"
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
}

// The function will find the index of the entry by its address and signature.
// If the entry is found, the index will be returned. If the entry is not found, the index will be added to the scheduler.
// If the entry is new
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
	// Get the static schedule for the given transactions first.
	sch, msgPairs := this.Prefilter(stdMsgs)
	if len(msgPairs) == 0 {
		return sch
	}

	// Sort the callees by the number of conflicts and the callee index in ascending order.
	// Need to use pairs not msgPairs
	sort.Slice(msgPairs, func(i, j int) bool {
		lft, rgt := len(this.callees[(msgPairs)[i].First].Indices), len(this.callees[(msgPairs)[j].First].Indices)
		if lft < rgt {
			return lft < rgt
		}
		return (msgPairs)[i].Second.ID < (msgPairs)[j].Second.ID
	})

	// The code below will search for the parallel transaction set from a set of conflicting transactions.
	// Whataever left is the sequential transaction set after this.
	for {
		// The conflict dictionary of all indices of the current transaction set.
		calleeDict := map[uint32]*product.Pair[uint32, *eucommon.StandardMessage]{}
		calleeDict[(msgPairs)[0].First] = (msgPairs)[0] // Start with the first callee.

		// The conflict dictionary of all the known conflict indices of the current transaction set.
		conflictDict := mapi.FromSlice(this.callees[(msgPairs)[0].First].Indices, func(k uint32) bool { return true })

		// The msg to include in the parallel transaction set must not have any conflicts with the other callees in the set.
		paraMsgs := product.Pairs[uint32, *eucommon.StandardMessage]{(msgPairs)[0]}
		for i, msgToInclude := range msgPairs {
			if calleeDict[msgToInclude.First] != nil {
				continue
			}

			// The current callee isn't in the conflict idx set or other callees and vice versa.
			if !conflictDict[msgToInclude.First] && !mapi.ContainsAny(calleeDict, this.callees[msgToInclude.First].Indices) {

				// Add the new callee's conflicts to the conflict dictionary.
				mapi.Insert(conflictDict, this.callees[msgToInclude.First].Indices, func(_ int, k uint32) (uint32, bool) {
					return k, true
				})

				calleeDict[msgToInclude.First] = msgToInclude // Add the current callee to the set.
				paraMsgs = append(paraMsgs, msgToInclude)     // Add the current callee to the parallel transaction set.
				(msgPairs)[i] = nil                           // Remove the current callee.
			}
		}

		// If it only contains one initial transaction, then there is no need to continue.
		if len(paraMsgs) == 1 {
			break
		}

		// Look for the deferred transactions and add them to the deferred transaction set.
		deferred := this.Deferred(&paraMsgs)
		sch.Generations = append(sch.Generations, paraMsgs.Seconds()) // Insert the parallel transaction first
		sch.Generations = append(sch.Generations, deferred.Seconds()) // Insert the deferred transaction set to the ne

		// Remove the first transaction from the msgPairs slice. since it is already in the parallel transaction set.
		(msgPairs)[0] = nil
		if len(slice.Remove(&msgPairs, nil)) == 0 {
			break // Nothing left to process.
		}
	}

	// Deferred array can be empty, so remove it if it is.
	slice.RemoveIf(&sch.Generations, func(i int, v []*eucommon.StandardMessage) bool {
		return len(v) == 0
	})

	// Whatever is left in the msgPairs array is the sequential transaction set.
	sch.WithConflict = (*product.Pairs[uint32, *eucommon.StandardMessage])(&msgPairs).Seconds()
	return sch
}

// The scheduler will scan through and look for multipl instances of the same callee and put one of them in the second
// consecutive set of transactions for deferred execution.
func (this *Scheduler) Deferred(paraMsgInfo *product.Pairs[uint32, *eucommon.StandardMessage]) product.Pairs[uint32, *eucommon.StandardMessage] {
	sort.Slice(*paraMsgInfo, func(i, j int) bool {
		if (*paraMsgInfo)[i].First != (*paraMsgInfo)[j].First {
			return (*paraMsgInfo)[i].First < (*paraMsgInfo)[j].First
		}
		return (*paraMsgInfo)[i].Second.ID < (*paraMsgInfo)[j].Second.ID
	})

	deferredMsgs := product.Pairs[uint32, *eucommon.StandardMessage]{}
	for i := 0; i < len(*paraMsgInfo); i++ {
		// Find the first and last instance of the same callee.
		first, _ := slice.FindFirstIf(*paraMsgInfo, func(v *product.Pair[uint32, *eucommon.StandardMessage]) bool {
			return (*paraMsgInfo)[i].First == v.First
		})

		// Find the first and last instance of the same callee.
		last, deferred := slice.FindLastIf(*paraMsgInfo, func(v *product.Pair[uint32, *eucommon.StandardMessage]) bool {
			return (*paraMsgInfo)[i].First == v.First
		})

		// If the first and last instance of the same callee are different, then
		// more than one instance of the same callee is there.
		if first != last {
			deferredMsgs = append(deferredMsgs, *deferred)
			slice.RemoveAt(paraMsgInfo.Array(), last)
		}
	}
	return deferredMsgs
}

// The scheduler will optimize the given transactions and look for the ones of specific types and return a schedule.
func (this *Scheduler) Prefilter(stdMsgs []*eucommon.StandardMessage) (*Schedule, []*product.Pair[uint32, *eucommon.StandardMessage]) {
	sch := &Schedule{}

	// Transfers won't have any conflicts, as long as they have enough balances. Deployments are less likely to have conflicts, but it's not guaranteed.
	sch.Transfers = slice.MoveIf(&stdMsgs, func(i int, msg *eucommon.StandardMessage) bool { return len(msg.Native.Data) == 0 })
	sch.Deployments = slice.MoveIf(&stdMsgs, func(i int, msg *eucommon.StandardMessage) bool { return msg.Native.To == nil })

	if len(stdMsgs) == 0 {
		return sch, []*product.Pair[uint32, *eucommon.StandardMessage]{} // All the transactions are transfers.
	}

	// Get the IDs for the given addresses and signatures, which will be used to find the callee index.
	pairs := slice.ParallelAppend(stdMsgs, 8, func(i int, msg *eucommon.StandardMessage) *product.Pair[uint32, *eucommon.StandardMessage] {
		idx, ok := this.calleeLookup[string(append((*msg.Native.To)[:ADDRESS_LENGTH], msg.Native.Data[:4]...))]
		if !ok {
			idx = math.MaxUint32 // The callee is new.
		}
		return &product.Pair[uint32, *eucommon.StandardMessage]{First: idx, Second: stdMsgs[i]}
	})

	msgPairs := (*product.Pairs[uint32, *eucommon.StandardMessage])(&pairs)
	if len(*msgPairs) == 0 {
		return sch, *msgPairs.Array()
	}

	// Move the transactions that have no known conflicts to the parallel trasaction array first.
	// If a callee has no known conflicts with anyone else, it is either a conflict-free implementation or has been fortunate enough to avoid conflicts so far.
	unknows := slice.MoveIf(msgPairs.Array(), func(_ int, v *product.Pair[uint32, *eucommon.StandardMessage]) bool {
		return v.First == math.MaxUint32
	})

	// Deployments are less likely to have conflicts, but it's not guaranteed.
	sequentialOnly := slice.MoveIf(msgPairs.Array(), func(_ int, v *product.Pair[uint32, *eucommon.StandardMessage]) bool {
		return this.callees[v.First].SequentialOnly
	})

	sch.Unknows = (*product.Pairs[uint32, *eucommon.StandardMessage])(&unknows).Seconds()
	sch.Sequentials = (*product.Pairs[uint32, *eucommon.StandardMessage])(&sequentialOnly).Seconds()
	return sch, *msgPairs.Array()
}
