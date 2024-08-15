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

package arbitrator

import (
	"testing"

	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
)

func TestArbiCreateTwoAccountsNoConflict(t *testing.T) {
	t.Run("one entry", func(t *testing.T) { // Reads only, should be no conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 1, 0, 0, noncommutative.NewBytes([]byte{1, 2}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 0 {
			t.Error("Error: There should be NO conflict")
		}
	})
	t.Run("read & read write", func(t *testing.T) { // Reads only, should be no conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 1, 0, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 1, 0, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 0 {
			t.Error("Error: There should be NO conflict")
		}
	})

	t.Run("delta write & delta write", func(t *testing.T) { // Delta writes only, should be no conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 0, 0, 1, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 0, 0, 2, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 0 {
			t.Error("Error: There should be NO conflict")
		}
	})

	t.Run("write & write", func(t *testing.T) { // Read delta write, should be 1 conflict
		// Write only, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 0, 2, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 0, 2, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("Read & Delta-write", func(t *testing.T) { // Read delta write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 0, 2, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 0, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		sorted := univalue.Univalues([]*univalue.Univalue{_0, _1}).Sort([]uint32{0, 1})

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, sorted)

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("Read & Delta-write", func(t *testing.T) { // Read delta write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 0, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 0, 2, noncommutative.NewBytes([]byte{2, 3}), nil)

		sorted := univalue.Univalues([]*univalue.Univalue{_0, _1}).Sort([]uint32{0, 1})

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, sorted)

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("Read & Delta-write", func(t *testing.T) { // Read delta write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 0, 0, 2, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 0, 2, noncommutative.NewBytes([]byte{2, 3}), nil)

		sorted := univalue.Univalues([]*univalue.Univalue{_0, _1}).Sort([]uint32{0, 1})

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, sorted)

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("read-Write & read-write write", func(t *testing.T) { // Read delta write, should be 1 conflict
		// Read write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("read-Write & read-write write", func(t *testing.T) { // Read delta write, should be 1 conflict
		// Read write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 0, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("read-Write & read-write write", func(t *testing.T) { // Read delta write, should be 1 conflict
		// Read write, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 0, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("read-Write & read-write write", func(t *testing.T) { // write vs delta wirte, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 2, 1, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})

	t.Run("read-Write & read-write write", func(t *testing.T) { // write vs delta wirte, should be 1 conflict
		_0 := univalue.NewUnivalue(0, "blcc://eth1.0/account/0x0000000", 2, 2, 0, noncommutative.NewBytes([]byte{1, 2}), nil)
		_1 := univalue.NewUnivalue(1, "blcc://eth1.0/account/0x0000000", 2, 2, 1, noncommutative.NewBytes([]byte{2, 3}), nil)

		arib := new(Arbitrator)
		ids := arib.Detect([]uint32{0, 1}, []*univalue.Univalue{_0, _1})

		conflictdict, _, _ := Conflicts(ids).ToDict()
		if len(conflictdict) != 1 {
			t.Error("Error: There should be ONE conflict")
		}
	})
}
