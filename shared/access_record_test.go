package shared

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestAccessRecordEncoding(t *testing.T) {
	records := &TxAccessRecords{
		Hash:     "0x1234567",
		ID:       99,
		Accesses: []byte{byte(67), byte(77)},
	}

	buffer := records.Encode()
	out := (&TxAccessRecords{}).Decode(buffer)
	if !reflect.DeepEqual(records, out) {
		t.Error("Error")
	}
}

func TestAccessRecordSetEncoding(t *testing.T) {
	_1 := &TxAccessRecords{
		Hash:     "0x1234567",
		ID:       99,
		Accesses: []byte{byte(66), byte(33)},
	}

	_2 := &TxAccessRecords{
		Hash:     "0xabcde",
		ID:       88,
		Accesses: []byte{byte(44), byte(55)},
	}

	_3 := &TxAccessRecords{
		Hash:     "0x8976542",
		ID:       77,
		Accesses: []byte{byte(66), byte(88)},
	}

	accessSet := TxAccessRecordSet{_1, _2, _3}

	buffer, _ := accessSet.GobEncode()
	out := TxAccessRecordSet{}
	out.GobDecode(buffer)
	if !reflect.DeepEqual(accessSet, out) {
		t.Error("Error")
	}
}

func BenchmarkAccessRecordSetEncoding(b *testing.B) {
	recordVec := make([]*TxAccessRecords, 1000000)
	for i := 0; i < len(recordVec); i++ {
		recordVec[i] = &TxAccessRecords{
			Hash:     "0x1234567",
			ID:       uint32(i),
			Accesses: []byte{byte(99), byte(110)},
		}
	}

	t0 := time.Now()
	set := TxAccessRecordSet(recordVec)
	buffer, _ := (&set).GobEncode()
	fmt.Println("GobEncode():", time.Now().Sub(t0))

	out := new(TxAccessRecordSet)
	t0 = time.Now()
	out.GobDecode(buffer)
	fmt.Println("GobDecode():", time.Now().Sub(t0))
}
