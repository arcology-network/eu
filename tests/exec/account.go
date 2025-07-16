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
package exectest

import (
	"strconv"

	eucommon "github.com/arcology-network/eu/common"
	stgcommon "github.com/arcology-network/storage-committer/common"
	stgcomm "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/type/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Addresses used in tests.
var (
	Coinbase = ethcommon.BytesToAddress([]byte("coinbase"))
	Owner    = ethcommon.BytesToAddress([]byte("owner"))
	Alice    = ethcommon.BytesToAddress([]byte("user1"))
	Bob      = ethcommon.BytesToAddress([]byte("user2"))

	Abby    = ethcommon.BytesToAddress([]byte("Abby"))
	Abu     = ethcommon.BytesToAddress([]byte("Abu"))
	Andy    = ethcommon.BytesToAddress([]byte("Andy"))
	Anna    = ethcommon.BytesToAddress([]byte("Anna"))
	Antonio = ethcommon.BytesToAddress([]byte("Antonio"))
	Bailey  = ethcommon.BytesToAddress([]byte("Bailey"))
	Baloo   = ethcommon.BytesToAddress([]byte("Baloo"))
	Bambi   = ethcommon.BytesToAddress([]byte("Bambi"))
	Banza   = ethcommon.BytesToAddress([]byte("Banza"))
	Beast   = ethcommon.BytesToAddress([]byte("Beast"))
)

func GenRandomAccounts(num int) []ethcommon.Address {
	accounts := make([]ethcommon.Address, num)
	for i := 0; i < num; i++ {
		accounts[i] = ethcommon.BytesToAddress([]byte(strconv.Itoa(i)))
	}
	return accounts
}

type TestEu struct {
	eu          *eucommon.EU
	config      *eucommon.Config
	store       stgcommon.ReadOnlyStore
	committer   *stgcomm.StateCommitter
	transitions []*univalue.Univalue
}
