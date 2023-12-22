package tests

import (
	"math/big"
	"os"
	"path"
	"path/filepath"
	"testing"

	commontypes "github.com/arcology-network/common-lib/types"
	concurrenturl "github.com/arcology-network/concurrenturl"
	eu "github.com/arcology-network/eu"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
	"github.com/arcology-network/vm-adaptor/compiler"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/crypto"
)

func TestBaseContainer(t *testing.T) {
	executor, config, db, url, _ := NewTestEU()

	// ================================== Compile the contract ==================================
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib", "lib")

	code, err := compiler.CompileContracts(targetPath, "/base/base_test.sol", "0.8.19", "BaseTest", true)
	if err != nil || len(code) == 0 {
		t.Error("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(adaptorcommon.Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	stdMsg := &adaptorcommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := executor.(*eu.EU).Run(stdMsg, eu.NewEVMBlockContext(config), eu.NewEVMTxContext(*stdMsg.Native)) // Execute it
	_, transitions := executor.(*eu.EU).Api().StateFilter().ByType()

	// msg := core.NewMessage(adaptorcommon.Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true) // Build the message
	// receipt, _, err := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, eu.NewEVMBlockContext(config), eu.NewEVMTxContext(msg)) // Execute it
	// _, transitions := eu.Api().StateFilter().ByType()

	//t.Log("\n" + adaptorcommon.FormatTransitions(transitions))
	// t.Log(receipt)

	if receipt.Status != 1 || err != nil {
		t.Error("Error: Deployment failed!!!", err)
	}

	contractAddress := receipt.ContractAddress
	url = concurrenturl.NewConcurrentUrl(db)
	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{1})

	// ================================== Call() ==================================
	// receipt, _, err = eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, eu.NewEVMBlockContext(config), eu.NewEVMTxContext(msg))
	// if err != nil {
	// 	fmt.Print(err)
	// }
	// return

	data := crypto.Keccak256([]byte("call()"))[:4]
	msg = core.NewMessage(adaptorcommon.Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	stdMsg = &adaptorcommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err = executor.(*eu.EU).Run(stdMsg, eu.NewEVMBlockContext(config), eu.NewEVMTxContext(*stdMsg.Native)) // Execute it
	_, transitions = executor.(*eu.EU).Api().StateFilter().ByType()

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
