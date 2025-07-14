/*
*   Copyright (c) 2025 Arcology Network

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
package gas

import (
	slice "github.com/arcology-network/common-lib/exp/slice"
)

/*
GasPrepayerLookup is a structure that holds a map of contract addresses to their prepaid gas messages.
It is used to manage the prepaid gas for contracts with deferred execution.
*/
type GasPrepayerLookup struct {
	Payers map[string][]*PrepayerInfo
}

func NewGasPrepayerLookup(payerInfo []*PrepayerInfo) *GasPrepayerLookup {
	lookup := &GasPrepayerLookup{
		Payers: make(map[string][]*PrepayerInfo),
	}
	return lookup.AddPrepayer(payerInfo)
}

func (this *GasPrepayerLookup) AddPrepayer(payerInfo []*PrepayerInfo) *GasPrepayerLookup {
	for _, info := range payerInfo {
		if this.Payers[info.UID()] == nil {
			this.Payers[info.UID()] = []*PrepayerInfo{}
		}
		this.Payers[info.UID()] = append(this.Payers[info.UID()], info)
	}
	return this
}

func (this *GasPrepayerLookup) SumPrepaidGas(addrSign string) ([]*PrepayerInfo, uint64) {
	successfulPayers := slice.TransformIf(this.Payers[addrSign], func(_ int, info *PrepayerInfo) (bool, *PrepayerInfo) {
		return info.Successful, info
	}) // Successful payers only.

	totalPrepaid := slice.Accumulate(successfulPayers, 0, func(i int, info *PrepayerInfo) uint64 { return info.PrepayedAmount })
	return successfulPayers, totalPrepaid
}

func (this *GasPrepayerLookup) Clear(addrSign string) bool {
	if this.Payers[addrSign] != nil {
		delete(this.Payers, addrSign)
		return true
	}
	return false
}
