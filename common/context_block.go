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
	"math/big"

	intf "github.com/arcology-network/eu/interface"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
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
func CanTransfer(db vm.StateDB, addr common.Address, amount *uint256.Int) bool {
	return db.PeekBalance(addr).Cmp(amount) >= 0
}

// Transfer subtracts amount from sender and adds amount to recipient using the given Db
func Transfer(db vm.StateDB, sender, recipient common.Address, amount *uint256.Int) {
	db.SubBalance(sender, amount)
	db.AddBalance(recipient, amount)
}
