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

func TestContractU256Cum(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "array/U256Cum_test.sol", "0.8.19", "U256CumArrayTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestContractU256CumMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	// _, err, _, _ := DeployThenInvoke(targetPath, "map/AddressU256Cum_test.sol", "0.8.19", "AddressU256MapTest", "", []byte{}, false)
	// if err != nil {
	// 	t.Error(err)
	// }

	_, err2, _, _ := DeployThenInvoke(targetPath, "map/AddressU256Cum_test.sol", "0.8.19", "AddressU256MapTest2", "call()", []byte{}, false)
	if err2 != nil {
		t.Error(err2)
	}
}
