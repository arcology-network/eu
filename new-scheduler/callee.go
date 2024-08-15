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
	"encoding/hex"
	"strings"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/common-lib/types"
	schtype "github.com/arcology-network/common-lib/types/scheduler"
	stgcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
)

// The callee struct stores the information of a contract function that is called by the EOA initiated transactions.
// It is mainly used to optimize the execution of the transactions. A callee is uniquely identified by a
// combination of the contract's address and the function signature.
type Callee struct {
	Index       uint32     // Index of the Callee in the Callee list
	AddrAndSign [12]byte   // Short address of the callee, first 8 bytes from the function + signature [4]byte
	Indices     []uint32   // Indices of the conflicting callee indices.
	Sequential  bool       // A sequential / parallel only calls
	Except      [][12]byte // Sequntial or Paralle call exceptions

	Calls      uint32 // Total number of calls
	AvgGas     uint32 // Average gas used
	Deferrable bool   // If one of the calls should be deferred to the second generation.
}

func NewCallee(idx uint32, addr []byte, funSign []byte) *Callee {
	return &Callee{
		Index:       idx,
		AddrAndSign: new(codec.Bytes12).FromSlice(schtype.Compact(addr, funSign)),
		Except:      [][12]byte{},
		Deferrable:  false,
	}
}

func (*Callee) IsPropertyPath(path string) bool {
	return len(path) > stgcommon.ETH10_ACCOUNT_FULL_LENGTH &&
		strings.Contains(path[stgcommon.ETH10_ACCOUNT_FULL_LENGTH:], "/func/")
}

// The function creates a compact representation of the callee information
// func (*Callee) Compact(addr []byte, funSign []byte) []byte {
// 	addr = slice.Clone(addr) // Make sure the original data is not modified
// 	return append(addr[:schtype.SHORT_CONTRACT_ADDRESS_LENGTH], funSign[:schtype.FUNCTION_SIGNATURE_LENGTH]...)
// }

// Convert the transaction to a map of callee information
func (this *Callee) ToCallee(trans []*univalue.Univalue) map[string]*Callee {
	propTrans := slice.MoveIf(&trans, func(_ int, v *univalue.Univalue) bool {
		return new(Callee).IsPropertyPath(*v.GetPath())
	})

	dict := map[string]*Callee{}
	for _, v := range propTrans {
		addrAndSign := this.parseCalleeSignature(*v.GetPath())
		if _, ok := dict[addrAndSign]; len(addrAndSign) != 0 && !ok {
			calleeInfo := &Callee{}
			calleeInfo.AddrAndSign = new(codec.Bytes12).FromBytes([]byte(addrAndSign))
			dict[addrAndSign] = calleeInfo
		}
	}
	this.setCalleeInfo(propTrans, dict)
	return dict
}

// Extract the callee signature from the path string
func (this *Callee) parseCalleeSignature(path string) string {
	idx := strings.Index(path, stgcommon.ETH10_FUNC_PROPERTY_PREFIX)
	if idx == len(path) {
		return ""
	}

	fullPath := path[idx+len(stgcommon.ETH10_FUNC_PROPERTY_PREFIX):]
	sign, _ := hex.DecodeString(fullPath)

	if len(sign) == 0 {
		return ""
	}
	addrStr := path[stgcommon.ETH10_ACCOUNT_PREFIX_LENGTH:]
	idx = strings.Index(addrStr, "/")
	addrStr = strings.TrimPrefix(addrStr[:idx], "0x")

	addr, _ := hex.DecodeString(addrStr)
	return string(append(addr[:schtype.SHORT_CONTRACT_ADDRESS_LENGTH], sign...))
}

// Use the transitions to set the callee information
func (this *Callee) setCalleeInfo(trans []*univalue.Univalue, dict map[string]*Callee) {
	for _, tran := range trans {
		addrAndSign := this.parseCalleeSignature(*tran.GetPath())
		calleeInfo := dict[addrAndSign]
		if calleeInfo == nil {
			continue
		}

		// Set execution method
		if strings.HasSuffix(*tran.GetPath(), schtype.EXECUTION_METHOD) && tran.Value() != nil {
			flag, _, _ := tran.Value().(stgcommon.Type).Get()
			calleeInfo.Sequential = flag.([]byte)[0] == schtype.SEQUENTIAL_EXECUTION
		}

		// Set the excepted transitions
		if strings.HasSuffix(*tran.GetPath(), schtype.EXECUTION_EXCEPTED) {
			subPaths, _, _ := tran.Value().(*commutative.Path).Get()
			subPathSet := subPaths.(*deltaset.DeltaSet[string])
			for _, subPath := range subPathSet.Elements() {
				k := new(codec.Bytes12).FromBytes([]byte(subPath))
				calleeInfo.Except = append(calleeInfo.Except, k)
			}
		}

		// Set the Deferrable value
		if strings.HasSuffix(*tran.GetPath(), schtype.DEFERRED_FUNC) && tran.Value() != nil {
			flag, _, _ := tran.Value().(stgcommon.Type).Get()
			calleeInfo.Deferrable = flag.([]byte)[0] > 0
		}
	}
}

// 10x faster and 2x smaller than json marshal/unmarshal
func (this *Callee) Encode() ([]byte, error) {
	return codec.Byteset([][]byte{
		codec.Uint32(this.Index).Encode(),
		this.AddrAndSign[:],
		codec.Uint32s(this.Indices).Encode(),
		codec.Bool(this.Sequential).Encode(),

		codec.Bytes12s(this.Except).Encode(),

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
	this.Sequential = bool(new(codec.Bool).Decode(fields[3]).(codec.Bool))
	this.Except = new(codec.Bytes12s).Decode(fields[4]).(codec.Bytes12s)
	this.Calls = uint32(new(codec.Uint32).Decode(fields[5]).(codec.Uint32))
	this.AvgGas = uint32(new(codec.Uint32).Decode(fields[6]).(codec.Uint32))
	this.Deferrable = bool(new(codec.Bool).Decode(fields[7]).(codec.Bool))
	return this
}

func (this *Callee) Equal(other *Callee) bool {
	return this.Index == other.Index &&
		slice.EqualSet(this.AddrAndSign[:], other.AddrAndSign[:]) &&
		slice.EqualSet(this.Indices, other.Indices) &&
		this.Sequential == other.Sequential &&
		slice.EqualSet(this.Except, other.Except) &&
		this.Calls == other.Calls &&
		this.AvgGas == other.AvgGas &&
		this.Deferrable == other.Deferrable
}

type Callees []*Callee

func (this Callees) Encode() []byte {
	buffer := slice.Transform(this, func(i int, callee *Callee) []byte {
		bytes, _ := (callee).Encode()
		return bytes
	})
	return codec.Byteset(buffer).Encode()
}

func (Callees) Decode(buffer []byte) interface{} {
	buffers := new(codec.Byteset).Decode(buffer).(codec.Byteset)
	callees := make(Callees, len(buffers))
	for i, buf := range buffers {
		callees[i] = new(Callee).Decode(buf)
	}
	return Callees(callees)
}

func (Callees) From(addr []byte, funSigns ...[]byte) [][schtype.CALLEE_ID_LENGTH]byte {
	callees := make([][schtype.CALLEE_ID_LENGTH]byte, len(funSigns))
	for i, funSign := range funSigns {
		callees[i] = new(codec.Bytes12).FromSlice(
			schtype.Compact(addr, funSign),
		)
	}
	return callees
}

// Get the callee key from a message
func ToKey(msg *eucommon.StandardMessage) string {
	if (*msg.Native).To == nil {
		return ""
	}

	if len(msg.Native.Data) == 0 {
		return string((*msg.Native.To)[:schtype.FUNCTION_SIGNATURE_LENGTH])
	}
	return schtype.CallToKey((*msg.Native.To)[:], msg.Native.Data[:schtype.FUNCTION_SIGNATURE_LENGTH])
}

// func CallToKey(addr []byte, funSign []byte) string {
// 	return string(addr[:schtype.FUNCTION_SIGNATURE_LENGTH]) + string(funSign[:schtype.FUNCTION_SIGNATURE_LENGTH])
// }
