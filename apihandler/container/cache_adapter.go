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

package api

// // Get the number of elements in the container, EXCLUDING the nil elements.
// func (this *BaseHandlers) Length(path string) (uint64, bool, int64) {
// 	if len(path) == 0 {
// 		return 0, false, 0
// 	}

// 	if path, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
// 		return path.(*deltaset.DeltaSet[string]).NonNilCount(), true, 0
// 	}
// 	return 0, false, 0
// }

// // Get the number of elements in the container, INCLUDING the nil elements.
// func (this *BaseHandlers) FullLength(path string) (uint64, bool, int64) {
// 	if len(path) == 0 {
// 		return 0, false, 0
// 	}

// 	if path, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(commutative.Path)); path != nil {
// 		return path.(*deltaset.DeltaSet[string]).Length(), true, 0
// 	}
// 	return 0, false, 0
// }

// // Export all the elements in the container to a two-dimensional slice.
// // This function will read all the elements in the container.
// func (this *BaseHandlers) ReadAll(path string) ([][]byte, []bool, []int64) {
// 	length, _, _ := this.Length(path)
// 	entries := make([][]byte, length)
// 	flags := make([]bool, length)
// 	fees := make([]int64, length)

// 	slice.NewDo(int(length), func(i int) []byte {
// 		entries[i], flags[i], fees[i] = this.GetByIndex(path, uint64(i))
// 		return []byte{}
// 	})
// 	return entries, flags, fees
// }

// // Get the index of the element by its key
// func (this *BaseHandlers) GetByIndex(path string, idx uint64) ([]byte, bool, int64) {
// 	if value, _, err := this.api.WriteCache().(*tempcache.WriteCache).ReadAt(
// 		this.api.GetEU().(interface{ ID() uint64 }).ID(),
// 		path,
// 		idx,
// 		new(noncommutative.Bytes),
// 	); err == nil && value != nil {
// 		return value.([]byte), true, 0
// 	}
// 	return []byte{}, false, 0
// }

// // Set the element by its index
// func (this *BaseHandlers) SetByIndex(path string, idx uint64, bytes []byte) (bool, int64) {
// 	if len(path) == 0 {
// 		return false, 0
// 	}

// 	value := common.IfThen(bytes == nil, nil, noncommutative.NewBytes(bytes))
// 	if _, err := this.api.WriteCache().(*tempcache.WriteCache).WriteAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, idx, value); err == nil {
// 		return true, 0
// 	}
// 	return false, 0
// }

// // Get the element by its key
// func (this *BaseHandlers) GetByKey(path string) ([]byte, bool, int64) {
// 	if value, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, new(noncommutative.Bytes)); value != nil {
// 		return value.([]byte), true, 0
// 	}
// 	return []byte{}, false, 0
// }

// // Set the element by its key
// func (this *BaseHandlers) SetByKey(path string, bytes []byte) (bool, int64) {
// 	if len(path) > 0 {
// 		value := common.IfThen(bytes == nil, nil, noncommutative.NewBytes(bytes))
// 		if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, value); err == nil {
// 			return true, 0
// 		}
// 	}
// 	return false, 0
// }

// // Get the index of a key
// func (this *BaseHandlers) KeyAt(path string, index uint64) (string, int64) {
// 	if len(path) > 0 {
// 		key, _ := this.api.WriteCache().(*tempcache.WriteCache).KeyAt(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, index, new(noncommutative.Bytes))
// 		return key, 0
// 	}
// 	return "", 0
// }

// // Get the index of a key
// func (this *BaseHandlers) IndexOf(path string, key string) (uint64, int64) {
// 	if len(path) > 0 {
// 		index, _ := this.api.WriteCache().(*tempcache.WriteCache).IndexOf(this.api.GetEU().(interface{ ID() uint64 }).ID(), path, key, new(noncommutative.Bytes))
// 		return index, 0
// 	}
// 	return math.MaxUint64, 0
// }
