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
	execution "github.com/arcology-network/eu"
	"github.com/arcology-network/eu/cache"
	"github.com/arcology-network/evm-adaptor/compiler"
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

	seq.ApiRouter.WriteCache().(*cache.WriteCache).FlushToStore(testEu.store)
	// acctTrans := univalue.Univalues(slice.Clone(accesses)).To(importer.IPTransition{})
	// committer := stgcomm.NewStorageCommitter(testEu.store)
	// committer.Import(acctTrans).Precommit([]uint32{1})
	// committer.Commit(0)
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
