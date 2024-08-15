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
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/common-lib/types"
)

type Schedule struct {
	Transfers    []*eucommon.StandardMessage // Transfers
	Deployments  []*eucommon.StandardMessage // Contract deployments
	Unknows      []*eucommon.StandardMessage // Messages with unknown conflicts with others
	WithConflict []*eucommon.StandardMessage // Messages with some known conflicts
	Sequentials  []*eucommon.StandardMessage // Callees that are marked as sequential only
	Generations  [][]*eucommon.StandardMessage
	CallCounts   []map[string]int
}

// The function outputs the optimized schedule. The shedule is a 3 dimensional array.
// The first dimension is the generation number. The second dimension is a set of
// parallel transaction arrays. These arrays are the transactions that can be executed in parallel.
// The third dimension is the transactions in the sequntial order.
func (this *Schedule) Optimize() [][][]*eucommon.StandardMessage {
	sch := [][][]*eucommon.StandardMessage{{ // Transfers and deployments will be executed first
		append(this.Transfers, this.Deployments...),
		append(this.WithConflict, this.Sequentials...),
	}}

	sch = append(sch, slice.Transform(this.Unknows, func(_ int, msg *eucommon.StandardMessage) []*eucommon.StandardMessage {
		return []*eucommon.StandardMessage{msg}
	}))

	for i := 0; i < len(this.Generations); i++ {
		if i == 0 {
			this.Generations[i] = append(this.Generations[i], this.Unknows...)
		}

		sch = append(sch, slice.Transform(this.Generations[i], func(_ int, msg *eucommon.StandardMessage) []*eucommon.StandardMessage {
			return []*eucommon.StandardMessage{msg}
		}))
	}

	slice.RemoveIf(&sch, func(i int, gen [][]*eucommon.StandardMessage) bool {
		slice.RemoveIf(&gen, func(_ int, msgs []*eucommon.StandardMessage) bool {
			return len(msgs) == 0
		})
		return len(gen) == 0
	})
	return sch
}
