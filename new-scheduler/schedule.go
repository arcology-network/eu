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
	"github.com/arcology-network/common-lib/exp/array"
	eucommon "github.com/arcology-network/eu/common"
)

type Schedule struct {
	Transfers    []*eucommon.StandardMessage // Transfers
	Deployments  []*eucommon.StandardMessage // Contract deployments
	Unknows      []*eucommon.StandardMessage // Messages with unknown conflicts with others
	WithConflict []*eucommon.StandardMessage // Messages with conflicts
	Sequentials  []*eucommon.StandardMessage // Callees that are marked as sequential only
	Generations  [][]*eucommon.StandardMessage
	CallSums     []map[string]int
}

// The function counts the number of each unique calls within each generation.
func (this *Schedule) GetCallSums() {
	this.CallSums = array.Append(this.Generations, func(i int, msgs []*eucommon.StandardMessage) map[string]int {
		dict := map[string]int{}
		array.Foreach(msgs, func(_ int, msg **eucommon.StandardMessage) { dict[ToKey(*msg)]++ })
		return dict
	})
}

// The function returns the number of instances of the callee.
func (this *Schedule) NumInstances(addr [20]byte, signature [4]byte, generation int) int {
	return 0
	// return this.CallSums[generation][ToKey(&eucommon.StandardMessage{})]
}
