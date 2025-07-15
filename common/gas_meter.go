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

package common

import (
	"math"
)

// GasMeter is a structure that holds the gas usage information for a transaction with respect to data read and written
// happening in the Arcology part of the execution.
type GasMeter struct {
	ReadDataSize  uint64
	WriteDataSize int64
	TotalGasUsed  int64
}

func NewGasMeter() *GasMeter {
	return &GasMeter{
		ReadDataSize:  0,
		WriteDataSize: 0,
		TotalGasUsed:  0,
	}
}

func (this *GasMeter) Use(readDataSize uint64, writeDataSize int64, gasUsed int64) *GasMeter {
	this.ReadDataSize += readDataSize
	this.WriteDataSize += writeDataSize
	this.TotalGasUsed += int64(math.Ceil(float64(readDataSize)/float64(DATA_UNIT_SIZE))+
		math.Ceil(float64(writeDataSize)/float64(DATA_UNIT_SIZE))) + gasUsed
	return this
}
