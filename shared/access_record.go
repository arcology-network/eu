package shared

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/array"
)

type TxAccessRecords struct {
	Hash     string
	ID       uint32
	Accesses []byte
}

func (this *TxAccessRecords) HeaderSize() uint32 {
	return 4 * codec.UINT32_LEN
}

func (this *TxAccessRecords) Size() uint32 {
	return this.HeaderSize() +
		codec.String(this.Hash).Size() +
		codec.UINT32_LEN +
		codec.Bytes(this.Accesses).Size()
}

func (this *TxAccessRecords) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.EncodeToBuffer(buffer)
	return buffer
}

func (this *TxAccessRecords) EncodeToBuffer(buffer []byte) int {
	if this == nil {
		return 0
	}

	offset := codec.Encoder{}.FillHeader(
		buffer,
		[]uint32{
			codec.String(this.Hash).Size(),
			codec.Uint32(this.ID).Size(),
			codec.Bytes(this.Accesses).Size(),
		},
	)

	offset += codec.String(this.Hash).EncodeToBuffer(buffer[offset:])
	offset += codec.Uint32(this.ID).EncodeToBuffer(buffer[offset:])
	offset += codec.Bytes(this.Accesses).EncodeToBuffer(buffer[offset:])
	return offset
}

func (this *TxAccessRecords) Decode(buffer []byte) *TxAccessRecords {
	fields := codec.Byteset{}.Decode(buffer).(codec.Byteset)
	this.Hash = codec.Bytes(fields[0]).ToString()
	this.ID = uint32(codec.Uint32(0).Decode(fields[1]).(codec.Uint32))
	this.Accesses = codec.Bytes{}.Decode(fields[2]).(codec.Bytes)
	return this
}

type TxAccessRecordSet []*TxAccessRecords

func (this *TxAccessRecordSet) HeaderSize() uint32 {
	return uint32((len(*this) + 1) * codec.UINT32_LEN)
}

func (this *TxAccessRecordSet) Size() uint32 {
	total := this.HeaderSize()        // Header length
	for i := 0; i < len(*this); i++ { // Body  length
		total += (*this)[i].Size()
	}
	return total
}

// Fill in the header info
func (this *TxAccessRecordSet) FillHeader(buffer []byte) {
	offset := uint32(0)
	codec.Uint32(len(*this)).EncodeToBuffer(buffer)
	for i := 0; i < len(*this); i++ {
		codec.Uint32(offset).EncodeToBuffer(buffer[(i+1)*codec.UINT32_LEN:])
		offset += (*this)[i].Size()
	}
}

func (this *TxAccessRecordSet) Encode() []byte {
	buffer := make([]byte, this.Size())
	this.FillHeader(buffer)

	headerLen := this.HeaderSize()
	offsets := make([]uint32, len(*this)+1)
	offsets[0] = 0
	for i := 0; i < len(*this); i++ {
		offsets[i+1] = offsets[i] + (*this)[i].Size()
	}

	array.ParallelForeach(*this, 4, func(i int, _ **TxAccessRecords) {
		(*this)[i].EncodeToBuffer(buffer[headerLen+offsets[i]:])
	})
	return buffer
}

func (this *TxAccessRecordSet) Decode(data []byte) interface{} {
	bytesset := codec.Byteset{}.Decode(data).(codec.Byteset)
	records := array.ParallelAppend(bytesset, 6, func(i int, _ []byte) *TxAccessRecords {
		this := &TxAccessRecords{}
		this.Decode(bytesset[i])
		return this
	})

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
