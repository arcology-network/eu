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
package stgtest

// func TestGrowOnlySet(t *testing.T) {
// 	store := chooseDataStore()
// 	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
// 	writeCache := sstore.WriteCache

// 	alice := AliceAccount()
// 	if _, err := eth.CreateNewAccount(stgcommon.SYSTEM, alice, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	bob := BobAccount()
// 	if _, err := eth.CreateNewAccount(stgcommon.SYSTEM, bob, writeCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}

// 	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
// 	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Precommit([]uint64{stgcommon.SYSTEM})
// 	committer.Commit(10)

// 	// Write the Prepayer info paths.
// 	_, err := writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/"+stgcommon.PREPAYERS, commutative.NewPath())
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/func/"+stgcommon.PREPAYERS, commutative.NewPath()); v == nil {
// 		t.Error("Error: Wrong value in grow-only set")
// 	}

// 	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/func/"+stgcommon.PREPAYERS, commutative.NewPath())
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	if v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/func/"+stgcommon.PREPAYERS, new(commutative.Path)); v == nil {
// 		t.Error("Error: Wrong value in grow-only set")
// 	}

// 	//Geneate prepayer info
// 	Alice, Bob := GenPrepayerInfo()
// 	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+alice+"/func/"+stgcommon.PREPAYERS+Alice.UID(), noncommutative.NewBytes(Alice.Encode()))
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	_, err = writeCache.Write(1, "blcc://eth1.0/account/"+bob+"/func/"+stgcommon.PREPAYERS+Alice.UID(), noncommutative.NewBytes(Bob.Encode()))
// 	if err != nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	aliceBuf, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+alice+"/func/"+stgcommon.PREPAYERS+Alice.UID(), new(commutative.Path))
// 	if aliceBuf == nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	bobBuf, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+bob+"/func/"+stgcommon.PREPAYERS+Alice.UID(), new(noncommutative.Bytes))
// 	if bobBuf == nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	alicePrepayer := new(gas.PrepayerInfo).Decode(aliceBuf.([]byte)).(*gas.PrepayerInfo)
// 	if !Alice.Equal(alicePrepayer) {
// 		t.Error("Error: The prepayer info should be the same")
// 	}

// 	buffPrepayer := new(gas.PrepayerInfo).Decode(bobBuf.([]byte)).(*gas.PrepayerInfo)
// 	if !Bob.Equal(buffPrepayer) {
// 		t.Error("Error: The prepayer info should be the same")
// 	}

// 	store = chooseDataStore()
// 	sstore = statestore.NewStateStore(store.(*stgproxy.StorageProxy))
// 	newWriteCache := sstore.WriteCache

// 	acctTrans = univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
// 	committer = stgcommitter.NewStateCommitter(store, sstore.GetWriters())
// 	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
// 	committer.Precommit([]uint64{1})
// 	committer.Commit(10)

// 	aliceBuf, _, _ = newWriteCache.Read(1, "blcc://eth1.0/account/"+alice+"/func/"+stgcommon.PREPAYERS+Alice.UID(), new(noncommutative.Bytes))
// 	if aliceBuf == nil {
// 		t.Error("Failed to write prepay function", err)
// 	}

// 	bobBuf, _, _ = newWriteCache.Read(1, "blcc://eth1.0/account/"+bob+"/func/"+stgcommon.PREPAYERS+Alice.UID(), new(noncommutative.Bytes))
// 	if bobBuf == nil {
// 		t.Error("Failed to write prepay function", err)
// 	}
// }
