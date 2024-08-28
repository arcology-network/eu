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

package compiler

import (
	"fmt"
	"os"
	"testing"
)

func TestPythonContractCompiler(t *testing.T) {
	currentPath, _ := os.Getwd()

	fmt.Println(currentPath)

	path := currentPath
	version := "0.5.0"
	solfilename := "compiler_test.sol"
	contractName := "Example"

	bincode, err := CompileContracts(path, solfilename, version, contractName, false)
	if err != nil {
		fmt.Printf("reading contract err:%v\n", err)
		return
	}
	fmt.Printf("bytes:%v\n", bincode)
}

func TestEnsure(t *testing.T) {
	currentPath, _ := os.Getwd()
	ensureOutpath(currentPath)
}
func TestGetSolMeta(t *testing.T) {
	currentPath, _ := os.Getwd()
	solfilename := "compiler_test.sol"
	contractName, err := GetContractMeta(currentPath + "/" + solfilename)
	if err != nil {
		fmt.Printf("Get contract meta err:%v\n", err)
		return
	}
	fmt.Printf("contractName:%v\n", contractName)
}
