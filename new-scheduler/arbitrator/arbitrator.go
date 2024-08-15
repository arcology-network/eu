/*
 *   Copyright (c) 2023 Arcology Network

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

package arbitrator

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	// github.com/arcology-network/common-lib/types/storage/common
)

type Arbitrator struct {
	groupIDs    []uint32
	transitions []*univalue.Univalue
}

func (this *Arbitrator) Insert(groupIDs []uint32, newTrans []*univalue.Univalue) int {
	this.transitions = append(this.transitions, newTrans...)
	this.groupIDs = append(this.groupIDs, groupIDs...)
	return len(this.groupIDs)
}

func (this *Arbitrator) Detect(groupIDs []uint32, newTrans []*univalue.Univalue) []*Conflict {
	if this.Insert(groupIDs, newTrans) == 0 {
		return []*Conflict{}
	}

	// t0 := time.Now()
	univalue.Univalues(newTrans).Sort(groupIDs)

	ranges := slice.FindAllIndics(newTrans, func(lhv, rhv *univalue.Univalue) bool {
		return *lhv.GetPath() == *rhv.GetPath()
	})

	conflicts := []*Conflict{}
	for i := 0; i < len(ranges)-1; i++ {
		if ranges[i]+1 == ranges[i+1] {
			continue // Only one entry
		}

		offset := int(1)
		if newTrans[ranges[i]].Writes() == 0 { // Whyen write == 0, there are only two possibilities: 1. delta write only; 2. read only
			subTrans := newTrans[ranges[i]+1 : ranges[i+1]]

			if newTrans[ranges[i]].IsReadOnly() || newTrans[ranges[i]].IsDeltaWriteOnly() { // Read delta write
				if newTrans[ranges[i]].IsReadOnly() { // Read only
					offset, _ = slice.FindFirstIf(subTrans, func(_ int, v *univalue.Univalue) bool { return !v.IsReadOnly() })
				}

				if newTrans[ranges[i]].IsDeltaWriteOnly() { // Delta write only
					offset, _ = slice.FindFirstIf(subTrans, func(_ int, v *univalue.Univalue) bool { return !v.IsDeltaWriteOnly() })
				}
				offset = common.IfThen(offset < 0, ranges[i+1]-ranges[i], offset+1) // offset == -1 means no conflict found
			}
		}

		if ranges[i]+offset == ranges[i+1] {
			continue
		}

		conflictTxs := []uint32{}
		slice.Foreach(newTrans[ranges[i]+offset:ranges[i+1]], func(_ int, v **univalue.Univalue) {
			conflictTxs = append(conflictTxs, (*v).GetTx())
		})

		conflicts = append(conflicts,
			&Conflict{
				key:     *newTrans[ranges[i]].GetPath(),
				self:    newTrans[ranges[i]].GetTx(),
				groupID: groupIDs[ranges[i]+offset : ranges[i+1]],
				txIDs:   conflictTxs,
				Err:     errors.New(stgcommcommon.WARN_ACCESS_CONFLICT),
			},
		)

		if len(conflicts) > 0 {
			if newTrans[ranges[i]].Writes() == 0 {
				if newTrans[ranges[i]].IsDeltaWriteOnly() { // Delta write only
					offset, _ = slice.FindFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
						func(_ int, v *univalue.Univalue) bool {
							return !v.IsDeltaWriteOnly()
						})
				} else { // Read only
					offset, _ = slice.FindFirstIf(newTrans[ranges[i]+1:ranges[i+1]],
						func(_ int, v *univalue.Univalue) bool {
							return v.Writes() > 0 || v.DeltaWrites() > 0
						})
				}
				offset = common.IfThen(offset < 0, ranges[i+1]-ranges[i], offset+1) // offset == -1 means no conflict found
			}
		}

		dict := common.MapFromSlice(conflictTxs, true) //Conflict dict
		trans := slice.CopyIf(newTrans[ranges[i]+offset:ranges[i+1]], func(_ int, v *univalue.Univalue) bool { return (*dict)[v.GetTx()] })

		if outOfLimits := (&Accumulator{}).CheckMinMax(trans); outOfLimits != nil {
			conflicts = append(conflicts, outOfLimits...)
		}
	}

	// if len(conflicts) > 0 {
	// 	fmt.Println("range: ", ranges)
	// }
	return conflicts
}
