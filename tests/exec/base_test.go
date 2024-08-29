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

	commontype "github.com/arcology-network/common-lib/types"
	adaptorcommon "github.com/arcology-network/eu/common"
	"github.com/arcology-network/eu/compiler"
	stgcommiter "github.com/arcology-network/storage-committer/storage/committer"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestBaseContainer(t *testing.T) {
	// eu, config, db, committer, _ := NewTestEU(Coinbase, Alice, Bob)
	testEu := NewTestEU(Coinbase, Alice, Bob)
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib", "lib")

	code, err := compiler.CompileContracts(targetPath, "/base/base_test.sol", "0.8.19", "BaseTest", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	StdMsg := &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := testEu.eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(testEu.config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()
	_, transitions := tempcache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

	// msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true) // Build the message
	// receipt, _, err := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()

	//t.Log("\n" + commontype.FormatTransitions(transitions))
	// t.Log(receipt)

	if receipt.Status != 1 || err != nil {
		t.Error("Error: Deployment failed!!!", err)
	}

	contractAddress := receipt.ContractAddress
	testEu.committer = stgcommiter.NewStateCommitter(testEu.store, nil)
	testEu.committer.Import(transitions)
	testEu.committer.Precommit([]uint32{1})
	testEu.committer.Commit(20)

	// ================================== Call() ==================================
	// receipt, _, err = eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg))
	// if err != nil {
	// 	fmt.Print(err)
	// }
	// return

	data := crypto.Keccak256([]byte("call()"))[:4]
	msg = core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	StdMsg = &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err = testEu.eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(testEu.config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions = eu.Api().WriteCacheFilter().ByType()
	_, transitions = tempcache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

	if err != nil {
		t.Error(err)
	}

	if execResult != nil && execResult.Err != nil {
		t.Error(execResult.Err)
	}

	if receipt.Status != 1 || err != nil {
		t.Error("Error: Failed to call!!!", err)
	}
}
