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

package stgtest

import (
	"fmt"
	"math"
	"math/big"
	"testing"

	"github.com/arcology-network/common-lib/exp/slice"
	typeexec "github.com/arcology-network/common-lib/types/execution"

	stgcommcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	platform "github.com/arcology-network/common-lib/types/storage/platform"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	cache "github.com/arcology-network/common-lib/types/storage/writecache"
	statestore "github.com/arcology-network/storage-committer"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	"github.com/holiman/uint256"
)

func TestSimpleBalance(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewUnboundedU256()); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Export variables
	in := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	buffer := univalue.Univalues(in).Encode()
	out := univalue.Univalues{}.Decode(buffer).(univalue.Univalues)
	for i := range in {
		if !in[i].Equal(out[i]) {
			t.Error("Accesses don't match")
		}
	}

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(out)
	committer.Precommit([]uint32{stgcommcommon.SYSTEM, 0, 1})
	committer.Commit(10)
	// Read alice's balance again

	balance, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))
	balanceAddr := balance.(uint256.Int)
	if (&balanceAddr).Cmp(uint256.NewInt(33)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value", balanceAddr)
	}

	writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(10), true))
	balance, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))

	balanceAddr = balance.(uint256.Int)
	if (&balanceAddr).Cmp(uint256.NewInt(43)) != 0 {
		t.Error("Error: Wrong blcc://eth1.0/account/alice/balance value", balanceAddr)
	}

	trans := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	records := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITAccess{})

	univalue.Univalues(trans).Encode()
	for _, v := range records {
		if v.Writes() == v.Reads() && v.Writes() == 0 && v.DeltaWrites() == 0 {
			t.Error("Error: Write == Reads == DeltaWrites == 0")
		}
	}
}

func TestBalance(t *testing.T) {
	store := chooseDataStore()

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	alice := AliceAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	// create a path
	path := commutative.NewPath()
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/", path); err != nil {
		t.Error(err, " Failed to MakePath: blcc://eth1.0/account/alice/storage/ctrn-0/")
	}

	// create a noncommutative bigint
	inV := noncommutative.NewBigint(100)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", inV); err != nil {
		t.Error(err, " Failed to Write: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/elem-0", new(noncommutative.Bigint))
	outV := v.(big.Int)
	value := (*big.Int)(inV.(*noncommutative.Bigint))
	if outV.Cmp(value) != 0 {
		t.Error("Failed to read: blcc://eth1.0/account/alice/storage/ctrn-0/elem-0")
	}

	// -------------------Create another commutative bigint ------------------------------
	comtVInit := commutative.NewBoundedU256(&commutative.U256_MIN, &commutative.U256_MAX)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", comtVInit); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(300), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(1), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", commutative.NewU256Delta(uint256.NewInt(2), true)); err != nil {
		t.Error(err, " Failed to Write: "+"/elem-0")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/storage/ctrn-0/comt-0", new(commutative.Path))
	vAdd := v.(uint256.Int)
	if vAdd.Cmp(uint256.NewInt(303)) != 0 {
		t.Error("comt-0 has a wrong returned value")
	}

	// ----------------------------U256 ---------------------------------------------------
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(0), true)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the first delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(22), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Add the second delta
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/balance", commutative.NewU256Delta(uint256.NewInt(11), true)); err != nil {
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	// Read alice's balance
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/balance", new(commutative.U256))
	vAdd = v.(uint256.Int)
	if vAdd.Cmp(uint256.NewInt(33)) != 0 {
		t.Error("blcc://eth1.0/account/" + alice + "/balance")
	}

	// Export variables
	transitions := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	// for i := range transitions {
	trans := transitions[9]

	_10 := trans.Encode()
	_10tran := (&univalue.Univalue{}).Decode(_10).(*univalue.Univalue)

	if !trans.Equal(_10tran) {
		t.Error("Accesses don't match", trans, _10tran)
	}
}

func TestNonce(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewBoundedUint64(0, math.MaxInt64)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(2)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(3)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	nonce, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	v := nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce should be ", 6)
	}

	trans := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(trans)

	committer.Precommit([]uint32{0})
	committer.Commit(10)
	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+alice+"/nonce", new(commutative.Uint64))
	v = nonce.(uint64)
	if v != 6 {
		t.Error("Error: blcc://eth1.0/account/alice/nonce ")
	}
}

func TestMultipleNonces(t *testing.T) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	NewAcountsInCache(writeCache, AliceAccount(), BobAccount())

	alice := AliceAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUnboundedUint64())
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+alice+"/balance")
	}

	trans0 := univalue.Univalues((writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})

	bob := BobAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}

	writeCache.Write(0, "blcc://eth1.0/account/"+alice+"/nonce", commutative.NewUnboundedUint64())
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+bob+"/nonce", commutative.NewUint64Delta(1)); err != nil { //initialization
		t.Error(err, "blcc://eth1.0/account/"+bob+"/balance")
	}

	nonce, _, _ := writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce := nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	raw := (writeCache.Export(univalue.Sorter))
	trans1 := univalue.Univalues(raw).To(univalue.ITTransition{})

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(trans0)
	committer.Import(trans1)

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be 2", " actual: ", bobNonce)
	}

	// committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Precommit([]uint32{0, stgcommcommon.SYSTEM})
	committer.Commit(10)

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be 2", " actual: ", bobNonce)
	}

	nonce, _, _ = writeCache.Read(0, "blcc://eth1.0/account/"+bob+"/nonce", new(commutative.Uint64))
	bobNonce = nonce.(uint64)
	if bobNonce != 2 {
		t.Error("Error: blcc://eth1.0/account/bob/nonce should be ", 2)
	}
}

func TestUint64Delta(t *testing.T) {
	store := stgproxy.NewMemDBStoreProxy()
	sstore := statestore.NewStateStore(store)
	writeCache := sstore.WriteCache

	alice := AliceAccount()
	if _, err := typeexec.CreateNewAccount(stgcommcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
		t.Error(err)
	}
	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})

	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans).Precommit([]uint32{stgcommcommon.SYSTEM})
	committer.Commit(stgcommcommon.SYSTEM)

	deltav1 := commutative.NewUint64Delta(11)
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/nonce", deltav1); err != nil {
		t.Error(err)
	}

	writeCache2 := cache.NewWriteCache(store, 1, 1, platform.NewPlatform())
	deltav2 := commutative.NewUint64Delta(21)
	if _, err := writeCache2.Write(2, "blcc://eth1.0/account/"+alice+"/nonce", deltav2); err != nil {
		t.Error(err)
	}

	acctTrans0 := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.IPTransition{})
	acctTrans1 := univalue.Univalues(slice.Clone(writeCache2.Export(univalue.Sorter))).To(univalue.IPTransition{})

	acctTrans0.Print()
	fmt.Println("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&")
	acctTrans1.Print()

	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(acctTrans0)
	committer.Import(acctTrans1)
	committer.Precommit([]uint32{1, 2})
	committer.Commit(10)

}
