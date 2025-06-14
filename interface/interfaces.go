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

// KernelAPI provides system level function calls supported by arcology platform.
package interfaces

import (
	"github.com/ethereum/go-ethereum/common"

	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
)

type ApiCallHandler interface {
	Address() [20]byte
	Call([20]byte, [20]byte, []byte, [20]byte, uint64, bool) ([]byte, bool, int64)
}

type ILog interface {
	GetByKey() string
	GetValue() string
}

type ChainContext interface {
	Engine() consensus.Engine                    // Engine retrieves the chain's consensus engine.
	GetHeader(common.Hash, uint64) *types.Header // GetHeader returns the hash corresponding to their hash.
}

type EthApiRouter interface {
	// Used in EVM to call the kernel API
	Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, blockhash evmcommon.Hash, isReadOnly bool) (bool, []byte, bool, int64)
	GetExecutionSubsidy() uint64 // Get the execution subsidy for the current call
	SetExecutionSubsidy(uint64)  // Set the execution subsidy for the current call

	// Arcology specific APIs
	GetDeployer() evmcommon.Address
	SetDeployer(evmcommon.Address)

	GetEU() interface{}
	SetEU(interface{})

	GetSchedule() interface{}
	SetSchedule(interface{})

	AuxDict() map[string]interface{}
	WriteCachePool() interface{}
	WriteCache() interface{}
	SetWriteCache(interface{}) EthApiRouter
	New(interface{}, interface{}, evmcommon.Address, interface{}) EthApiRouter
	Cascade() EthApiRouter

	Origin() evmcommon.Address
	Coinbase() evmcommon.Address

	VM() interface{} //*vm.EVM

	CheckRuntimeConstrains() bool

	DecrementDepth() uint8
	Depth() uint8
	AddLog(key, value string)

	GetSerialNum(int) uint64
	Pid() [32]byte
	UUID() []byte
	ElementUID() []byte
}
