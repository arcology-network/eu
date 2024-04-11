// KernelAPI provides system level function calls supported by arcology platform.
package tests

import (
	"strconv"

	eu "github.com/arcology-network/eu"
	adaptorcommon "github.com/arcology-network/evm-adaptor/common"
	stgcomm "github.com/arcology-network/storage-committer/committer"
	ccurlintf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
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
	eu          *eu.EU
	config      *adaptorcommon.Config
	store       ccurlintf.Datastore
	committer   *stgcomm.StateCommitter
	transitions []*univalue.Univalue
}
