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

package tests

import (
	"math/big"
	"os"
	"path"
	"path/filepath"
	"testing"

	slice "github.com/arcology-network/common-lib/exp/slice"
	eu "github.com/arcology-network/eu"
	execution "github.com/arcology-network/eu"
	"github.com/arcology-network/evm-adaptor/compiler"
	statestore "github.com/arcology-network/storage-committer"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	tests "github.com/arcology-network/storage-committer/tests"
	"github.com/arcology-network/storage-committer/univalue"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSequence(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "native")
	code, err := compiler.CompileContracts(targetPath, "/NativeStorage.sol", "0.8.19", "NativeStorage", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	seq := execution.NewJobSequence(1, []uint64{1}, []*evmcore.Message{&deployMsg}, [32]byte{}, api)
	seq.Run(testEu.config, testEu.eu.Api(), 0)
	contractAddr := seq.Results[0].Receipt.ContractAddress

	tests.FlushToStore(testEu.store.(*statestore.StateStore))
	// acctTrans := univalue.Univalues(slice.Clone(accesses)).To(committer.IPTransition{})
	// committer := stgcomm.NewStateCommitter(testEu.store)
	// committer.Import(acctTrans).Precommit([]uint32{1})
	// committer.Commit(20)
	// committer.Clear()

	// Prepare the messages for the contract calls
	data := crypto.Keccak256([]byte("call()"))[:4]
	msgCallAdd1 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("check()"))[:4]
	msgCallAdd2 := core.NewMessage(Alice, &contractAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	// Put the messages into the sequence and run it in sequence.
	testEu.eu.Api().WriteCache().(*cache.WriteCache).Clear()
	seq = execution.NewJobSequence(1, []uint64{1, 2}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2), [32]byte{}, testEu.eu.Api())
	seq.Run(testEu.config, api, 0)
}

func TestSequence2(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "native")
	code, err := compiler.CompileContracts(targetPath, "/Sequential.sol", "0.8.19", "SequentialTest", true)
	if err != nil || len(code) == 0 {
		t.Fatal("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	seq := execution.NewJobSequence(1, []uint64{1}, []*evmcore.Message{&deployMsg}, [32]byte{}, api)
	seq.Run(testEu.config, testEu.eu.Api(), 0)
	contractAddr := seq.Results[0].Receipt.ContractAddress

	// seq.SeqAPI.WriteCache().(*cache.WriteCache).FlushToStore(testEu.store)
	// tests.FlushToStore(seq.SeqAPI.WriteCache().(*cache.WriteCache), testEu.store)
	tests.FlushToStore(testEu.store.(*statestore.StateStore))
	// // Prepare the messages for the contract calls
	data := crypto.Keccak256([]byte("add()"))[:4]
	msgCallAdd1 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("add()"))[:4]
	msgCallAdd2 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("check()"))[:4]
	msgCallCheck := core.NewMessage(Alice, &contractAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	// // Put the messages into the sequence and run it in sequence.
	// testEu.eu.Api().WriteCache().(*cache.WriteCache).Clear()
	seq = execution.NewJobSequence(1, []uint64{1, 2, 3}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2, &msgCallCheck), [32]byte{}, testEu.eu.Api())
	seq.Run(testEu.config, api, 0)
}

func TestGeneration(t *testing.T) {
	// ================================== Compile the 1st contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "native")
	codeNativeStorage, err := compiler.CompileContracts(targetPath, "/NativeStorage.sol", "0.8.19", "NativeStorage", true)
	if err != nil || len(codeNativeStorage) == 0 {
		t.Fatal("Error: Failed to generate the byte code")
	}
	deployNativeStorageMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(codeNativeStorage), nil, false)

	// ================================== Compile the 2nd contract ==================================
	targetPath = path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "native")
	codeSequential, err := compiler.CompileContracts(targetPath, "/Sequential.sol", "0.8.19", "SequentialTest", true)
	if err != nil || len(codeSequential) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}
	deploySequentialMsg := core.NewMessage(Alice, nil, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(codeSequential), nil, false)

	// ================================== contract Deployment  ==================================
	testEu := NewTestEU(Coinbase, Alice, Bob)
	api := testEu.eu.Api()
	_0thSeq := execution.NewJobSequence(1, []uint64{1, 2, 3}, slice.ToSlice(&deployNativeStorageMsg, &deploySequentialMsg), [32]byte{1}, testEu.eu.Api())
	// _, seqUniv := _0thSeq.Run(testEu.config, api, 0)
	gen := eu.NewGeneration(0, 2, []*execution.JobSequence{_0thSeq})
	gen.Execute(testEu.config, api)

	tests.FlushToStore(testEu.store.(*statestore.StateStore))

	// ================================== 1st contract Call  ==================================
	contractNativeStorageAddr := _0thSeq.Results[0].Receipt.ContractAddress
	msgNativeCall := core.NewMessage(Alice, &contractNativeStorageAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("call()"))[:4], nil, false)
	msgNativeCheck := core.NewMessage(Alice, &contractNativeStorageAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check()"))[:4], nil, false)
	nativeSeq := execution.NewJobSequence(1, []uint64{1, 2}, slice.ToSlice(&msgNativeCall, &msgNativeCheck), [32]byte{1}, testEu.eu.Api())

	// ================================== 2nd contract Call  ==================================
	contractSequentialAddr := _0thSeq.Results[1].Receipt.ContractAddress
	msgSequentialAdd := core.NewMessage(Alice, &contractSequentialAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("add()"))[:4], nil, false)
	msgSequentialCheck := core.NewMessage(Alice, &contractSequentialAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check()"))[:4], nil, false)
	sequentialSeq := execution.NewJobSequence(2, []uint64{3, 4}, slice.ToSlice(&msgSequentialAdd, &msgSequentialCheck), [32]byte{1}, testEu.eu.Api())

	_1stGen := eu.NewGeneration(0, 2, []*execution.JobSequence{nativeSeq, sequentialSeq})
	clearTransitions := _1stGen.Execute(testEu.config, api) // Export transitions

	// // ================================== Commit to DB  ==================================
	acctTrans := univalue.Univalues(clearTransitions).To(univalue.IPTransition{})
	testEu.eu.Api().WriteCache().(*cache.WriteCache).Insert(acctTrans)

	msgNativeCheck2 := core.NewMessage(Alice, &contractNativeStorageAddr, 3, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check2()"))[:4], nil, false)
	// msgSequentialCheck2 := core.NewMessage(Alice, &contractSequentialAddr, 4, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check2()"))[:4], nil, false)

	seq := execution.NewJobSequence(1, []uint64{1}, slice.ToSlice(&msgNativeCheck2), [32]byte{}, testEu.eu.Api())
	_2ndGen := eu.NewGeneration(0, 2, []*execution.JobSequence{seq})
	_2ndGen.Execute(testEu.config, testEu.eu.Api())

	// Add again
	addMsg := core.NewMessage(Alice, &contractNativeStorageAddr, 4, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("call2()"))[:4], nil, false)
	seq = execution.NewJobSequence(1, []uint64{1}, slice.ToSlice(&addMsg), [32]byte{}, testEu.eu.Api())
	clearTransitions = eu.NewGeneration(0, 2, []*execution.JobSequence{seq}).Execute(testEu.config, testEu.eu.Api())
	acctTrans = univalue.Univalues(clearTransitions).To(univalue.IPTransition{})

	testEu.eu.Api().WriteCache().(*cache.WriteCache).Clear().Insert(acctTrans)

	checkMsg := core.NewMessage(Alice, &contractNativeStorageAddr, 5, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check3()"))[:4], nil, false)
	seq = execution.NewJobSequence(1, []uint64{1}, slice.ToSlice(&checkMsg), [32]byte{}, testEu.eu.Api())
	eu.NewGeneration(0, 2, []*execution.JobSequence{seq}).Execute(testEu.config, testEu.eu.Api())
}

func TestMultiCummutiaves(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "lib", "commutative")
	code, err := compiler.CompileContracts(targetPath, "/u256Cum_test.sol", "0.8.19", "MultiCummutative", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	seq := execution.NewJobSequence(1, []uint64{1}, []*evmcore.Message{&deployMsg}, [32]byte{}, api)
	seq.Run(testEu.config, testEu.eu.Api(), 0)
	contractAddr := seq.Results[0].Receipt.ContractAddress

	// Move all the transitions from local write cache to the global write cache, so they can be inserted.
	wcache := seq.SeqAPI.WriteCache().(*cache.WriteCache)
	testEu.store.(*statestore.StateStore).WriteCache.Insert(wcache.Export())
	tests.FlushGeneration(testEu.store.(*statestore.StateStore))
	testEu.store.(*statestore.StateStore).Commit(10)
	testEu.store.(*statestore.StateStore).Clear()

	// Prepare the messages for the contract calls
	data := crypto.Keccak256([]byte("add1()"))[:4]
	msgCallAdd1 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("add2()"))[:4]
	msgCallAdd2 := core.NewMessage(Alice, &contractAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	// // Put the messages into the sequence and run it in sequence.
	testEu.eu.Api().WriteCache().(*cache.WriteCache).Clear()
	seq = execution.NewJobSequence(1, []uint64{1, 2}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2), [32]byte{}, testEu.eu.Api())
	seq.Run(testEu.config, api, 0)

	if seq.Results[0].Receipt.Status != 1 || seq.Results[1].Receipt.Status != 1 {
		t.Error("Error: Failed to call")
	}

	// Move all the transitions from local write cache to the global write cache, so they can be inserted.
	wcache = seq.SeqAPI.WriteCache().(*cache.WriteCache)
	testEu.store.(*statestore.StateStore).WriteCache.Insert(wcache.Export())
	tests.FlushGeneration(testEu.store.(*statestore.StateStore))

	data = crypto.Keccak256([]byte("check()"))[:4]
	msgCallCheck := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	testEu.eu.Api().WriteCache().(*cache.WriteCache).Clear()
	seq = execution.NewJobSequence(1, []uint64{1, 2}, slice.ToSlice(&msgCallCheck), [32]byte{}, testEu.eu.Api())
	seq.Run(testEu.config, api, 0)

	if seq.Results[0].Receipt.Status != 1 {
		t.Error("Error: Failed to call")
	}
}
