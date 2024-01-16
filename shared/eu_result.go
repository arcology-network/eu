package shared

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/univalue"
)

type EuResult struct {
	H            string
	ID           uint32
	Transitions  []byte
	TransitTypes []byte
	// DC           *DeferredCall
	Trans   []*univalue.Univalue
	Status  uint64
	GasUsed uint64
}

func (this *EuResult) HeaderSize() uint32 {
	return 8 * codec.UINT32_LEN
}

func (this *EuResult) Size() uint32 {
	return this.HeaderSize() +
		uint32(len(this.H)) +
		codec.UINT32_LEN +
		codec.Bytes(this.Transitions).Size() +
		codec.Bytes(this.TransitTypes).Size() +
		// this.DC.Size() +
		codec.UINT64_LEN +
		codec.UINT64_LEN
}

func (this *EuResult) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *EuResult) EncodeToBuffer(buffer []byte) int {
	if this == nil {
		return 0
	}

	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			codec.String(this.H).Size(),
			codec.Uint32(this.ID).Size(),
			codec.Bytes(this.Transitions).Size(),
			codec.Bytes(this.TransitTypes).Size(),
			// this.DC.Size(),
			codec.UINT64_LEN,
			codec.UINT64_LEN,
		},
	)

	offset += codec.String(this.H).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.ID).EncodeToBuffer(buffer[offset:])
	offset += codec.Bytes(this.Transitions).EncodeToBuffer(buffer[offset:])
	offset += codec.Bytes(this.TransitTypes).EncodeToBuffer(buffer[offset:])
	// offset += this.DC.EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64(this.Status).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint64(this.GasUsed).EncodeToBuffer(buffer[offset:])

	return offset
}

func (this *EuResult) Decode(buffer []byte) *EuResult {
	fields := [][]byte(codec.Byteset{}.Decode(buffer).(codec.Byteset))

	this.H = string(fields[0])
	this.ID = uint32(codec.Uint32(0).Decode(fields[1]).(codec.Uint32))

	this.Transitions = []byte(codec.Bytes{}.Decode(fields[2]).(codec.Bytes))
	this.TransitTypes = []byte(codec.Bytes{}.Decode(fields[3]).(codec.Bytes))

	// if len(fields[4]) > 0 {
	// 	this.DC = (&DeferredCall{}).Decode(fields[4])
	// }
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

func (this *Euresults) HeaderSize() uint32 {
	return uint32((len(*this) + 1) * codec.UINT32_LEN) // Header length
}

func (this *Euresults) Size() uint32 {
	total := this.HeaderSize()
	for i := 0; i < len(*this); i++ {
		total += (*this)[i].Size()
	}
	return total
}

// Fill in the header info
func (this *Euresults) FillHeader(buffer []byte) {
	codec.Uint32(len(*this)).EncodeToBuffer(buffer)

	offset := uint32(0)
	for i := 0; i < len(*this); i++ {
		codec.Uint32(offset).EncodeToBuffer(buffer[codec.UINT32_LEN*(i+1):])
		offset += (*this)[i].Size()
	}
}

func (this Euresults) GobEncode() ([]byte, error) {
	buffer := make([]byte, this.Size())
	this.FillHeader(buffer)

	offsets := make([]uint32, len(this)+1)
	offsets[0] = 0
	for i := 0; i < len(this); i++ {
		offsets[i+1] = offsets[i] + this[i].Size()
	}

	headerLen := this.HeaderSize()
	worker := func(start, end, index int, args ...interface{}) {
		for i := start; i < end; i++ {
			this[i].EncodeToBuffer(buffer[headerLen+offsets[i]:])
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
