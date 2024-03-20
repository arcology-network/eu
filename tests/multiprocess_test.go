package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestU256CumulativeParallelInitTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256CumulativeParallelInitTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ParallelInitTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParallelInitTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

// //434343
func TestU256ParallelPop(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParallelPopTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ParallelPushPopGet(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParallelTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParallelBasic(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaNativeAssignmentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParallelWithConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaFixedLengthWithConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaFixedLengthWithConflictAndRollback(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaFixedLengthWithConflictRollbackTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiGlobalParaSingleInUse(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiGlobalParaSingleInUse", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiGlobalParaTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiprocessConcurrentBool", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiLocalPara(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiTempParaTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaMultiWithClear(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiLocalParaTestWithClear", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

// func TestParaVote(t *testing.T) {
func TestMultiParaU256Conflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParallelConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiParaU256Array(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ArrayOfU256ParallelTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiParaCumulativeU256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiParaCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultiCumulativeU256ConcurrentOperation(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiCumulativeU256ConcurrentOperation", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParallelizerArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelizerArrayTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMultipleParallelArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MultiParaCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRecursiveParallelizerOnNativeArray(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "RecursiveParallelizerOnNativeArrayTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRecursiveAssigner(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "RecursiveAssignerTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestRecursiveParallelizerOnContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "RecursiveParallelizerOnContainerTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxSelfRecursiveDepth4Test(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxSelfRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxSelfRecursiveDepth(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxRecursiveDepthOffLimits(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepthOffLimitTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Recursive(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MixedRecursiveMultiprocessTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case1(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call1()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256ThreadingMultiRuns(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ThreadingCumulativeU256SameMpMulti", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ParaCompute(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParaCompute", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestNativeStorageAssignment(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "NativeStorageAssignmentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaConflictDifferentContracts(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaRwConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaRwConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaSubbranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaSubbranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParentChildBranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParentChildBranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestSimpleConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "SimpleConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaCumU256Sub(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaCumU256SubTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
