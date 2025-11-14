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
	"github.com/arcology-network/storage-committer/type/univalue"
)

type TxAccessRecords struct {
	Hash     string
	ID       uint64
	Accesses []*univalue.Univalue
}

func (this *TxAccessRecords) HeaderSize() uint64 {
	return 3 * codec.UINT64_LEN
}

func (this *TxAccessRecords) Size() uint64 {
	return this.HeaderSize() +
		codec.String(this.Hash).Size() +
		codec.UINT64_LEN +
		univalue.Univalues(this.Accesses).Size()
}

func (this *TxAccessRecords) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeTo(buffer)
	return buffer
}

func (this *TxAccessRecords) EncodeTo(buffer []byte) int {
	if this == nil {
		return 0
	}

	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint64{
			codec.String(this.Hash).Size(),
			codec.Uint64(this.ID).Size(),
			univalue.Univalues(this.Accesses).Size(),
		},
	)

	offset += codec.String(this.Hash).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.ID).EncodeTo(buffer[offset:])
	offset += codec.Bytes(univalue.Univalues(this.Accesses).Encode()).EncodeTo(buffer[offset:])
	return offset
}

func (this *TxAccessRecords) Decode(buffer []byte) *TxAccessRecords {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	this.Hash = codec.Bytes(fields[0]).ToString()
	this.ID = uint64(codec.Uint64(0).Decode(fields[1]).(codec.Uint64))
	uns := univalue.Univalues(this.Accesses).Decode(fields[2])
	this.Accesses = uns.(univalue.Univalues)
	return this
}

type TxAccessRecordSet []*TxAccessRecords

func (this *TxAccessRecordSet) HeaderSize() uint64 {
	return uint64(len(*this)+1) * codec.UINT64_LEN
}

func (this *TxAccessRecordSet) Size() uint64 {
	total := this.HeaderSize()
	for i := 0; i < len(*this); i++ {
		total += (*this)[i].Size()
	}
	return total
}

// Fill in the header info
func (this *TxAccessRecordSet) FillHeader(buffer []byte) {
	offset := uint64(0)
	codec.Uint64(len(*this)).EncodeTo(buffer)
	for i := 0; i < len(*this); i++ {
		codec.Uint64(offset).EncodeTo(buffer[uint64(i+1)*codec.UINT64_LEN:])
		offset += (*this)[i].Size()
	}
}

func (this *TxAccessRecordSet) Encode() []byte {
	bys := make([][]byte, len(*this))
	for i := range *this {
		bys[i] = (*this)[i].Encode()
	}
	return codec.Byteset(bys).Encode()
}

func (this *TxAccessRecordSet) Decode(data []byte) interface{} {
	bytesset := codec.Byteset{}.Decode(data).(codec.Byteset)
	records := make([]*TxAccessRecords, len(bytesset))
	for i := range bytesset {
		this := &TxAccessRecords{}
		this.Decode(bytesset[i])
		records[i] = this
	}
	v := (TxAccessRecordSet)(records)
	return &(v)
}

func (this *TxAccessRecordSet) GobEncode() ([]byte, error) {
	return this.Encode(), nil
}

func (this *TxAccessRecordSet) GobDecode(data []byte) error {
	*this = *(this.Decode(data).(*TxAccessRecordSet))
	return nil
}
