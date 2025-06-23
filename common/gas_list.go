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

package common

import "github.com/ethereum/go-ethereum/params"

const (
	CONTAINER_GAS_READ    = int64(params.SloadGasEIP2200 * 2) // 800 * 2 = 1600, default gas for reading a container.
	GAS_READ              = int64(params.SloadGasEIP2200 / 2) // 800 / 2 = 400
	GAS_WRITE             = int64(params.SstoreSetGas / 2)    // 20,000 / 2 = 10,000
	GAS_DELTA_WRITE       = int64(params.SstoreSetGas / 2)    // 20,000 / 2 = 10,000
	GAS_UNCOMMITTED_RESET = int64(params.SstoreSetGas / 4)    // 20,000 / 4 = 5,000
	GAS_COMMITTED_SET     = int64(params.SstoreSetGas / 8)    // 20,000 / 8 = 2,500

	// GAS_SPONSOR_GAS       = int64(5000)  // Transfer gas another account
	// GAS_USE_SPONSORED_GAS = int64(20000) // Apply sponsored gas to the current transaction's gas
	// GAS_GET_SPONSORED_GAS = int64(50000) // Read total sponsored gas, more expensive because of it is conflict prone.
	GAS_CALL_UNKNOW = int64(1000) // Call an unknown function, which is not defined in the contract, but the gas is still charged.
	GAS_PID         = int64(1000) // The gas for the PID, which is used to identify the transaction in the storage.
	GAS_UUID        = int64(1000)
	GAS_SET_EXEC    = int64(1000)
	GAS_DECODE      = int64(1000)

	GAS_DEBUG_PRINT = int64(1000)
	GAS_DEFER       = int64(10000)

	GAS_NEW_CONTAINER = int64(10000)
)
