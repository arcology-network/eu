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

package abi

import (
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func DecodeEth(abiDefinition, encodedData, functionName string, types []interface{}) {
	parsedABI, err := abi.JSON(strings.NewReader(abiDefinition))
	if err != nil {
		log.Fatal("Failed to parse ABI:", err)
	}

	// Check if the function exists in the parsed ABI
	method, exists := parsedABI.Methods[functionName]
	if !exists {
		log.Fatalf("Could not locate named method: %s", functionName)
	}

	// Convert hex string to bytes
	dataBytes, err := hexutil.Decode(encodedData)
	if err != nil {
		log.Fatal("Failed to decode hex data:", err)
	}

	// Decode the ABI-encoded data using the method reference
	decoded, err := method.Inputs.Unpack(dataBytes)
	if err != nil {
		log.Fatalf("Failed to unpack data: %v", err)
	}

	// Assign the decoded values to the types slice
	for i, val := range decoded {
		if i < len(types) {
			switch v := types[i].(type) {
			case *big.Int:
				*v = *((val).(*big.Int))
			case *[]byte:
				*v = val.([]byte)
			default:
				log.Printf("Unsupported type for index %d: %T", i, val)
			}
		}
	}
}

func DecodeInt256(bytes []byte) (*big.Int, error) {
	intVal := new(big.Int).SetBytes(bytes) // Convert bytes to a big.Int

	// If the value is negative in two's complement, adjust it
	if bytes[0]&0x80 != 0 { // If the highest bit is set, it means it's a negative number
		maxValue := new(big.Int).Lsh(big.NewInt(1), 256) // 2^256
		intVal = intVal.Sub(intVal, maxValue)
	}
	return intVal, nil
}
