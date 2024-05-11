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

const (
	SHORT_CONTRACT_ADDRESS_LENGTH = 8
	FUNCTION_SIGNATURE_LENGTH     = 4
)

// The callee struct stores the information of a contract function that is called by the EOA initiated transactions.
// It is mainly used to optimize the execution of the transactions. A callee is uniquely identified by a
// combination of the contract's address and the function signature.
type Callee struct {
	Index          uint32   // Index of the Callee in the Callee list
	AddrAndSign    [12]byte // Short address of the callee, first 8 bytes from the function + signature [4]byte
	Indices        []uint32 // Indices of the conflicting callee indices.
	SequentialOnly bool     // A sequential only function
	SeqOnlyWith    [][12]byte
	paraOnlyWith   [][12]byte

	Calls      uint32 // Total number of calls
	AvgGas     uint32 // Average gas used
	Deferrable bool   // If one of the calls should be Deferrable to the second generation.
}

func NewCallee(idx uint32, addr []byte, funSign []byte) *Callee {
	return &Callee{
		Index:        idx,
		AddrAndSign:  new(codec.Bytes12).FromSlice(new(Callee).Compact(addr, funSign)),
		SeqOnlyWith:  [][12]byte{},
		paraOnlyWith: [][12]byte{},
		Deferrable:   false,
	}
}

func (this *Callee) Compact(addr []byte, funSign []byte) []byte {
	addr = slice.Clone(addr) // Make sure the original data is not modified
	return append(addr[:SHORT_CONTRACT_ADDRESS_LENGTH], funSign[:FUNCTION_SIGNATURE_LENGTH]...)
}

// 10x faster and 2x smaller than json marshal/unmarshal
func (this *Callee) Encode() ([]byte, error) {
	return codec.Byteset([][]byte{
		codec.Uint32(this.Index).Encode(),
		this.AddrAndSign[:],
		codec.Uint32s(this.Indices).Encode(),
		codec.Bool(this.SequentialOnly).Encode(),

		// SeqOnlyWith    [][12]byte
		// paraOnlyWith   [][4]byte

		codec.Bytes12s(this.SeqOnlyWith).Encode(),
		codec.Bytes12s(this.paraOnlyWith).Encode(),

		codec.Uint32(this.Calls).Encode(),
		codec.Uint32(this.AvgGas).Encode(),
		codec.Bool(this.Deferrable).Encode(),
	}).Encode(), nil
}

// new(codec.Bytes12).FromSlice(slice.Clone(fields[1])[:])
func (this *Callee) Decode(data []byte) *Callee {
	fields, _ := codec.Byteset{}.Decode(data).(codec.Byteset)
	this.Index = uint32(new(codec.Uint32).Decode(fields[0]).(codec.Uint32))
	this.AddrAndSign = new(codec.Bytes12).FromSlice(slice.Clone(fields[1])[:])
	this.Indices = new(codec.Uint32s).Decode(fields[2]).(codec.Uint32s)
	this.SequentialOnly = bool(new(codec.Bool).Decode(fields[3]).(codec.Bool))
	this.Calls = uint32(new(codec.Uint32).Decode(fields[4]).(codec.Uint32))
	this.AvgGas = uint32(new(codec.Uint32).Decode(fields[5]).(codec.Uint32))
	this.Deferrable = bool(new(codec.Bool).Decode(fields[6]).(codec.Bool))
	return this
}

func (this *Callee) Equal(other *Callee) bool {
	return this.Index == other.Index &&
		slice.EqualSet(this.AddrAndSign[:], other.AddrAndSign[:]) &&
		slice.EqualSet(this.Indices, other.Indices) &&
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

func (Callees) From(addr []byte, funSigns ...[]byte) [][12]byte {
	callees := make([][12]byte, len(funSigns))
	for i, funSign := range funSigns {
		callees[i] = new(codec.Bytes12).FromSlice(
			new(Callee).Compact(addr, funSign),
		)
	}
	return callees
}
