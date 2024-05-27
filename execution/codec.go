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

package execution

import (
	"github.com/arcology-network/common-lib/codec"
)

func (this *Result) HeaderSize() uint32 {
	return 8 * codec.UINT32_LEN
}

// func (this *Result) Size() uint32 {
// 	return this.HeaderSize() +
// 		uint32(len(this.H)) +
// 		codec.UINT32_LEN +
// 		codec.Byteset(this.Transitions).Size() +
// 		codec.Bytes(this.TransitTypes).Size() +
// 		this.DC.Size() +
// 		codec.UINT64_LEN +
// 		codec.UINT64_LEN
// }

// func (this *Result) Encode() []byte {
// 	buffer := make([]byte, this.Size())
// 	this.EncodeToBuffer(buffer)
// 	return buffer
// }

// func (this *Result) EncodeToBuffer(buffer []byte) int {
// 	if this == nil {
// 		return 0
// 	}

// 	offset := codec.Encoder{}.FillHeader(
// 		buffer,
// 		[]uint32{
// 			codec.String(this.H).Size(),
// 			codec.Uint32(this.ID).Size(),
// 			codec.Byteset(this.Transitions).Size(),
// 			codec.Bytes(this.TransitTypes).Size(),
// 			this.DC.Size(),
// 			codec.UINT64_LEN,
// 			codec.UINT64_LEN,
// 		},
// 	)

// 	offset += codec.String(this.H).EncodeToBuffer(buffer[offset:])
// 	offset += codec.Uint32(this.ID).EncodeToBuffer(buffer[offset:])
// 	offset += codec.Byteset(this.Transitions).EncodeToBuffer(buffer[offset:])
// 	offset += codec.Bytes(this.TransitTypes).EncodeToBuffer(buffer[offset:])
// 	offset += this.DC.EncodeToBuffer(buffer[offset:])
// 	offset += codec.Uint64(this.Status).EncodeToBuffer(buffer[offset:])
// 	offset += codec.Uint64(this.GasUsed).EncodeToBuffer(buffer[offset:])

// 	return offset
// }

// func (this *Result) Decode(buffer []byte) *Result {
// 	fields := [][]byte(codec.Byteset{}.Decode(buffer).(codec.Byteset))

// 	this.H = string(fields[0])
// 	this.ID = uint32(codec.Uint32(0).Decode(fields[1]).(codec.Uint32))

// 	this.Transitions = [][]byte(codec.Byteset{}.Decode(fields[2]).(codec.Byteset))
// 	this.TransitTypes = []byte(codec.Bytes{}.Decode(fields[3]).(codec.Bytes))

// 	if len(fields[4]) > 0 {
// 		this.DC = (&DeferredCall{}).Decode(fields[4])
// 	}
// 	this.Status = uint64(codec.Uint64(0).Decode(fields[5]).(codec.Uint64))
// 	this.GasUsed = uint64(codec.Uint64(0).Decode(fields[6]).(codec.Uint64))
// 	return this
// }

// func (this *Result) GobEncode() ([]byte, error) {
// 	return this.Encode(), nil
// }

// func (this *Result) GobDecode(buffer []byte) error {
// 	this.Decode(buffer)
// 	return nil
// }
