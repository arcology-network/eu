package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestParallelBasic(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaNativeAssignmentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParallelWithConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaFixedLengthWithConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaFixedLengthWithConflictAndRollback(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaFixedLengthWithConflictRollbackTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiGlobalParaSingleInUse(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiGlobalParaSingleInUse", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiGlobalParaTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiprocessConcurrentBool", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiLocalPara(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiTempParaTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaMultiWithClear(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiLocalParaTestWithClear", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiParaCumulativeU256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiParaCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiCumulativeU256ConcurrentOperation(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiCumulativeU256ConcurrentOperation", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParallelizerArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelizerArrayTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultipleParallelArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiParaCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRecursiveParallelizerOnNativeArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "RecursiveParallelizerOnNativeArrayTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRecursiveParallelizerOnContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "RecursiveParallelizerOnContainerTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxSelfRecursiveDepth4Test(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxSelfRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxSelfRecursiveDepth(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxRecursiveDepthOffLimits(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepthOffLimitTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Recursive(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")
	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MixedRecursiveMultiprocessTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case1(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call1()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256ThreadingMultiRuns(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ThreadingCumulativeU256SameMpMulti", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ParaCompute(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParaCompute", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestNativeStorageAssignment(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "NativeStorageAssignmentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaConflictDifferentContracts(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaRwConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaRwConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaSubbranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaSubbranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParentChildBranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParentChildBranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

// func TestParaTransfer(t *testing.T) {
// 	currentPath, _ := os.Getwd()
// 	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/apps/")

// 	eu, contractAddress, ccurl, err := AliceDeploy(targetPath, contractFile, version, contractName)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if len(funcName) == 0 {
// 		return
// 	}

// 	abi.Encode(contractAddress)
// 	AliceCall(eu, *contractAddress, funcName, ccurl), eu, nil

// 	// err, _, _ := DeployThenInvoke(targetPath, "transfer/transfer_test.sol", "0.8.19", "Transfer", "call()", []byte{}, false)
// 	// if err != nil {
// 	// 	t.Error(err)
// 	// }
// }

// func TestParaTransfer(t *testing.T) {
// 	currentPath, _ := os.Getwd()
// 	targetPath := path.Join((path.Dir(filepath.Dir(currentPath))), "concurrentlib/")

// 	// Deploy coin contract
// 	err, eu, receipt := DeployThenInvoke(targetPath, "apps/transfer/transfer_test.sol", "0.8.19", "ParaTransfer", "", []byte{}, false)
// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// 	coinAddress := receipt.ContractAddress

// 	// Deploy the caller contrat
// 	callerCode, err := compiler.CompileContracts(targetPath, "apps/transfer/transfer_test.sol", "0.8.19", "ParaTransferTestCaller", false)
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

// may have to do with the callContracts()

// set nonce didn't write to the DB / cache
