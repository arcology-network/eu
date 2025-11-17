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
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/native/")
	_, err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "NativeStorage", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}


func TestGasDebitInFailedTx(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/native/")
	_, err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "TestFailed", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
