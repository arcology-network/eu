package common

import (
	"math"
	"math/big"

	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
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
	Chain       adaptorcommon.ChainContext
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
