package tests

import (
	"math/big"
	"os"
	"path"
	"path/filepath"
	"testing"

	commontypes "github.com/arcology-network/common-lib/types"
	eucommon "github.com/arcology-network/eu/common"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	"github.com/arcology-network/evm-adaptor/compiler"
	stgcomm "github.com/arcology-network/storage-committer/storage/committer"
	cache "github.com/arcology-network/storage-committer/storage/writecache"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestBaseContainer(t *testing.T) {
	// eu, config, db, committer, _ := NewTestEU(Coinbase, Alice, Bob)
	testEu := NewTestEU(Coinbase, Alice, Bob)
	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "lib")

	code, err := compiler.CompileContracts(targetPath, "/base/base_test.sol", "0.8.19", "BaseTest", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	StdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := testEu.eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(testEu.config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()
	_, transitions := cache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

	// msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true) // Build the message
	// receipt, _, err := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()

	//t.Log("\n" + eucommon.FormatTransitions(transitions))
	// t.Log(receipt)

	if receipt.Status != 1 || err != nil {
		t.Error("Error: Deployment failed!!!", err)
	}

	contractAddress := receipt.ContractAddress
	testEu.committer = stgcomm.NewStateCommitter(testEu.store, nil)
	testEu.committer.Import(transitions)
	testEu.committer.Precommit([]uint32{1})
	testEu.committer.Commit(0)

	// ================================== Call() ==================================
	// receipt, _, err = eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg))
	// if err != nil {
	// 	fmt.Print(err)
	// }
	// return

	data := crypto.Keccak256([]byte("call()"))[:4]
	msg = core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	StdMsg = &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err = testEu.eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(testEu.config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions = eu.Api().WriteCacheFilter().ByType()
	_, transitions = cache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

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
