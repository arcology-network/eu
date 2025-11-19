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

package stgtest

import (
	// "github.com/arcology-network/state-engine/interfaces"
	interfaces "github.com/arcology-network/state-engine/common"
	ethstg "github.com/arcology-network/state-engine/storage/ethstorage"
	stgproxy "github.com/arcology-network/state-engine/storage/proxy"
	// trie "github.com/ethereum/go-ethereum/trie"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	// encoder = stgtypecommonCodec{}.Encode
	// decoder = stgtypecommonCodec{}.Decode

	encoder = ethstg.Rlp{}.Encode
	decoder = ethstg.Rlp{}.Decode
)

func chooseDataStore() interfaces.ReadOnlyStore {
	// ethPair := associative.Pair[intf.Indexer[*statecell.StateCell], []intf.Datastore]{
	// 	First:  ethstg.NewIndexer(store),
	// 	Second: []intf.Datastore{store.(*stgproxy.StorageProxy).EthStore()},
	// }

	// ccPair := associative.Pair[intf.Indexer[*statecell.StateCell], []intf.Datastore]{
	// 	First:  ccstg.NewIndexer(store),
	// 	Second: []intf.Datastore{store.(*stgproxy.StorageProxy).CCStore()},
	// }

	// return storage.NewParallelEthMemDataStore() // Eth trie datastore
	// return storage.NewMemDBStoreProxy() // Eth trie datastore
	store := stgproxy.NewMemDBStoreProxy()
	// store.DisableCache()
	return store
	// return storage.NewLevelDBDataStore("/tmp")
	// return datastore.NewDataStore( datastore.NewCachePolicy(1000000, 1), memdb.NewMemoryDB(), encoder, decoder)
	// return storage.NewDataStore( storage.NewCachePolicy(0, 1), storage.NewMemoryDB(), encoder, decoder)
}
