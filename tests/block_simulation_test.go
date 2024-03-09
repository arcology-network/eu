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

package tests

import (
	"testing"
)

func TestBlocks(b *testing.T) {
	// acct := GenRandomAccounts(100)
	// eus := make([]*TestEu, len(acct))
	// for i := 0; i < len(eus); i++ {
	// 	eus[i] = NewTestEU(Coinbase, acct...)
	// }

	// // Deploy the contract
	// currentPath, _ := os.Getwd()
	// targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/")
	// _, _, db, _, err := AliceDeploy(targetPath, "examples/vote/parallelVote-Mp_test.sol", "0.8.19", "BallotTest")
	// if err != nil {
	// 	b.Fatal(err)
	// }

	// msgs := make([]*eucommon.StandardMessage, 10000)
	// for i := 0; i < len(msgs); i++ {
	// 	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	// 	msgs[i] = &eucommon.StandardMessage{
	// 		ID:     1,
	// 		TxHash: [32]byte{1, 1, 1},
	// 		Native: &msg, // Build the message
	// 		Source: commontypes.TX_SOURCE_LOCAL,
	// 	}
	// }

}
