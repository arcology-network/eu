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

import (
	"github.com/ethereum/go-ethereum/params"
)

const (
	GAS_GET_CONTAINER_META = int64(params.SloadGasEIP2200 * 2) // 800 * 2 = 1600, default gas for reading a container.
	GAS_READ               = int64(params.SloadGasEIP2200 / 2) // 800 / 2 = 400
	GAS_WRITE              = int64(params.SstoreSetGas / 2)    // 20,000 / 2 = 10,000
	GAS_DELTA_WRITE        = int64(params.SstoreSetGas / 2)    // 20,000 / 2 = 10,000
	GAS_UNCOMMITTED_RESET  = int64(params.SstoreSetGas / 4)    // 20,000 / 4 = 5,000
	GAS_COMMITTED_SET      = int64(params.SstoreSetGas / 8)    // 20,000 / 8 = 2,500

	GAS_CALL_UNKNOW      = int64(1000) // Call an unknown function, which is not defined in the contract, but the gas is still charged.
	GAS_CALL_API         = int64(1000) // The gas to pay for calling an API function, regardless if the result is successful or not.
	GAS_ENCODE           = int64(2000)
	GAS_DECODE           = int64(1000)
	GAS_GET_RUNTIME_INFO = int64(1000) // Get the information of the current  transaction, such as the gas, the pid, the uuid, etc.
	GAS_SET_RUNTIME_INFO = int64(2000)
	GAS_DEFER            = int64(10000)

	GAS_NEW_CONTAINER  = int64(10000)
	GAS_CONTAINER_META = int64(1000)

	DATA_UNIT_SIZE      = uint64(32)
	DATA_MIN_READ_SIZE  = DATA_UNIT_SIZE
	DATA_MIN_WRITE_SIZE = DATA_UNIT_SIZE
)
