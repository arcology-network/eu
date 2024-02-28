package execution

import (
	"math/big"

	intf "github.com/arcology-network/evm-adaptor/interface"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
)

func NewEVMBlockContext(cfg *Config) vm.BlockContext {
	return vm.BlockContext{
		CanTransfer: CanTransfer,
		Transfer:    Transfer,
		GetHash:     GetHashFn(cfg.BlockNumber, cfg.ParentHash, cfg.Chain),

		Coinbase:    *cfg.Coinbase,
		GasLimit:    cfg.GasLimit,
		BlockNumber: new(big.Int).Set(cfg.BlockNumber),
		Time:        cfg.Time.Uint64(),
		Difficulty:  new(big.Int).Set(cfg.Difficulty),
	}
}

// GetHashFn returns a GetHashFunc which retrieves header hashes by number
func GetHashFn(blockNumber *big.Int, parentHash common.Hash, chain intf.ChainContext) func(n uint64) common.Hash {
	return func(n uint64) common.Hash { return common.Hash{} }
}

// CanTransfer checks whether there are enough funds in the address' account to make a transfer.
// This does not take the necessary gas in to account to make the transfer valid.
func CanTransfer(db vm.StateDB, addr common.Address, amount *big.Int) bool {
	return db.PeekBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *big.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
