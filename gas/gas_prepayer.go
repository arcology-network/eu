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
	eucommon "github.com/arcology-network/eu/common"
)

/*
GasPrepayerLookup is a structure that holds a map of contract addresses to their prepaid gas messages.
It is used to manage the prepaid gas for contracts with deferred execution.
*/
type GasPrepayerLookup struct {
	// PayersNew map[string][]*eucommon.Job
	PayersNew map[string][]*PrepayerInfo
}

func NewGasPrepayer() *GasPrepayerLookup {
	return &GasPrepayerLookup{
		// PayersNew:    make(map[string][]*eucommon.Job),
		PayersNew: make(map[string][]*PrepayerInfo),
	}
}

// func (this *GasPrepayerLookup) Add(payerInfo []*PrepayerInfo) uint64 {
// 	for _, info := range payerInfo {
// 		addrSign := info.UID()
// 		gasAmount := info.PrepayedAmount
// 		if gasAmount == 0 {
// 			return 0
// 		}

// 		if _, exists := this.PayersNew[addrSign]; !exists {
// 			this.PayersNew[addrSign] = []*eucommon.Job{}
// 		}

// 		this.PayersNew[addrSign] = append(this.PayersNew[addrSign], info)
// 	}
// 	return 0
// }

func (this *GasPrepayerLookup) AddPrepayer(job *eucommon.Job) uint64 {
	addrSign := job.StdMsg.AddrAndSignature()
	gasAmount := job.StdMsg.PrepaidGas
	if len(addrSign) == 0 || gasAmount == 0 {
		return 0
	}

	if _, exists := this.PayersNew[addrSign]; !exists {
		this.PayersNew[addrSign] = []*PrepayerInfo{}
	}
	this.PayersNew[addrSign] = append(this.PayersNew[addrSign], (&PrepayerInfo{}).FromJob(job))
	return gasAmount
}

func (this *GasPrepayerLookup) SumPrepaidGas(addrSign string) (uint64, uint64) {
	totalPayers := uint64(0)
	totalPrepaid := slice.Accumulate(this.PayersNew[addrSign], 0, func(i int, info *PrepayerInfo) uint64 {
		if info.Successful {
			totalPayers++
			return info.PrepayedAmount
		}
		return 0
	})
	return totalPayers, totalPrepaid
}
