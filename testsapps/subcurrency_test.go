package tests

// func TestSubcurrencyMint(t *testing.T) {
// 	currentPath, _ := os.Getwd()
// 	targetPath := path.Join((path.Dir(filepath.Dir(currentPath))), "concurrentlib/")

// 	// Deploy coin contract
// 	err, eu, receipt := tests.DeployThenInvoke(targetPath, "examples/subcurrency/Subcurrency.sol", "0.8.19", "Coin", "", []byte{}, false)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	coinAddress := receipt.ContractAddress

// 	// Deploy the caller contrat
// 	callerCode, err := compiler.CompileContracts(targetPath, "examples/subcurrency/subcurrency_test.sol", "0.8.19", "SubcurrencyCaller", false)
// 	if err != nil || len(callerCode) == 0 {
// 		t.Error(err)
// 	}

// 	config := tests.MainTestConfig()
// 	config.Coinbase = &adaptorcommon.Coinbase
// 	config.BlockNumber = new(big.Int).SetUint64(10000000)
// 	config.Time = new(big.Int).SetUint64(10000000)
// 	err, config, eu, receipt = tests.DepolyContract(eu, config, callerCode, "", []byte{}, 2, false)
// 	if err != nil || receipt.Status != 1 {
// 		t.Error(err)
// 	}

// 	addr := codec.Bytes32{}.Decode(common.PadLeft(coinAddress[:], 0, 32)).(codec.Bytes32) // Callee contract address
// 	funCall := crypto.Keccak256([]byte("call(address)"))[:4]
// 	funCall = append(funCall, addr[:]...)

// 	var execResult *evmcore.ExecutionResult
// 	err, eu, execResult, receipt = tests.CallContract(eu, receipt.ContractAddress, funCall, 0, false)
// 	if receipt.Status != 1 {
// 		t.Error(execResult.Err)
// 	}
// }
