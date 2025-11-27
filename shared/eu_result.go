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

package shared

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/crdt/statecell"
)

type EuResult struct {
	H  string
	ID uint64
	// Transitions  []byte
	TransitTypes []byte
	// DC           *DeferredCall
	Trans   []*statecell.StateCell
	Status  uint64
	GasUsed uint64
}

func (this *EuResult) HeaderSize() uint64 {
	return 7 * codec.UINT64_LEN
}

func (this *EuResult) Size() uint64 {
	return this.HeaderSize() +
		uint64(len(this.H)) +
		codec.UINT64_LEN +
		uint64(statecell.StateCells(this.Trans).Size()) +
		codec.Bytes(this.TransitTypes).Size() +
		codec.Uint64(this.Status).Size() +
		codec.Uint64(this.GasUsed).Size()
}

func (this *EuResult) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeTo(buffer)
	return buffer
}

func (this *EuResult) EncodeTo(buffer []byte) int {
	if this == nil {
		return 0
	}

	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint64{
			codec.String(this.H).Size(),
			codec.Uint64(this.ID).Size(),
			statecell.StateCells(this.Trans).Size(),
			codec.Bytes(this.TransitTypes).Size(),
			codec.Uint64(this.Status).Size(),
			codec.Uint64(this.GasUsed).Size(),
		},
	)

	offset += codec.String(this.H).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.ID).EncodeTo(buffer[offset:])
	offset += codec.Bytes(statecell.StateCells(this.Trans).Encode()).EncodeTo(buffer[offset:])
	offset += codec.Bytes(this.TransitTypes).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.Status).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.GasUsed).EncodeTo(buffer[offset:])

	return offset
}

func (this *EuResult) Decode(buffer []byte) *EuResult {
	fields := [][]byte(codec.Byteset{}.Decode(buffer).(codec.Byteset))

	this.H = string(fields[0])
	this.ID = uint64(codec.Uint64(0).Decode(fields[1]).(codec.Uint64))
	this.Trans = statecell.StateCells(this.Trans).Decode(fields[2]).(statecell.StateCells)
	this.TransitTypes = []byte(codec.Bytes{}.Decode(fields[3]).(codec.Bytes))
	this.Status = uint64(codec.Uint64(0).Decode(fields[4]).(codec.Uint64))
	this.GasUsed = uint64(codec.Uint64(0).Decode(fields[5]).(codec.Uint64))
	return this
}

func (this *EuResult) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *EuResult) GobDecode(buffer []byte) error {
	this.Decode(buffer)
	return nil
}

func (tar *TxAccessRecords) GobEncode() ([]byte, error) {
	return tar.Encode(), nil
}

func (tar *TxAccessRecords) GobDecode(buffer []byte) error {
	tar.Decode(buffer)
	return nil
}

type Euresults []*EuResult

func (this *Euresults) HeaderSize() uint64 {
	return uint64(len(*this)+1) * codec.UINT64_LEN // Header length
}

func (this *Euresults) Size() uint64 {
	total := this.HeaderSize()
	for i := 0; i < len(*this); i++ {
		total += (*this)[i].Size()
	}
	return total
}

// Fill in the header info
func (this *Euresults) FillHeader(buffer []byte) {
	codec.Uint64(len(*this)).EncodeTo(buffer)

	offset := uint64(0)
	for i := 0; i < len(*this); i++ {
		codec.Uint64(offset).EncodeTo(buffer[codec.UINT64_LEN*uint64(i+1):])
		offset += (*this)[i].Size()
	}
}

func (this Euresults) GobEncode() ([]byte, error) {
	buffer := make([]byte, this.Size())
	this.FillHeader(buffer)

	offsets := make([]uint64, len(this)+1)
	offsets[0] = 0
	for i := 0; i < len(this); i++ {
		offsets[i+1] = offsets[i] + this[i].Size()
	}

	headerLen := this.HeaderSize()
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			this[i].EncodeTo(buffer[headerLen+offsets[i]:])
		}
	}
	common.ParallelWorker(len(this), 4, worker)
	return buffer, nil
}

func (this *Euresults) GobDecode(buffer []byte) error {
	bytesset := [][]byte(codec.Byteset{}.Decode(buffer).(codec.Byteset))
	euresults := make([]*EuResult, len(bytesset))
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			euresults[i] = (&EuResult{}).Decode(bytesset[i])
		}
	}
	common.ParallelWorker(len(bytesset), 4, worker)
	*this = euresults
	return nil
}
