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
	"fmt"

	mapi "github.com/arcology-network/common-lib/exp/map"
)

type Conflict struct {
	key     string
	self    uint32
	groupID []uint32
	txIDs   []uint32
	Err     error
}

func (this Conflict) ToPairs() [][2]uint32 {
	pairs := make([][2]uint32, 0, len(this.txIDs)*(len(this.txIDs)+1)/2-len(this.txIDs))
	for i := 0; i < len(this.txIDs); i++ {
		pairs = append(pairs, [2]uint32{this.self, this.txIDs[i]})
	}
	return pairs
}

type Conflicts []*Conflict

func (this Conflicts) ToDict() (map[uint32]uint64, map[uint32]uint64, [][2]uint32) {
	txDict := make(map[uint32]uint64)
	groupIDdict := make(map[uint32]uint64)
	for _, v := range this {
		for i := 0; i < len(v.txIDs); i++ {
			txDict[v.txIDs[i]] += 1
			groupIDdict[v.groupID[i]] += 1
		}
	}

	return txDict, groupIDdict, this.ToPairs()
}

func (this Conflicts) Keys() []string {
	keys := make([]string, 0, len(this))
	for _, v := range this {
		keys = append(keys, v.key)
	}
	return keys
}

func (this Conflicts) ToPairs() [][2]uint32 {
	dict := make(map[[2]uint32]int)
	for _, v := range this {
		pairs := v.ToPairs()
		for _, pair := range pairs {
			dict[pair]++
		}
	}
	return mapi.Keys(dict)
}

func (this Conflicts) Print() {
	for _, v := range this {
		fmt.Println(v.key, "      ", v.txIDs)
	}
}
