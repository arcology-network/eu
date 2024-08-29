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

func TestAddressContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/address_test.sol", "0.8.19", "AddressTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestBoolContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/bool_test.sol", "0.8.19", "BoolTest", "check()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestBytesContainer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/bytes_test.sol", "0.8.19", "ByteTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractBytes32(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/bytes32_test.sol", "0.8.19", "Bytes32Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractU256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/u256_test.sol", "0.8.19", "U256Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractInt256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/int256_test.sol", "0.8.19", "Int256Test", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractString(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/string_test.sol", "0.8.19", "StringTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
