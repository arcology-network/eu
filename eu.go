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

package execution

import (
	"bytes"
	"fmt"
	"math"
	"math/big"

	commontype "github.com/arcology-network/common-lib/types"
	eth "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/types"
	evmcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
)

type EU struct {
	StdMsg      *commontype.StandardMessage
	evm         *vm.EVM           // Original ETH EVM
	statedb     vm.StateDB        // Arcology Implementation of Eth StateDB
	api         intf.EthApiRouter // Arcology API calls
	ChainConfig *params.ChainConfig
	VmConfig    vm.Config
}

func NewEU(chainConfig *params.ChainConfig, vmConfig vm.Config, statedb vm.StateDB, api intf.EthApiRouter) *EU {
	eu := &EU{
		ChainConfig: chainConfig,
		VmConfig:    vmConfig,
		evm:         vm.NewEVM(vm.BlockContext{BlockNumber: new(big.Int).SetUint64(100000000)}, vm.TxContext{}, statedb, chainConfig, vmConfig),
		statedb:     statedb,
		api:         api,
	}

	eu.api.SetEU(eu)
	eu.evm.ArcologyNetworkAPIs.APIs = api
	return eu
}

func (this *EU) ID() uint32         { return uint32(this.StdMsg.ID) }
func (this *EU) TxHash() [32]byte   { return this.StdMsg.TxHash }
func (this *EU) GasPrice() *big.Int { return this.StdMsg.Native.GasPrice }
func (this *EU) Coinbase() [20]byte { return this.evm.Context.Coinbase }
func (this *EU) Origin() [20]byte   { return this.evm.TxContext.Origin }

func (this *EU) Message() interface{}            { return this.StdMsg }
func (this *EU) VM() interface{}                 { return this.evm }
func (this *EU) Statedb() vm.StateDB             { return this.statedb }
func (this *EU) Api() intf.EthApiRouter          { return this.api }
func (this *EU) SetApi(newApi intf.EthApiRouter) { this.api = newApi }

func (this *EU) SetRuntimeContext(statedb vm.StateDB, api intf.EthApiRouter) {
	this.api = api
	this.statedb = statedb

	this.evm.StateDB = this.statedb
	this.evm.ArcologyNetworkAPIs.APIs = api
}

func (this *EU) Run(stdmsg *commontype.StandardMessage, blockContext vm.BlockContext, txContext vm.TxContext) (*evmcoretypes.Receipt, *evmcore.ExecutionResult, error) {
	this.statedb.(*eth.ImplStateDB).PrepareFormer(stdmsg.TxHash, ethcommon.Hash{}, uint32(stdmsg.ID))

	this.evm.Context = blockContext
	this.evm.TxContext = txContext
	this.StdMsg = stdmsg

	gasPool := core.GasPool(math.MaxUint64)
	result, err := core.ApplyMessage(this.evm, this.StdMsg.Native, &gasPool) // Execute the transcation

	if err != nil {
		result = &core.ExecutionResult{
			Err: err,
		}
	}

	assertLog := GetAssertion(result.ReturnData) // Get the assertion
	if len(assertLog) > 0 {
		this.api.AddLog("assert", assertLog)
	}

	// Create a new receipt
	receipt := types.NewReceipt(nil, result.Failed(), result.UsedGas)
	receipt.TxHash = stdmsg.TxHash
	receipt.GasUsed = result.UsedGas

	// Check the newly created address
	if stdmsg.Native.To == nil {
		userSpecifiedAddress := crypto.CreateAddress(this.evm.Origin, stdmsg.Native.Nonce)
		receipt.ContractAddress = result.ContractAddress
		if !bytes.Equal(userSpecifiedAddress.Bytes(), result.ContractAddress.Bytes()) {
			this.api.AddLog("ContractAddressWarning", fmt.Sprintf("user specified address = %v, inner address = %v", userSpecifiedAddress, result.ContractAddress))
		}
	}
	receipt.Logs = this.statedb.(*eth.ImplStateDB).GetLogs(stdmsg.TxHash)
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	return receipt, result, err
}

// Get the assertion info from the execution result
func GetAssertion(ret []byte) string {
	offset := 4 + 32 + 32
	pattern := []byte{8, 195, 121, 160}
	if ret != nil && len(ret) > offset {
		starts := ret[:4]
		if string(pattern) == string(starts) {
			remains := ret[offset:]
			end := 0
			for i := range remains {
				if remains[i] == 0 {
					end = i
					break
				}
			}
			return string(remains[:end])
		}
	}
	return ""
}
