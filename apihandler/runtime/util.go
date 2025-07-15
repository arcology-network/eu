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
package runtime

import (
	eucommon "github.com/arcology-network/eu/common"
	stgcommon "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	"github.com/arcology-network/storage-committer/type/commutative"
)

// Check if the property parent path exists, if not create it.
func (this *RuntimeHandlers) CreateFuncParentPath(caller [20]byte, txID uint64, cache *cache.WriteCache, gasMeter *eucommon.GasMeter) bool {
	propertyParent := stgcommon.PropertyPath(caller)
	if !cache.IfExists(propertyParent) {
		writeDataSize, err := cache.Write(txID, propertyParent, commutative.NewPath()) // Create the property path only when needed.
		gasMeter.Use(0, writeDataSize, 0)
		return err == nil
	}
	return true // If the property path write
}
