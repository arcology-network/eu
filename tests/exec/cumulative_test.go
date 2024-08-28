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

func TestCumulativeU256InMapping(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "commutative/u256Cum_test.sol", "0.8.19", "CumulativeU256InMappingTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "commutative/u256Cum_test.sol", "0.8.19", "CumulativeU256Test", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestCumulativeU256Counter(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")
	_, err, _, _ := DeployThenInvoke(targetPath, "commutative/u256Cum_test.sol", "0.8.19", "VisitCounter", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
