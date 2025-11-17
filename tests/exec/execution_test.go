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

package exectest

import (
	"math/big"
	"os"
	"path"
	"path/filepath"
	"testing"

	slice "github.com/arcology-network/common-lib/exp/slice"
	eu "github.com/arcology-network/eu"
	"github.com/arcology-network/eu/compiler"
	tests "github.com/arcology-network/eu/tests/storage"
	workload "github.com/arcology-network/scheduler/workload"
	statestore "github.com/arcology-network/storage-committer"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	statecell "github.com/arcology-network/storage-committer/type/statecell"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestSequence(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "native")
	code, err := compiler.CompileContracts(targetPath, "/NativeStorage.sol", "0.8.19", "NativeStorage", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	seq := new(workload.JobSequence).FromEthMessages(1, []uint64{1}, []*evmcore.Message{&deployMsg}, slice.New(1, [32]byte{}))
	// seq.ExecuteSequence(testEu.config, testEu.eu.Api(), 0)
	eu.ExecuteSequence(seq, testEu.config, testEu.eu.Api(), 0)
	contractAddr := seq.Jobs[0].Result.Receipt.ContractAddress

	tests.FlushToStore(testEu.store.(*statestore.StateStore))
	// acctTrans := statecell.StateCells(slice.Clone(accesses)).To(committer.IPTransition{})
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
	testEu.eu.Api().StateCache().(*cache.StateCache).Clear()
	seq = new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2), slice.New(2, [32]byte{}))
	eu.ExecuteSequence(seq, testEu.config, api, 0)
}

func TestSequence2(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "native")
	code, err := compiler.CompileContracts(targetPath, "/Sequential.sol", "0.8.19", "SequentialTest", true)
	if err != nil || len(code) == 0 {
		t.Fatal("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	seq := new(workload.JobSequence).FromEthMessages(1, []uint64{1}, []*evmcore.Message{&deployMsg}, slice.New(1, [32]byte{}))
	eu.ExecuteSequence(seq, testEu.config, testEu.eu.Api(), 0)
	contractAddr := seq.Jobs[0].Result.Receipt.ContractAddress

	// seq.SeqAPI.StateCache().(*cache.StateCache).FlushToStore(testEu.store)
	// tests.FlushToStore(seq.SeqAPI.StateCache().(*cache.StateCache), testEu.store)
	tests.FlushToStore(testEu.store.(*statestore.StateStore))
	// // Prepare the messages for the contract calls
	data := crypto.Keccak256([]byte("add()"))[:4]
	msgCallAdd1 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("add()"))[:4]
	msgCallAdd2 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("check()"))[:4]
	msgCallCheck := core.NewMessage(Alice, &contractAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	// // Put the messages into the sequence and run it in sequence.
	// testEu.eu.Api().StateCache().(*cache.StateCache).Clear()
	seq = new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2, 3}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2, &msgCallCheck), slice.New(3, [32]byte{}))
	eu.ExecuteSequence(seq, testEu.config, api, 0)
}

func TestGeneration(t *testing.T) {
	// ================================== Compile the 1st contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "native")
	codeNativeStorage, err := compiler.CompileContracts(targetPath, "/NativeStorage.sol", "0.8.19", "NativeStorage", true)
	if err != nil || len(codeNativeStorage) == 0 {
		t.Fatal("Error: Failed to generate the byte code")
	}
	deployNativeStorageMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(codeNativeStorage), nil, false)

	// ================================== Compile the 2nd contract ==================================
	targetPath = path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "native")
	codeSequential, err := compiler.CompileContracts(targetPath, "/Sequential.sol", "0.8.19", "SequentialTest", true)
	if err != nil || len(codeSequential) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}
	deploySequentialMsg := core.NewMessage(Alice, nil, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(codeSequential), nil, false)

	// ================================== contract Deployment  ==================================
	testEu := NewTestEU(Coinbase, Alice, Bob)
	api := testEu.eu.Api()
	_0thSeq := new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2, 3}, slice.ToSlice(&deployNativeStorageMsg, &deploySequentialMsg), slice.New(3, [32]byte{}))
	// _, seqUniv := _0thseq.ExecuteSequence(testEu.config, api, 0)
	gen := workload.NewGeneration(0, 2, []*workload.JobSequence{_0thSeq})
	eu.ExecuteGeneration(gen, 8, testEu.config, api)

	tests.FlushToStore(testEu.store.(*statestore.StateStore))

	// ================================== 1st contract Call  ==================================
	contractNativeStorageAddr := _0thSeq.Jobs[0].Result.Receipt.ContractAddress
	msgNativeCall := core.NewMessage(Alice, &contractNativeStorageAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("call()"))[:4], nil, false)
	msgNativeCheck := core.NewMessage(Alice, &contractNativeStorageAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check()"))[:4], nil, false)
	nativeSeq := new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2}, slice.ToSlice(&msgNativeCall, &msgNativeCheck), slice.New(2, [32]byte{}))

	// ================================== 2nd contract Call  ==================================
	contractSequentialAddr := _0thSeq.Jobs[1].Result.Receipt.ContractAddress
	msgSequentialAdd := core.NewMessage(Alice, &contractSequentialAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("add()"))[:4], nil, false)
	msgSequentialCheck := core.NewMessage(Alice, &contractSequentialAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check()"))[:4], nil, false)
	sequentialSeq := new(workload.JobSequence).FromEthMessages(2, []uint64{3, 4}, slice.ToSlice(&msgSequentialAdd, &msgSequentialCheck), slice.New(2, [32]byte{}))

	_1stGen := workload.NewGeneration(0, 2, []*workload.JobSequence{nativeSeq, sequentialSeq})
	clearTransitions := eu.ExecuteGeneration(_1stGen, 8, testEu.config, api) // Export transitions

	// // ================================== Commit to DB  ==================================
	acctTrans := statecell.StateCells(clearTransitions).To(statecell.IPTransition{})
	testEu.eu.Api().StateCache().(*cache.StateCache).Insert(acctTrans)

	msgNativeCheck2 := core.NewMessage(Alice, &contractNativeStorageAddr, 3, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check2()"))[:4], nil, false)
	// msgSequentialCheck2 := core.NewMessage(Alice, &contractSequentialAddr, 4, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check2()"))[:4], nil, false)

	seq := new(workload.JobSequence).FromEthMessages(1, []uint64{1}, slice.ToSlice(&msgNativeCheck2), slice.New(1, [32]byte{}))
	_2ndGen := workload.NewGeneration(0, 2, []*workload.JobSequence{seq})
	eu.ExecuteGeneration(_2ndGen, 8, testEu.config, testEu.eu.Api())

	// Add again
	addMsg := core.NewMessage(Alice, &contractNativeStorageAddr, 4, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("call2()"))[:4], nil, false)
	seq = new(workload.JobSequence).FromEthMessages(1, []uint64{1}, slice.ToSlice(&addMsg), slice.New(1, [32]byte{}))

	newGen := workload.NewGeneration(0, 2, []*workload.JobSequence{seq})
	clearTransitions = eu.ExecuteGeneration(newGen, 8, testEu.config, testEu.eu.Api())
	acctTrans = statecell.StateCells(clearTransitions).To(statecell.IPTransition{})

	testEu.eu.Api().StateCache().(*cache.StateCache).Clear().Insert(acctTrans)

	checkMsg := core.NewMessage(Alice, &contractNativeStorageAddr, 5, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), crypto.Keccak256([]byte("check3()"))[:4], nil, false)
	seq = new(workload.JobSequence).FromEthMessages(1, []uint64{1}, slice.ToSlice(&checkMsg), slice.New(1, [32]byte{}))

	nextGen := workload.NewGeneration(0, 2, []*workload.JobSequence{seq})
	eu.ExecuteGeneration(nextGen, 8, testEu.config, testEu.eu.Api())
}

func TestMultiCummutiaves(t *testing.T) {
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "lib", "commutative")
	code, err := compiler.CompileContracts(targetPath, "/u256Cum_test.sol", "0.8.19", "MultiCummutative", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== contract Deploymehnt message ==================================
	deployMsg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	testEu := NewTestEU(Coinbase, Alice, Bob)

	api := testEu.eu.Api()
	jobSeq := new(workload.JobSequence).FromEthMessages(1, []uint64{1}, []*evmcore.Message{&deployMsg}, slice.New(1, [32]byte{}))
	eu.ExecuteSequence(jobSeq, testEu.config, testEu.eu.Api(), 0)
	contractAddr := jobSeq.Jobs[0].Result.Receipt.ContractAddress

	// Move all the transitions from local write cache to the global write cache, so they can be inserted.
	wcache := api.StateCache().(*cache.StateCache)
	testEu.store.(*statestore.StateStore).StateCache.Insert(wcache.Export())
	tests.FlushGeneration(testEu.store.(*statestore.StateStore))
	testEu.store.(*statestore.StateStore).Commit(10)
	testEu.store.(*statestore.StateStore).Clear()

	// Prepare the messages for the contract calls
	data := crypto.Keccak256([]byte("add1()"))[:4]
	msgCallAdd1 := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	data = crypto.Keccak256([]byte("add2()"))[:4]
	msgCallAdd2 := core.NewMessage(Alice, &contractAddr, 2, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	// // Put the messages into the sequence and run it in sequence.
	testEu.eu.Api().StateCache().(*cache.StateCache).Clear()
	jobSeq = new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2}, slice.ToSlice(&msgCallAdd1, &msgCallAdd2), slice.New(2, [32]byte{}))
	eu.ExecuteSequence(jobSeq, testEu.config, api, 0)

	if jobSeq.Jobs[0].Result.Receipt.Status != 1 || jobSeq.Jobs[1].Result.Receipt.Status != 1 {
		t.Error("Error: Failed to call")
	}

	// Move all the transitions from local write cache to the global write cache, so they can be inserted.
	wcache = api.StateCache().(*cache.StateCache)
	testEu.store.(*statestore.StateStore).StateCache.Insert(wcache.Export())
	tests.FlushGeneration(testEu.store.(*statestore.StateStore))

	data = crypto.Keccak256([]byte("check()"))[:4]
	msgCallCheck := core.NewMessage(Alice, &contractAddr, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)

	testEu.eu.Api().StateCache().(*cache.StateCache).Clear()
	jobSeq = new(workload.JobSequence).FromEthMessages(1, []uint64{1, 2}, slice.ToSlice(&msgCallCheck), slice.New(1, [32]byte{}))
	eu.ExecuteSequence(jobSeq, testEu.config, api, 0)

	if jobSeq.Jobs[0].Result.Receipt.Status != 1 {
		t.Error("Error: Failed to call")
	}
}
