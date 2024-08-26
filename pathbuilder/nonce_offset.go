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

package pathbuilder

import (
	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/exp/slice"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func (this *ImplStateDB) CalculateNonceOffset(addr evmcommon.Address, nonce uint64) uint64 {
	if this.api.Origin() == addr || this.api.GetEU() == nil {
		return 0
	}

	id := uint64(this.api.GetEU().(interface{ ID() uint32 }).ID())
	encoded := slice.Flatten([][]byte{
		this.api.GetDeployer().Bytes(),
		codec.Uint64(id).Encode(),
		codec.Uint64(nonce).Encode(),
	})

	return uint64(new(codec.Uint64).Decode(crypto.Keccak256(encoded)[:8]).(codec.Uint64)) >> 16
}
