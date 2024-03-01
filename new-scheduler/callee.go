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
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"
)

// The callee struct stores the information of a contract that is called by EOA initiated transactions.
// It is mainly used to optimize the execution of the transactions. A callee is uniquely identified by a
// combination of the contract's address and the function signature.
type Callee struct {
	Index          uint32   // Index of the Callee in the Callee list
	AddrAndSign    []byte   // Short address, the first 8 bytes of the contract address + Function signature [4]byte
	Indices        []uint32 // Indices of the conflicting callee indices.
	SequentialOnly bool     // A sequential only function
	Calls          uint32   // Total number of calls
	AvgGas         uint32   // Average gas used
	Deferrable     bool     // If one of the calls should be Deferrable to the second generation.
}

// 10x faster and 2x smaller than json marshal/unmarshal
func (this *Callee) Encode() ([]byte, error) {
	return codec.Byteset([][]byte{
		codec.Uint32(this.Index).Encode(),
		this.AddrAndSign[:],
		codec.Uint32s(this.Indices).Encode(),
		codec.Bool(this.SequentialOnly).Encode(),
		codec.Uint32(this.Calls).Encode(),
		codec.Uint32(this.AvgGas).Encode(),
		codec.Bool(this.Deferrable).Encode(),
	}).Encode(), nil
}

func (this *Callee) Decode(data []byte) *Callee {
	fields, _ := codec.Byteset{}.Decode(data).(codec.Byteset)
	this.Index = uint32(new(codec.Uint32).Decode(fields[0]).(codec.Uint32))
	this.AddrAndSign = slice.Clone(fields[1])
	this.Indices = new(codec.Uint32s).Decode(fields[2]).(codec.Uint32s)
	this.SequentialOnly = bool(new(codec.Bool).Decode(fields[3]).(codec.Bool))
	this.Calls = uint32(new(codec.Uint32).Decode(fields[4]).(codec.Uint32))
	this.AvgGas = uint32(new(codec.Uint32).Decode(fields[5]).(codec.Uint32))
	this.Deferrable = bool(new(codec.Bool).Decode(fields[6]).(codec.Bool))
	return this
}

func (this *Callee) Equal(other *Callee) bool {
	return this.Index == other.Index &&
		slice.Equal(this.AddrAndSign, other.AddrAndSign) &&
		slice.Equal(this.Indices, other.Indices) &&
		this.SequentialOnly == other.SequentialOnly &&
		this.Calls == other.Calls &&
		this.AvgGas == other.AvgGas &&
		this.Deferrable == other.Deferrable
}

// Get the callee key from a message
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
