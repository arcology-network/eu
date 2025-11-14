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
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestMaxSelfRecursiveDepth4Test(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxSelfRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxSelfRecursiveDepth(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepth4Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestMaxRecursiveDepthOffLimits(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MaxRecursiveDepthOffLimitTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Recursive(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "MixedRecursiveMultiprocessTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaCumU256Sub(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaCumU256SubTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Case1(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParallelCumulativeU256", "call1()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256ThreadingMultiRuns(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ThreadingCumulativeU256SameMpMulti", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ParaCompute(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "U256ParaCompute", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestNativeStorageAssignment(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "NativeStorageAssignmentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaConflictDifferentContracts(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaRwConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaRwConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaSubbranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaSubbranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParentChildBranchConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParentChildBranchConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestSimpleConflict(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "SimpleConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaDeletions(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaDeletions", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestParaArrayUint256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "multiprocess/multiprocess_test.sol", "0.8.19", "ParaAddressUint256ConflictTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
