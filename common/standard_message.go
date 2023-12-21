package common

import (
	"bytes"
	"sort"

	evmcore "github.com/ethereum/go-ethereum/core"
)

type StandardMessage struct {
	ID     uint64
	TxHash [32]byte
	Native *evmcore.Message
	Source uint8
}

type StandardMessages []*StandardMessage

func (this StandardMessages) SortByFee() {
	// this.Native.
	sort.SliceStable(
		this,
		func(i, j int) bool {
			return this[i].Native.GasLimit < this[j].Native.GasLimit
		},
	)
}

func (this StandardMessages) SortByHash() {
	sort.Slice(this, func(i, j int) bool { return string(this[i].TxHash[:]) < string(this[j].TxHash[:]) })
}

func (this StandardMessages) Count(value *StandardMessage) int {
	counter := 0
	for i := range this {
		if bytes.Equal(this[i].TxHash[:], value.TxHash[:]) {
			counter++
		}
	}
	return counter
}

// func (this StandardMessages) FromSequence(baseApiRouter eucommon.EthApiRouter) []*StandardMessage {
// 	jobs := make([]*Job, len(this.Msgs))
// 	for i, msg := range this.Msgs {
// 		jobs[i].Predecessors = this.Predecessors
// 		jobs[i].Message = msg.Native
// 		jobs[i].ApiRouter = baseApiRouter
// 	}
// 	return jobs
// }

// func (this StandardMessages) Hashes() []evmcommon.Hash {
// 	hashes := make([]evmcommon.Hash, len(this))
// 	for i := range this {
// 		hashes[i] = this[i].TxHash
// 	}
// 	return hashes
// }

// func (this StandardMessages) QuickSort(less func(this *StandardMessage, rgt *StandardMessage) bool) {
// 	if len(this) < 2 {
// 		return
// 	}
// 	left, right := 0, len(this)-1
// 	pivotIndex := rand.Int() % len(this)

// 	this[pivotIndex], this[right] = this[right], this[pivotIndex]
// 	for i := range this {
// 		if less(this[i], this[right]) {
// 			this[i], this[left] = this[left], this[i]
// 			left++
// 		}
// 	}
// 	this[left], this[right] = this[right], this[left]

// 	StandardMessages(this[:left]).QuickSort(less)
// 	StandardMessages(this[left+1:]).QuickSort(less)
// }

// func (this StandardMessages) EncodeToBytes() [][]byte {
// 	if this == nil {
// 		return [][]byte{}
// 	}
// 	data := make([][]byte, len(this))
// 	worker := func(start, end, idx int, args ...interface{}) {
// 		this := args[0].([]interface{})[0].(StandardMessages)
// 		data := args[0].([]interface{})[1].([][]byte)

// 		for i := start; i < end; i++ {
// 			if encoded, err := this[i].Native.GobEncode(); err == nil {
// 				tmpData := [][]byte{
// 					this[i].TxHash.Bytes(),
// 					[]byte{this[i].Source},
// 					encoded,
// 					//this[i].TxRawData,
// 					[]byte{}, //remove TxRawData
// 				}
// 				data[i] = encoding.Byteset(tmpData).Encode()
// 			}
// 		}
// 	}
// 	common.ParallelWorker(len(this), concurrency, worker, this, data)
// 	return data
// }

// func (this StandardMessages) Encode() ([]byte, error) {
// 	if this == nil {
// 		return []byte{}, nil
// 	}
// 	data := make([][]byte, len(this))
// 	worker := func(start, end, idx int, args ...interface{}) {
// 		this := args[0].([]interface{})[0].(StandardMessages)
// 		data := args[0].([]interface{})[1].([][]byte)

// 		for i := start; i < end; i++ {
// 			if encoded, err := this[i].Native.GobEncode(); err == nil {
// 				//data[i] = encoding.Byteset([][]byte{this[i].TxHash.Bytes()[:], {this[i].Source}, encoded}).Flatten()
// 				tmpData := [][]byte{
// 					this[i].TxHash.Bytes(),
// 					[]byte{this[i].Source},
// 					encoded,
// 					this[i].TxRawData,
// 				}
// 				data[i] = encoding.Byteset(tmpData).Encode()
// 			}
// 		}
// 	}
// 	common.ParallelWorker(len(this), concurrency, worker, this, data)
// 	return encoding.Byteset(data).Encode(), nil
// }

// func (this *StandardMessages) Decode(data []byte) ([]*StandardMessage, error) {
// 	fields := encoding.Byteset{}.Decode(data)
// 	msgs := make([]*StandardMessage, len(fields))

// 	worker := func(start, end, idx int, args ...interface{}) {
// 		data := args[0].([]interface{})[0].([][]byte)
// 		messages := args[0].([]interface{})[1].([]*StandardMessage)

// 		for i := start; i < end; i++ {
// 			standredMessage := new(StandardMessage)

// 			fields := encoding.Byteset{}.Decode(data[i])
// 			standredMessage.TxHash = evmcommon.BytesToHash(fields[0])
// 			standredMessage.Source = uint8(fields[1][0])
// 			msg := new(ethTypes.Message)
// 			err := msg.GobDecode(fields[2])
// 			if err != nil {
// 				return
// 			}
// 			standredMessage.Native = msg
// 			standredMessage.TxRawData = fields[3]

// 			messages[i] = standredMessage
// 		}
// 	}
// 	common.ParallelWorker(len(fields), concurrency, worker, fields, msgs)

// 	return msgs, nil
// }
