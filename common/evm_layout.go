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

package common

// EvmWordSize represents the size of a word in the Ethereum Virtual Machine (EVM).
const (
	EvmWordSize = 32
)

// AlignToEvmForString aligns a string to the EVM word size by padding it with null bytes.
// It takes a string as input and returns a byte slice with the aligned string.
func AlignToEvmForString(str string) []byte {
	strLength := len(str)
	EvmWordBytes := strLength / EvmWordSize
	if strLength%EvmWordSize != 0 {
		EvmWordBytes = EvmWordBytes + 1
	}
	finalLengths := make([]byte, EvmWordBytes*EvmWordSize)
	for i := 0; i < len(finalLengths); i++ {
		if i < strLength {
			finalLengths[i] = str[i]
		} else {
			finalLengths[i] = byte(0)
		}
	}
	return finalLengths
}

// AlignToEvmForInt aligns an integer to the EVM word size by converting it to a byte slice.
// It takes an integer as input and returns a byte slice with the aligned integer.
func AlignToEvmForInt(length int) []byte {
	lens := []byte{}
	for {
		by := length % 256
		lens = append(lens, byte(by))
		length = length >> 8
		if length == 0 {
			break
		}
	}
	bysLength := len(lens)
	revertLengths := make([]byte, bysLength)
	idx := 0
	for i := bysLength - 1; i >= 0; i-- {
		revertLengths[idx] = lens[i]
		idx++
	}

	EvmWordBytes := bysLength / EvmWordSize
	if bysLength%EvmWordSize != 0 {
		EvmWordBytes = EvmWordBytes + 1
	}

	finalLengths := make([]byte, EvmWordBytes*EvmWordSize)
	idx = 0
	for ; idx < len(finalLengths)-bysLength; idx++ {
		finalLengths[idx] = byte(0)
	}
	for j := 0; j < bysLength; j++ {
		finalLengths[idx] = revertLengths[j]
		idx++
	}

	return finalLengths
}
