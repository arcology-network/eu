/*
 *   Copyright (c) 2025 Arcology Network

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

package multiprocessor

import (
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

// To
type CallResult struct {
	Success    bool
	ReturnData []byte
}

func EncodeCallReturns(returnData [][]byte, flags []bool) ([]byte, error) {
	tupleArrayType, err := abi.NewType("tuple[]", "",
		[]abi.ArgumentMarshaling{
			{Name: "success", Type: "bool"},
			{Name: "returnData", Type: "bytes"},
		})
	if err != nil {
		return nil, err
	}

	args := abi.Arguments{
		{Type: tupleArrayType},
	}

	results := make([]CallResult, len(returnData))
	for i := range returnData {
		results[i] = CallResult{
			Success:    flags[i],
			ReturnData: returnData[i],
		}
	}
	return args.Pack(results)
}

func DecodeCallReturns(data []byte) ([]CallResult, error) {
	// Define the type: tuple(bool, bytes)[]
	abiJSON := `[{"name":"getCallResults","type":"function","inputs":[],"outputs":[{"name":"","type":"tuple[]","components":[{"name":"success","type":"bool"},{"name":"returnData","type":"bytes"}]}]}]`

	// Parse the ABI
	parsedABI, err := abi.JSON(strings.NewReader(abiJSON))
	if err != nil {
		return nil, err
	}

	// Prepare the variable to hold the unpacked data
	var results []CallResult

	// Use UnpackIntoInterface with the method name
	err = parsedABI.UnpackIntoInterface(&results, "getCallResults", data)
	if err != nil {
		return nil, err
	}

	return results, nil
}
