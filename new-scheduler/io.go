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
	"os"

	"github.com/arcology-network/common-lib/codec"
	mapi "github.com/arcology-network/common-lib/exp/map"
	slice "github.com/arcology-network/common-lib/exp/slice"
)

func LoadScheduler(filepath string) (*Scheduler, error) {
	buffer, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	sch := new(Scheduler)
	buffers := new(codec.Byteset).Decode(buffer).(codec.Byteset)
	contractBytes := new(codec.Byteset).Decode(buffers[0]).(codec.Byteset)
	sch.callees = slice.ParallelTransform(contractBytes, 4, func(i int, _ []byte) *Callee {
		return new(Callee).Decode(contractBytes[i])
	})

	contractKeys := new(codec.Strings).Decode(buffers[1]).(codec.Strings)
	contractIdx := new(codec.Uint32s).Decode(buffers[2]).(codec.Uint32s)

	sch.calleeDict = make(map[string]uint32)
	for i := range contractKeys {
		sch.calleeDict[contractKeys[i]] = contractIdx[i]
	}
	return sch, nil
}

func SaveScheduler(this *Scheduler, filepath string) error {
	buffer := slice.ParallelTransform(this.callees, 4, func(i int, _ *Callee) []byte {
		v, _ := this.callees[i].Encode()
		return v
	})

	keys, values := mapi.KVs(this.calleeDict)
	codec.Strings(keys).Encode()
	codec.Uint32s(values).Encode()

	data := codec.Byteset(
		[][]byte{
			codec.Byteset(buffer).Encode(),
			codec.Strings(keys).Encode(),
			codec.Uint32s(values).Encode(),
		}).Encode()
	return os.WriteFile(filepath, data, 0644)
}
