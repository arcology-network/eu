// KernelAPI provides system level function calls supported by arcology platform.
package tests

import (
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
