/*
 *   Copyright (c) 2025 Arcology Network

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

package gas

import (
	"encoding/hex"
	"math/big"

	codec "github.com/arcology-network/common-lib/codec"
	eucommon "github.com/arcology-network/eu/common"
)

// PrepayerInfo represents the information of a gas prepayment for a deferred execution contract.
type PrepayerInfo struct {
	Hash           [32]byte // Transaction hash
	TX             uint64   // Transaction number
	From           [20]byte // Sender address
	To             [20]byte // Contract address
	Signature      [4]byte  // Function signature
	PrepayedAmount uint64   // Amount of prepaid gas
	InitialGas     uint64   // Initial gas amount excluding prepaid gas.
	GasUsed        uint64   // Gas used in the transaction, including prepaid gas.
	GasPrice       *big.Int
	GasRemaining   uint64 // Remaining gas after the transaction execution
	Successful     bool   // Whether the transaction was successful
}

func (this *PrepayerInfo) UID() string {
	return this.GenUID(this.To, this.Signature)
}

func (*PrepayerInfo) GenUID(to [20]byte, signature [4]byte) string {
	return hex.EncodeToString(to[:]) + ":" + hex.EncodeToString(signature[:])
}

func (*PrepayerInfo) FromJob(job *eucommon.Job) *PrepayerInfo {
	info := &PrepayerInfo{
		Hash:           job.StdMsg.TxHash,
		TX:             job.StdMsg.ID,
		From:           job.StdMsg.Native.From,
		PrepayedAmount: job.StdMsg.PrepaidGas,
		GasPrice:       job.StdMsg.Native.GasPrice,
		InitialGas:     job.InitialGas,
		GasRemaining:   job.GasRemaining,
	}

	if len(job.StdMsg.Native.Data) > 0 {
		info.Signature = codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data[:])
		// info.Signature = [4]byte(job.StdMsg.Native.Data[:4])
	}

	if job.StdMsg.Native.To != nil {
		info.To = *job.StdMsg.Native.To
	}

	if job.Results != nil {
		info.GasUsed = job.Results.Receipt.GasUsed
		info.Successful = job.Successful()
	}
	return info
}

func (this *PrepayerInfo) Equal(other *PrepayerInfo) bool {
	return this.Hash == other.Hash &&
		this.TX == other.TX &&
		this.From == other.From &&
		this.To == other.To &&
		this.Signature == other.Signature &&
		this.PrepayedAmount == other.PrepayedAmount &&
		this.InitialGas == other.InitialGas &&
		this.GasUsed == other.GasUsed &&
		this.GasRemaining == other.GasRemaining &&
		this.Successful == other.Successful
}

func (this *PrepayerInfo) Size() uint64 {
	return uint64(len(this.Hash) +
		32 + // Hash
		8 + // TX
		20 + // From
		20 + // To
		4 + // Signature
		8 + // PrepayedAmount
		8 + // InitialGas
		8 + // GasUsed
		8 + // GasRemaining
		1, // Successful
	)
}

func (this *PrepayerInfo) FullSize() uint64 {
	return this.HeaderSize() + this.Size()
}

func (this *PrepayerInfo) HeaderSize() uint64 {
	return 10 * 8
}

func (this *PrepayerInfo) Encode() []byte {
	buffer := make([]byte, this.FullSize())
	this.EncodeTo(buffer)
	return buffer
}

func (this *PrepayerInfo) EncodeTo(buffer []byte) int {
	offset := codec.Encoder{}.FillHeader(buffer,
		[]uint64{
			32,
			8,
			20,
			20,
			4,
			8,
			8,
			8,
			8,
			1, // Successful

		},
	)

	offset += codec.Bytes32(this.Hash).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.TX).EncodeTo(buffer[offset:])
	offset += codec.Bytes20(this.From[:]).EncodeTo(buffer[offset:])
	offset += codec.Bytes20(this.To[:]).EncodeTo(buffer[offset:])
	offset += codec.Bytes(this.Signature[:]).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.PrepayedAmount).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.InitialGas).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.GasUsed).EncodeTo(buffer[offset:])
	offset += codec.Uint64(this.GasRemaining).EncodeTo(buffer[offset:])
	codec.Bool(this.Successful).EncodeTo(buffer[offset:])
	return int(this.FullSize())
}

func (*PrepayerInfo) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return nil
	}
	this := &PrepayerInfo{}
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)

	this.Hash = codec.Bytes32{}.Decode(fields[0]).(codec.Bytes32)
	this.TX = uint64(codec.Uint64(0).Decode(fields[1]).(codec.Uint64))
	this.From = codec.Bytes20{}.Decode(fields[2]).(codec.Bytes20)
	this.To = codec.Bytes20{}.Decode(fields[3]).(codec.Bytes20)
	this.Signature = codec.Bytes4{}.Decode(fields[4]).(codec.Bytes4)
	this.PrepayedAmount = uint64(codec.Uint64(0).Decode(fields[5]).(codec.Uint64))
	this.InitialGas = uint64(codec.Uint64(0).Decode(fields[6]).(codec.Uint64))
	this.GasUsed = uint64(codec.Uint64(0).Decode(fields[7]).(codec.Uint64))
	this.GasRemaining = uint64(codec.Uint64(0).Decode(fields[8]).(codec.Uint64))
	this.Successful = bool(codec.Bool(false).Decode(fields[9]).(codec.Bool))
	return this
}

type PrepayerInfoArr []*PrepayerInfo

func (this PrepayerInfoArr) Encode() []byte {
	if this == nil {
		return nil
	}

	offset := 0
	buffer := make([]byte, len(this)*int((&PrepayerInfo{}).FullSize()))
	for i := 0; i < len(this); i++ {
		offset += (this)[i].EncodeTo(buffer[offset:])
	}
	return buffer
}

func (PrepayerInfoArr) Decode(buffer []byte) any {
	if len(buffer) == 0 {
		return nil
	}

	total := (len(buffer)) / int((&PrepayerInfo{}).FullSize())
	result := make(PrepayerInfoArr, total)
	for i := 0; i < len(result); i++ {
		buf := buffer[i*int((&PrepayerInfo{}).FullSize()) : (i+1)*int((&PrepayerInfo{}).FullSize())]
		info := (&PrepayerInfo{}).Decode(buf).(*PrepayerInfo)
		result[i] = info
	}
	return result
}
