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
package scheduler

import (
	"testing"
	"time"
)

func TestCallee(t *testing.T) {
	numCalls := 100

	callees := make([]*Callee, numCalls)
	for i := 0; i < numCalls; i++ {
		callees[i] = &Callee{
			Index:          uint32(i),
			Address:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Signature:      [4]byte{1, 2, 3, 4},
			Indices:        []uint32{1, 2, 3, 4},
			SequentialOnly: false,
			Calls:          uint32(i),
		}
	}

	t0 := time.Now()
	for i := 0; i < numCalls; i++ {
		encoded, err := callees[i].Encode()
		if err != nil {
			t.Error(err)
		}
		callee2 := &Callee{}
		callee2.Decode(encoded)

		if callees[i].Index != callee2.Index {
			t.Error("Failed to encode/decode")
		}
	}
	t.Log("Time Spend to encode / Decode :", numCalls, time.Since(t0))
}

func BenchmarkTestCallee(t *testing.B) {
	numCalls := 1000000

	callees := make([]*Callee, numCalls)
	for i := 0; i < numCalls; i++ {
		callees[i] = &Callee{
			Index:          uint32(i),
			Address:        [8]byte{1, 2, 3, 4, 5, 6, 7, 8},
			Signature:      [4]byte{1, 2, 3, 4},
			Indices:        []uint32{1, 2, 3, 4},
			SequentialOnly: false,
			Calls:          uint32(i),
		}
	}

	t0 := time.Now()
	for i := 0; i < numCalls; i++ {
		encoded, err := callees[i].Encode()
		if err != nil {
			t.Error(err)
		}
		callee2 := &Callee{}
		callee2.Decode(encoded)
	}
	t.Log("Time Spend to encode / Decode :", numCalls, time.Since(t0))
}
