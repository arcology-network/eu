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
package apihandler

import (
	associative "github.com/arcology-network/common-lib/exp/associative"

	eucommon "github.com/arcology-network/eu/common"
)

/*
GasPrepayer is a structure that holds a map of contract addresses to their prepaid gas messages.
It is used to manage the prepaid gas for contracts with deferred execution.
*/
type GasPrepayer struct {
	payers map[string]associative.Pair[uint64, []*eucommon.Job]
}

func NewGasPrepayer() *GasPrepayer {
	return &GasPrepayer{
		payers: make(map[string]associative.Pair[uint64, []*eucommon.Job]),
	}
}

func (this *GasPrepayer) AddPrepayer(job *eucommon.Job) uint64 {
	if !job.Successful() {
		return 0
	}

	addrSign := job.StdMsg.AddrAndSignature()
	gasAmount := job.StdMsg.PrepaidGas
	if len(addrSign) == 0 || gasAmount == 0 {
		return 0
	}

	if _, exists := this.payers[addrSign]; !exists {
		this.payers[addrSign] = associative.Pair[uint64, []*eucommon.Job]{First: job.StdMsg.PrepaidGas, Second: []*eucommon.Job{job}}
	}

	this.payers[addrSign] = associative.Pair[uint64, []*eucommon.Job]{
		First:  this.payers[addrSign].First + job.StdMsg.PrepaidGas,
		Second: append(this.payers[addrSign].Second, job)}
	return gasAmount
}

// func (this *GasPrepayer) RemovePrepayer(job *eucommon.Job) bool {
// 	addrSign := job.StdMsg.AddrAndSignature()
// 	gasAmount := job.StdMsg.PrepaidGas
// 	if len(addrSign) == 0 || gasAmount == 0 {
// 		return false
// 	}

// 	_, exists := this.payers[addrSign]
// 	if exists {
// 		delete(this.payers, addrSign)
// 	}
// 	return exists
// }

func (this *GasPrepayer) GetPrepaiedGas(addrSign string) uint64 { return this.payers[addrSign].First }
