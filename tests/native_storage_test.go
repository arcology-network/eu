package tests

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/holiman/uint256"
	sha3 "golang.org/x/crypto/sha3"
)

func TestSlotHash(t *testing.T) {
	_ctrn := uint256.NewInt(2).Bytes32()
	_elem := uint256.NewInt(2).Bytes32()

	hash := sha3.NewLegacyKeccak256()
	hash.Write(append(_elem[:], _ctrn[:]...))
	v := hash.Sum(nil)
	fmt.Println(v)
	fmt.Println("0x" + hex.EncodeToString(v))
}

func TestNativeStorage(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/native/")
	err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "NativeStorage", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestGasDebitInFailedTx(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/native/")
	err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "TestFailed", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

// func TestStorageSlot(t *testing.T) {
// 	eu, config, _, _, _ := NewTestEU()

// 	// ================================== Compile the contract ==================================
// 	currentPath, _ := os.Getwd()
// 	project := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib")
// 	code, err := compiler.CompileContracts(project+"/apps", "/storagenative/local_test.sol", "0.8.19", "LocalTest", false)
// 	if err != nil || len(code) == 0 {
// 		t.Error(err)
// 	}

// 	// ================================== Deploy the contract ==================================
// 	msg := core.NewMessage(adaptorcommon.Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true) // Build the message
// 	receipt, _, err := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, ccEu.NewEVMBlockContext(config), ccEu.NewEVMTxContext(msg))           // Execute it
// 	_, transitions := eu.Api().StateFilter().ByType()

// 	// ---------------

// 	// t.Log("\n" + FormatTransitions(accesses))
// 	//t.Log("\n" + adaptorcommon.FormatTransitions(transitions))
// 	t.Log(receipt)
// 	// contractAddress := receipt.ContractAddress
// 	if receipt.Status != 1 || err != nil {
// 		t.Error("Error: Deployment failed!!!", err)
// 	}
// }

// func TestNativeContractSameBlock(t *testing.T) {
// 	persistentDB := cachedstorage.NewDataStore()
// 	meta := commutative.NewPath()
// 	persistentDB.Inject((&concurrenturl.Platform{}).Eth10Account(), meta)
// 	db := curstorage.NewTransientDB(persistentDB)

// 	url := concurrenturl.NewConcurrentUrl(db)
// 	api := ccapi.NewAPI(url)

// 	statedb := cceueth.NewImplStateDB(api)
// 	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
// 	statedb.CreateAccount(adaptorcommon.Coinbase)
// 	// User 1
// 	statedb.CreateAccount(adaptorcommon.Alice)
// 	statedb.AddBalance(adaptorcommon.Alice, new(big.Int).SetUint64(1e18))
// 	// user2
// 	statedb.CreateAccount(adaptorcommon.Bob)
// 	statedb.AddBalance(adaptorcommon.Bob, new(big.Int).SetUint64(1e18))
// 	// Contract owner
// 	statedb.CreateAccount(adaptorcommon.Owner)
// 	statedb.AddBalance(adaptorcommon.Owner, new(big.Int).SetUint64(1e18))

// 	// ================================== Compile ==================================
// 	currentPath, _ := os.Getwd()
// 	project := filepath.Dir(filepath.Dir(currentPath))
// 	targetPath := project + "/apps/native"

// 	bytecode, err := compiler.CompileContracts(targetPath, "NativeStorage.sol", "0.8.19", "NativeStorage", false)
// 	if err != nil || len(bytecode) == 0 {
// 		t.Error("Error: Failed to generate the byte code")
// 		return
// 	}

// 	// Compile
// 	// ================================ Deploy the contract==================================
// 	_, transitions := url.ExportAll()
// 	eu, config := tests.Prepare(db, 10000000, transitions, []uint32{0})
// 	transitions, receipt, err := tests.Deploy(eu, config, adaptorcommon.Owner, 0, bytecode)
// 	//t.Log("\n" + adaptorcommon.FormatTransitions(transitions))
// 	t.Log(receipt)
// 	address := receipt.ContractAddress
// 	t.Log(address)
// 	if receipt.Status != 1 {
// 		t.Error("Error: Failed to deploy!!!", err)
// 	}

// 	// Increment x by one
// 	if _, _, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "incrementX()"); receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to call incrementX() 1!!!", err)
// 	}

// 	if _, _, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "incrementX()"); receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to call incrementX() 2!!!", err)
// 	}

// 	if _, _, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "incrementX()"); receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to call incrementX() 3!!!", err)
// 	}

// 	encoded, _ := abi.Encode(uint64(102))
// 	if _, _, receipt, err := tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "checkY(uint256)", encoded); receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to check checkY() 1!!!", err)
// 	}

// 	if _, _, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "incrementY()"); receipt.Status != 1 {
// 		t.Error("Error: Failed to call incrementY() 1!!!", err)
// 	}

// 	encoded, _ = abi.Encode(uint64(104))
// 	if _, _, receipt, err := tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "checkY(uint256)", encoded); receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to check checkY() 2!!!", err)
// 	}
// }

// func TestNativeContractAcrossBlocks(t *testing.T) {
// 	persistentDB := cachedstorage.NewDataStore()
// 	meta := commutative.NewPath()
// 	persistentDB.Inject((&concurrenturl.Platform{}).Eth10Account(), meta)
// 	db := curstorage.NewTransientDB(persistentDB)

// 	url := concurrenturl.NewConcurrentUrl(db)
// 	api := ccapi.NewAPI(url)
// 	statedb := cceueth.NewImplStateDB(api)
// 	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
// 	statedb.CreateAccount(adaptorcommon.Coinbase)
// 	// User 1
// 	statedb.CreateAccount(adaptorcommon.Alice)
// 	statedb.AddBalance(adaptorcommon.Alice, new(big.Int).SetUint64(1e18))
// 	// user2
// 	statedb.CreateAccount(adaptorcommon.Bob)
// 	statedb.AddBalance(adaptorcommon.Bob, new(big.Int).SetUint64(1e18))
// 	// Contract owner
// 	statedb.CreateAccount(adaptorcommon.Owner)
// 	statedb.AddBalance(adaptorcommon.Owner, new(big.Int).SetUint64(1e18))

// 	// // ================================== Compile ==================================
// 	currentPath, _ := os.Getwd()
// 	project := filepath.Dir(filepath.Dir(currentPath))
// 	targetPath := project + "/apps/native"

// 	bytecode, err := compiler.CompileContracts(targetPath, "NativeStorage.sol", "0.8.19", "NativeStorage", false)
// 	if err != nil || len(bytecode) == 0 {
// 		t.Error("Error: Failed to generate the byte code")
// 	}
// 	// ================================ Deploy the contract==================================
// 	_, transitions := url.ExportAll()
// 	eu, config := tests.Prepare(db, 10000000, transitions, []uint32{0})
// 	transitions, receipt, err := tests.Deploy(eu, config, adaptorcommon.Owner, 0, bytecode)
// 	//t.Log("\n" + adaptorcommon.FormatTransitions(transitions))
// 	t.Log(receipt)
// 	address := receipt.ContractAddress
// 	t.Log(address)
// 	if receipt.Status != 1 {
// 		t.Error("Error: Failed to deploy!!!", err)
// 	}

// 	eu, config = tests.Prepare(db, 10000001, transitions, []uint32{1})
// 	// encoded, _ := abi.Encode(uint64(2))
// 	_, transitions, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "incrementX()")
// 	//t.Log("\n" + adaptorcommon.FormatTransitions(transitions))
// 	t.Log(receipt)
// 	if receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to call incrementX()!!!", err)
// 	}

// 	eu, config = tests.Prepare(db, 10000001, transitions, []uint32{0})
// 	encodedInput, _ := abi.Encode(uint64(3))
// 	acc, transitions, receipt, err := tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "checkX(uint256)", encodedInput)
// 	t.Log("\n" + adaptorcommon.FormatTransitions(acc))
// 	t.Log(receipt)
// 	if receipt.Status != 1 {
// 		t.Error("Error: Failed to call checkX()!!!", err)
// 	}

// 	eu, config = tests.Prepare(db, 10000001, transitions, []uint32{0})
// 	encodedInput, _ = abi.Encode(uint64(102))
// 	acc, transitions, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "checkY(uint256)", encodedInput)
// 	t.Log("\n" + adaptorcommon.FormatTransitions(acc))
// 	t.Log(receipt)
// 	if receipt.Status != 1 {
// 		t.Error("Error: Failed to call checkY()!!!", err)
// 	}

// 	eu, config = tests.Prepare(db, 10000001, transitions, []uint32{0})
// 	encodedInput, _ = abi.Encode(uint64(3))
// 	acc, _, receipt, err = tests.CallFunc(eu, config, &adaptorcommon.Alice, &address, 0, true, "checkX(uint256)", encodedInput)
// 	t.Log("\n" + adaptorcommon.FormatTransitions(acc))
// 	t.Log(receipt)
// 	if receipt.Status != 1 || err != nil {
// 		t.Error("Error: Failed to call checkX()!!!", err)
// 	}
// }
