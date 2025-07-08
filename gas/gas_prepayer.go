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
GasPrepayer is a structure that holds a map of contract addresses to their prepaid gas messages.
It is used to manage the prepaid gas for contracts with deferred execution.
*/
type GasPrepayer struct {
	Payers    map[string][]*eucommon.Job
	PayersNew map[string][]*PrepayerInfo
}

func NewGasPrepayer() *GasPrepayer {
	return &GasPrepayer{
		Payers:    make(map[string][]*eucommon.Job),
		PayersNew: make(map[string][]*PrepayerInfo),
	}
}

func (this *GasPrepayer) Add(payerInfo []*PrepayerInfo) uint64 {
	for _, info := range payerInfo {
		addrSign := info.UID()
		gasAmount := info.PrepayedAmount
		if gasAmount == 0 {
			return 0
		}

		if _, exists := this.Payers[addrSign]; !exists {
			this.Payers[addrSign] = []*eucommon.Job{}
		}

		this.PayersNew[addrSign] = append(this.PayersNew[addrSign], info)
	}
	return 0
}

func (this *GasPrepayer) AddPrepayer(job *eucommon.Job) uint64 {
	addrSign := job.StdMsg.AddrAndSignature()
	gasAmount := job.StdMsg.PrepaidGas
	if len(addrSign) == 0 || gasAmount == 0 {
		return 0
	}

	if _, exists := this.Payers[addrSign]; !exists {
		this.Payers[addrSign] = []*eucommon.Job{}
	}

	this.Payers[addrSign] = append(this.Payers[addrSign], job)
	return gasAmount
}

func (this *GasPrepayer) SumPrepaidGas(addrSign string) (uint64, uint64) {
	totalPayers := uint64(0)
	totalPrepaid := slice.Accumulate(this.Payers[addrSign], 0, func(i int, job *eucommon.Job) uint64 {
		if job.Successful() {
			totalPayers++
			return job.StdMsg.PrepaidGas
		}
		return 0
	})
	return totalPayers, totalPrepaid
}
