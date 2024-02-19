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
	"github.com/arcology-network/common-lib/codec"
	eucommon "github.com/arcology-network/eu/common"
)

// The callee struct stores the information of a contract that is called by EOA initiated transactions.
// It is mainly used to optimize the execution of the transactions.
type Callee struct {
	Index          uint32   // Index of the contract in the contract list
	Address        [8]byte  // Short address
	Signature      [4]byte  // Function signature
	Indices        []uint32 // Indices of the conflicting callee indices.
	SequentialOnly bool     // A sequential only contract
	Calls          uint32   // Total number of calls
	AvgGas         uint32   // Average gas used
}

// 10x faster and 2x smaller than json marshal/unmarshal
func (this *Callee) Encode() ([]byte, error) {
	return codec.Byteset([][]byte{
		codec.Uint32(this.Index).Encode(),
		this.Address[:],
		this.Signature[:],
		codec.Uint32s(this.Indices).Encode(),
		codec.Bool(this.SequentialOnly).Encode(),
		codec.Uint32(this.Calls).Encode(),
		codec.Uint32(this.AvgGas).Encode(),
	}).Encode(), nil
}

func (this *Callee) Decode(data []byte) *Callee {
	fields, _ := codec.Byteset{}.Decode(data).(codec.Byteset)
	this.Index = uint32(new(codec.Uint32).Decode(fields[0]).(codec.Uint32))
	copy(this.Address[:], fields[1])
	copy(this.Signature[:], fields[2])
	this.Indices = new(codec.Uint32s).Decode(fields[3]).(codec.Uint32s)
	this.SequentialOnly = bool(new(codec.Bool).Decode(fields[4]).(codec.Bool))
	this.Calls = uint32(new(codec.Uint32).Decode(fields[5]).(codec.Uint32))
	this.AvgGas = uint32(new(codec.Uint32).Decode(fields[5]).(codec.Uint32))
	return this
}

func ToKey(msg *eucommon.StandardMessage) string {
	if (*msg.Native).To == nil {
		return ""
	}

	if len(msg.Native.Data) == 0 {
		return string((*msg.Native.To)[:ADDRESS_LENGTH])
	}
	return CallToKey((*msg.Native.To)[:], msg.Native.Data[:4])
}

func CallToKey(addr []byte, funSign []byte) string {
	return string(addr[:ADDRESS_LENGTH]) + string(funSign[:4])
}

type Callees []*Callee
