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

func TestAddressBooleanMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/addressBoolean_test.sol", "0.8.19", "AddressBooleanMapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestAddressBooleanMapConcurrentTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/addressBoolean_test.sol", "0.8.19", "AddressBooleanMapConcurrentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestAddressUint256Map(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/addressUint256_test.sol", "0.8.19", "AddressU256MapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}

	_, err, _, _ = DeployThenInvoke(targetPath, "map/addressUint256_test.sol", "0.8.19", "AddressU256MapConcurrentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestConcurrentAddressBooleanMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/addressBoolean_test.sol", "0.8.19", "AddressBooleanMapConcurrentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestStringUint256Map(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/stringUint256_test.sol", "0.8.19", "StringUint256MapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256MapTest(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/u256_test.sol", "0.8.19", "U256MapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ConcurrentMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "lib/map/u256_test.sol", "0.8.19", "ConcurrenctU256MapTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestHashUint256Map(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/hashU256Cum_test.sol", "0.8.19", "HashU256MapTest", "resetter()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestAddressUint256CumMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "map/addressU256Cum_test.sol", "0.8.19", "AddressU256CumMapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
