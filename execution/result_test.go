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
	"encoding/hex"
	"math/big"
	"testing"

	eucommon "github.com/arcology-network/common-lib/types"
	intf "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
	ethcore "github.com/ethereum/go-ethereum/core"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

func TestResultPostprocessor(t *testing.T) {
	sender := [20]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	other := [20]byte{10, 9, 8, 7, 6, 5, 4, 3, 2, 1}
	coinbase := [20]byte{11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	results := Result{
		From:     sender,
		Coinbase: coinbase,
		immuned:  []*univalue.Univalue{},
		RawStateAccesses: []*univalue.Univalue{
			// sender transfer -> coinbase 50
			// sender gas fee -> Coinbase 100
			// Other transfer -> Coinbase 50
			univalue.NewUnivalue(0, "blcc:/"+hex.EncodeToString(sender[:])+"/nonce", 0, 0, 0, commutative.NewUnboundedUint64(), nil),
			univalue.NewUnivalue(0, "blcc:/"+hex.EncodeToString(sender[:])+"/balance", 0, 0, 0, commutative.NewU256Delta(uint256.NewInt(150), false), nil),
			univalue.NewUnivalue(0, "blcc:/"+hex.EncodeToString(coinbase[:])+"/balance", 0, 0, 0, commutative.NewU256Delta(uint256.NewInt(200), true), nil),
			univalue.NewUnivalue(0, "blcc:/"+hex.EncodeToString(other[:])+"/random", 0, 0, 0, noncommutative.NewString("Random"), nil),
			univalue.NewUnivalue(0, "blcc:/"+hex.EncodeToString(other[:])+"/balance", 0, 0, 0, commutative.NewU256Delta(uint256.NewInt(50), false), nil),
		},
		StdMsg: &eucommon.StandardMessage{
			Native: &ethcore.Message{
				GasPrice: big.NewInt(1),
			},
		},

		Receipt: &ethcoretypes.Receipt{GasUsed: uint64(100)},
		// Err:     errors.New("Error msg"),
	}

	if len(results.RawStateAccesses) != 5 {
		t.Errorf("Postprocess failed, expecting 5, got %d", len(results.RawStateAccesses))
	}
	results.Postprocess()

	if len(results.RawStateAccesses)+len(results.immuned) != 8 {
		t.Errorf("Postprocess failed, expecting 7, got %d", len(results.RawStateAccesses)+len(results.immuned))
	}

	if v := results.RawStateAccesses[2].Value().(intf.Type).Delta().(uint256.Int); (&v).Uint64() != 200 && results.RawStateAccesses[2].Value().(intf.Type).DeltaSign() {
		t.Errorf("Postprocess failed, expecting 100, got %d", v)
	}

	// Sender pay gas fee -100.
	if v := results.immuned[0].Value().(intf.Type).Delta().(uint256.Int); (&v).Uint64() != 100 && !results.immuned[1].Value().(intf.Type).DeltaSign() {
		t.Errorf("Postprocess failed, expecting 50, got %d", v)
	}

	// Coinbase gas fee + 100.
	if v := results.immuned[1].Value().(intf.Type).Delta().(uint256.Int); (&v).Uint64() != 100 && results.immuned[1].Value().(intf.Type).DeltaSign() {
		t.Errorf("Postprocess failed, expecting 50, got %d", v)
	}

	// Sender transfers -50.
	if v := results.RawStateAccesses[1].Value().(intf.Type).Delta().(uint256.Int); (&v).Uint64() != 50 && !results.immuned[1].Value().(intf.Type).DeltaSign() {
		t.Errorf("Postprocess failed, expecting 50, got %d", v)
	}
}
