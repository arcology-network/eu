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
	"fmt"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/arcology-network/common-lib/exp/array"
	mapi "github.com/arcology-network/common-lib/exp/map"
	eucommon "github.com/arcology-network/eu/common"
	ethcommon "github.com/ethereum/go-ethereum/common"
	ethcore "github.com/ethereum/go-ethereum/core"
)

func TestSchedulerAdd(t *testing.T) {
	file := "../tmp/history"
	os.Remove(file)

	sch, err := NewScheduler(file)
	if err != nil {
		t.Error(err)
	}

	alice := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bob := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	carol := []byte("cccccccccccccccccccccccccccccccccccccccc")
	david := []byte("dddddddddddddddddddddddddddddddddddddddd")

	// eva := []byte("eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
	// frank := []byte("ffffffffffffffffffffffffffffffffffffffff")

	sch.Add([20]byte(alice), [4]byte{1, 1, 1, 1}, [20]byte(bob), [4]byte{2, 2, 2, 2})
	sch.Add([20]byte(carol), [4]byte{3, 3, 3, 3}, [20]byte(david), [4]byte{4, 4, 4, 4})

	sch.Add([20]byte(alice), [4]byte{1, 1, 1, 1}, [20]byte(bob), [4]byte{2, 2, 2, 2})
	sch.Add([20]byte(carol), [4]byte{3, 3, 3, 3}, [20]byte(david), [4]byte{4, 4, 4, 4})

	if len(sch.callees) != 4 {
		t.Error("Failed to add contracts")
	}
	SaveScheduler(sch, file)

	sch, err = LoadScheduler(file)
	if len(sch.callees) != 4 {
		t.Error("Failed to add contracts")
	}

	if sch.Add([20]byte(alice), [4]byte{1, 1, 1, 1}, [20]byte(bob), [4]byte{2, 2, 2, 2}) {
		t.Error("Should not exist")
	}

	if !sch.Add([20]byte(alice), [4]byte{1, 2, 1, 1}, [20]byte(bob), [4]byte{2, 2, 2, 2}) {
		t.Error("Failed to add contracts")
	}
	os.Remove(file)
}

func TestScheduler(t *testing.T) {
	scheduler, _ := NewScheduler("")

	alice := []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa")
	bob := []byte("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb")
	carol := []byte("cccccccccccccccccccccccccccccccccccccccc")
	david := []byte("dddddddddddddddddddddddddddddddddddddddd")

	scheduler.Add([20]byte(alice), [4]byte{1, 1, 1, 1}, [20]byte(bob), [4]byte{2, 2, 2, 2})
	scheduler.Add([20]byte(carol), [4]byte{3, 3, 3, 3}, [20]byte(david), [4]byte{4, 4, 4, 4})

	aaddr := ethcommon.BytesToAddress(alice)
	callAlice := &eucommon.StandardMessage{
		ID:     0,
		Native: &ethcore.Message{To: &aaddr, Data: []byte{1, 1, 1, 1}},
	}

	baddr := ethcommon.BytesToAddress(bob)
	callBob := &eucommon.StandardMessage{
		ID:     1,
		Native: &ethcore.Message{To: &baddr, Data: []byte{2, 2, 2, 2}},
	}

	caddr := ethcommon.BytesToAddress(carol)
	callCarol := &eucommon.StandardMessage{
		ID:     2,
		Native: &ethcore.Message{To: &caddr, Data: []byte{3, 3, 3, 3}},
	}

	daddr := ethcommon.BytesToAddress(david)
	callDavid := &eucommon.StandardMessage{
		ID:     3,
		Native: &ethcore.Message{To: &daddr, Data: []byte{4, 4, 4, 4}},
	}

	// Produce a schedule for the given transactions based on the conflicts information.
	sch := scheduler.New([]*eucommon.StandardMessage{callAlice, callBob, callCarol, callDavid})
	if len(sch.Sequentials) != 2 {
		t.Error("Failed to add contracts")
	}
}

func TestMapArrayComparison(t *testing.T) {
	arr := array.NewWith[int](1000000, func(_ int) int { return rand.Intn(1000000) })

	t0 := time.Now()
	m := mapi.FromArray(arr, func(int) bool { return true })
	v := false
	for i := 0; i < len(arr); i++ {
		v = m[i]
	}
	fmt.Println("Map", time.Since(t0), v)

	t0 = time.Now()
	sort.Ints(arr)
	for i := 0; i < len(arr); i++ {
		sort.SearchInts(arr, i)
	}
	fmt.Println("Binary", time.Since(t0), v)
}
