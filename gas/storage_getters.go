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

// import (
// 	eucommon "github.com/arcology-network/eu/common"
// 	cache "github.com/arcology-network/storage-committer/storage/cache"
// 	"github.com/arcology-network/storage-committer/type/noncommutative"
// )

// func SetRequiredPrepayment(caller [20]byte, funSign [4]byte, cache *cache.WriteCache, eu any, gasMeter *eucommon.GasMeter) bool {
// 	txID := eu.(interface{ ID() uint64 }).ID()
// 	RequiredPrepaymentPath := stgcommon.RequiredPrepaymentPath(caller, funSign) // Generate the sub path for the prepaid gas amount.
// 	writeDataSize, err := cache.Write(txID, RequiredPrepaymentPath, noncommutative.NewInt64(int64(prepaidGas.(uint64))))
// 	gasMeter.Use(0, writeDataSize, 0)
// 	return err != nil
// }

// func GetRuntimeInfo() {
// 	return this.api.GetEU().(interface{ ID() uint64 }).ID(),
// 		this.api.WriteCache().(*cache.WriteCache)
// }
