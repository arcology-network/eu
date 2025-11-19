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
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
)

// func TestClearCommitted(t *testing.T) {
// 	currentPath, _ := os.Getwd()
// 	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/crdt/")
// 	_, err, _, _ := DeployThenInvoke(targetPath, "array/clear_committed.t.sol", "0.8.19", "ClearCommittedTest", "clear()", []byte{}, false)
// 	if err != nil {
// 		t.Error(err)
// 	}
// }

func TestAddressContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/address.t.sol", "0.8.19", "AddressTest", "", []byte{}, false)
	if err != nil {
		fmt.Println("Error is:", err.Error())
		t.Error(err)
	}

	_, err, _, _ = DeployThenInvoke(targetPath, "test/crdt/array/address.t.sol", "0.8.19", "AddressTestTransient", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestBoolContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/bool.t.sol", "0.8.19", "BoolTest", "check()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestBytesContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/bytes.t.sol", "0.8.19", "ByteTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractBytes32(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/bytes32.t.sol", "0.8.19", "Bytes32Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractU256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/u256.t.sol", "0.8.19", "U256Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractU256Cum(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/U256Cum.t.sol", "0.8.19", "U256CumArrayTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}

	_, err, _, _ = DeployThenInvoke(targetPath, "test/crdt/array/U256Cum.t.sol", "0.8.19", "U256CumArrayTestTransient", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractInt256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/int256.t.sol", "0.8.19", "Int256Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractString(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/string.t.sol", "0.8.19", "StringTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContainerPair(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")

	_, err, _, _ := DeployThenInvoke(targetPath, "test/crdt/array/bytes_bool.t.sol", "0.8.19", "PairTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
