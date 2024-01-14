// KernelAPI provides system level function calls supported by arcology platform.
package tests

import (
	evmcommon "github.com/ethereum/go-ethereum/common"
)

// Addresses used in tests.
var (
	Coinbase = evmcommon.BytesToAddress([]byte("coinbase"))
	Owner    = evmcommon.BytesToAddress([]byte("owner"))
	Alice    = evmcommon.BytesToAddress([]byte("user1"))
	Bob      = evmcommon.BytesToAddress([]byte("user2"))

	Abby    = evmcommon.BytesToAddress([]byte("Abby"))
	Abu     = evmcommon.BytesToAddress([]byte("Abu"))
	Andy    = evmcommon.BytesToAddress([]byte("Andy"))
	Anna    = evmcommon.BytesToAddress([]byte("Anna"))
	Antonio = evmcommon.BytesToAddress([]byte("Antonio"))
	Bailey  = evmcommon.BytesToAddress([]byte("Bailey"))
	Baloo   = evmcommon.BytesToAddress([]byte("Baloo"))
	Bambi   = evmcommon.BytesToAddress([]byte("Bambi"))
	Banza   = evmcommon.BytesToAddress([]byte("Banza"))
	Beast   = evmcommon.BytesToAddress([]byte("Beast"))
)
