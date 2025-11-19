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
	"testing"
)

func TestContractTransfer(t *testing.T) {
	// currentPath, _ := os.Getwd()
	// targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/")
	// _, err, _, _ := DeployThenInvoke(targetPath, "examples/eth/TransferTest.sol", "0.8.19", "TransferTest", "transferToContract()", []byte{}, false, 0)
	// if err != nil {
	// 	t.Error(err.Error())
	// }

	// if !commonlibcommon.FileExists(filepath.Join(targetPath, "examples/eth/TransferTest.sol")) {
	// 	t.Error("Error: The contract is not found!!!")
	// }

	// eu, contractAddress, db, _, err := AliceDeploy(targetPath, "examples/eth/TransferTest.sol", "0.8.19", "transferToContract()")
	// if err != nil {
	// 	t.Error("Error: The contract is not found!!!")
	// }

	// _, err = AliceCall(eu, *contractAddress, "examples/eth/TransferTest.sol", db, 11)
	// if err != nil {
	// 	t.Error("Error: The contract is not found!!!")
	// }
}
