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

package common

import (
	"math"
	"math/big"

	adaptorintf "github.com/arcology-network/eu/interface"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/consensus"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"
)

// DummyChain implements the ChainContext interface.
type DummyChain struct{}

func (chain *DummyChain) GetHeader(evmcommon.Hash, uint64) *types.Header { return &types.Header{} }
func (chain *DummyChain) Engine() consensus.Engine                       { return nil }

// Config contains all the static settings used in Schedule.
type Config struct {
	ChainConfig *params.ChainConfig
	VMConfig    *vm.Config
	BlockNumber *big.Int    // types.Header.Number
	ParentHash  common.Hash // types.Header.ParentHash
	Time        *big.Int    // types.Header.Time
	Chain       adaptorintf.ChainContext
	Coinbase    *evmcommon.Address
	GasLimit    uint64   // types.Header.GasLimit
	Difficulty  *big.Int // types.Header.Difficulty
}

func (this *Config) SetCoinbase(coinbase evmcommon.Address) *Config {
	this.Coinbase = &coinbase
	return this
}

func NewConfig() *Config {
	cfg := &Config{
		ChainConfig: params.MainnetChainConfig,
		VMConfig:    &vm.Config{},
		BlockNumber: big.NewInt(0),
		ParentHash:  evmcommon.Hash{},
		Time:        big.NewInt(0),
		Coinbase:    &evmcommon.Address{},
		GasLimit:    math.MaxUint64,
		Difficulty:  big.NewInt(0),
	}
	cfg.Chain = new(DummyChain)
	return cfg
}

func NewConfigFromBlockContext(context vm.BlockContext) *Config {
	cfg := &Config{
		ChainConfig: params.MainnetChainConfig,
		VMConfig:    &vm.Config{},
		BlockNumber: context.BlockNumber,
		ParentHash:  evmcommon.Hash{},
		Time:        big.NewInt(int64(context.Time)),
		Coinbase:    &context.Coinbase,
		GasLimit:    context.GasLimit,
		Difficulty:  context.Difficulty,
	}
	cfg.Chain = new(DummyChain)
	return cfg
}
