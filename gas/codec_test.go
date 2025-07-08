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

package gas

import (
	"bytes"
	"fmt"
	"math/big"
	"reflect"
	"testing"
	"time"

	codec "github.com/arcology-network/common-lib/codec"

	rlp "github.com/ethereum/go-ethereum/rlp"
)

func TestPrepayerCodec(t *testing.T) {
	Alice := &PrepayerInfo{}
	Alice.Hash = [32]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
		0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14,
		0x15, 0x16, 0x17, 0x18, 0x19, 0x1a,
		0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

	Alice.TX = 12345
	Alice.From = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}

	Alice.To = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}
	Alice.Signature = [4]byte{0x12, 0x34, 0x56, 0x78}
	Alice.GasUsed = 100000
	Alice.InitialGas = 200000
	Alice.PrepayedAmount = 50000
	Alice.GasRemaining = 150000
	Alice.Successful = true

	AliceBuffer := Alice.Encode()
	AliceBackup := (&PrepayerInfo{}).Decode(AliceBuffer)

	if !reflect.DeepEqual(Alice, AliceBackup) {
		t.Error("Error: PrepayerInfo Encoding/decoding error, structs don't match")
	}

	Bob := &PrepayerInfo{}
	Bob.Hash = [32]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26,
		0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30,
		0x31, 0x32, 0x33, 0x34, 0x35,
		0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b,
		0x3c, 0x3d, 0x3e,
		0x3f, 0x40}

	Bob.TX = 54321

	Bob.From = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

	Bob.To = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

	Bob.Signature = [4]byte{0x56, 0x78, 0x9a, 0xbc}
	Bob.GasUsed = 200000
	Bob.InitialGas = 300000
	Bob.PrepayedAmount = 100000
	Bob.GasRemaining = 200000
	Bob.Successful = false

	BobBuffer := Bob.Encode()
	BobBackup := (&PrepayerInfo{}).Decode(BobBuffer)
	if !reflect.DeepEqual(Bob, BobBackup) {
		t.Error("Error: PrepayerInfo Encoding/decoding error, structs don't match")
	}

	infoArr := []*PrepayerInfo{Alice, Bob} // Create an array of PrepayerInfo

	infoArrBuffer := PrepayerInfoArr(infoArr).Encode()

	if !bytes.Equal(infoArrBuffer[:len(AliceBuffer)], AliceBuffer) {
		t.Error("Error: PrepayerInfoArr Encoding/decoding error, Alice's data mismatch")
	}

	infoArrBackup := PrepayerInfoArr{}.Decode(infoArrBuffer).(PrepayerInfoArr)

	if len(infoArrBackup) != 2 {
		t.Error("Error: PrepayerInfoArr Encoding/decoding error, array length mismatch")
	}

	if !infoArrBackup[0].Equal(Alice) || !infoArrBackup[1].Equal(Bob) { // Check if Alice's data matches
		t.Error("Error: PrepayerInfoArr Encoding/decoding error, Alice's data mismatch")
	}

}

func BenchmarkRlpComparePerformance(t *testing.B) {
	num := big.NewInt(100)

	expected, err := rlp.EncodeToBytes(num)
	if err != nil {
		t.Error(expected, err)
	}

	var decoded big.Int
	if err := rlp.DecodeBytes(expected, &decoded); err != nil {
		t.Error(expected, err)
	}

	if num.Cmp(&decoded) != 0 {
		t.Error("Mismatch")
	}

	t0 := time.Now()
	for i := 0; i < 1000000; i++ {
		num = big.NewInt(100)
	}
	fmt.Println("big NewInt RLP Encode:            "+fmt.Sprint(1000000), time.Since(t0))

	t0 = time.Now()
	for i := 0; i < 1000000; i++ {
		v := codec.Bigint(*num)
		v.Encode()
	}
	fmt.Println("big NewInt Codec Encode:            "+fmt.Sprint(1000000), time.Since(t0))
}
