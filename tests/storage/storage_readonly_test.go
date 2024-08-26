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

package committertest

// import (
// 	"fmt"
// 	"net/http"
// 	"testing"
// 	"time"

// 	storage "github.com/arcology-network/common-lib/storage"
// 	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
// 	storage "github.com/arcology-network/storage-committer/storage/proxy"
// 	univalue "github.com/arcology-network/common-lib/types/storage/univalue"
// )

// func TestReadonlyStorageLocal(t *testing.T) {
// 	// Server end
// 	persistentDB := storage.NewMemoryDB()
// 	serverCachePolicy := storage.NewCachePolicy(1, 0.8)
// 	serverDataStore := storage.NewDataStore( serverCachePolicy, persistentDB, platform.Codec{}.Encode, platform.Codec{}.Decode)

// 	keys := []string{}
// 	values := []interface{}{}
// 	for i := 0; i < 8; i++ { // 8 in the server db
// 		keys = append(keys, fmt.Sprint(i))
// 		v := univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i)))
// 		values = append(values, v)

// 		persistentDB.Set(fmt.Sprint(i), platform.Codec{}.Encode(noncommutative.NewInt64(int64(i)))) // save to the DB directly
// 	}
// 	serverDataStore.Precommit(keys[:4], values[:4]) // 4 in the server side cache
// 	serverDataStore.Commit()

// 	// Simulated Client
// 	keys1 := []string{}
// 	values1 := []interface{}{}
// 	for i := 0; i < 8; i++ { // 8 in the server db
// 		keys1 = append(keys1, fmt.Sprint(i))
// 		values1 = append(values1, univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i))))
// 	}

// 	placeholderEncoder := func(v interface{}) []byte { return platform.Codec{}.Encode(v) }
// 	placeholderDecoder := func(bytes []byte) interface{} { return platform.Codec{}.Decode(bytes) }

// 	readonlyClientProxy := storage.NewReadonlyClient("", "", nil, serverDataStore)
// 	clientCachePolicy := storage.NewCachePolicy(1, 0.8)
// 	clientDataStore := storage.NewDataStore( clientCachePolicy, readonlyClientProxy, placeholderEncoder, placeholderDecoder)
// 	clientDataStore.Precommit(keys1[:2], values1[:2]) // 2 in the client side cache
// 	clientDataStore.Commit()

// 	// Retrive two entries from the client cache
// 	v0, err := clientDataStore.Retrive(keys1[0])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[0])
// 	}
// 	v1, err := clientDataStore.Retrive(keys1[1])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[1])
// 	}
// 	if v0 != values1[0] || v1 != values1[1] {
// 		t.Errorf("Error: Failed to retrive entries from client cache !")
// 	}
// 	v2, err := clientDataStore.Retrive(keys1[2])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[2])
// 	}
// 	v3, err := clientDataStore.Retrive(keys1[3])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[3])
// 	}
// 	if v2 == nil || v3 == nil {
// 		t.Error("Error: Failed to retrive entries from client cache !")
// 	}
// 	//readonlyClientProxy
// }

// func TestReadonlyStorageRemote(t *testing.T) {
// 	// Server end
// 	persistentDB := storage.NewMemoryDB()
// 	serverCachePolicy := storage.NewCachePolicy(1, 0.8)
// 	serverDataStore := storage.NewDataStore( serverCachePolicy, persistentDB, platform.Codec{}.Encode, platform.Codec{}.Decode)

// 	keys := []string{}
// 	values := []interface{}{}
// 	for i := 0; i < 8; i++ { // 8 in the server db
// 		keys = append(keys, fmt.Sprint(i))
// 		v := univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i)))
// 		values = append(values, v)
// 		persistentDB.Set(fmt.Sprint(i), platform.Codec{}.Encode(noncommutative.NewInt64(int64(i)))) // save to the DB directly
// 	}
// 	serverDataStore.Precommit(keys[:4], values[:4]) // 4 in the server side cache
// 	serverDataStore.Commit()

// 	server := storage.NewReadonlyServer("", platform.Codec{}.Encode, platform.Codec{}.Decode, serverDataStore)
// 	go func() {
// 		http.HandleFunc("/store", server.Receive)
// 		http.ListenAndServe(":8090", nil)
// 	}()
// 	time.Sleep(5 * time.Second)

// 	keys1 := []string{}
// 	values1 := []interface{}{}
// 	for i := 0; i < 8; i++ { // 8 in the server db
// 		keys1 = append(keys1, fmt.Sprint(i))
// 		values1 = append(values1, univalue.NewUnivalue(uint32(i), fmt.Sprint(i), 1, 1, 2, noncommutative.NewInt64(int64(i))))
// 	}

// 	proxyEncoder := func(v interface{}) []byte { return platform.Codec{}.Encode(v) }
// 	proxyDecoder := func(bytes []byte) interface{} { return platform.Codec{}.Decode(bytes) }

// 	readonlyClientProxy := storage.NewReadonlyClient("http://localhost:8090", "store", nil)
// 	clientCachePolicy := storage.NewCachePolicy(1, 0.8)
// 	clientDataStore := storage.NewDataStore( clientCachePolicy, readonlyClientProxy, proxyEncoder, proxyDecoder)
// 	clientDataStore.Precommit(keys1[:2], values1[:2]) // 2 in the client side cache
// 	clientDataStore.Commit()

// 	// Retrive two entries from the client cache
// 	v0, err := clientDataStore.Retrive(keys1[0])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[0])
// 	}
// 	v1, err := clientDataStore.Retrive(keys1[1])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[1])
// 	}
// 	if v0 != values1[0] || v1 != values1[1] {
// 		t.Error("Error: Failed to retrive entries from client cache !")
// 	}

// 	// Retrive two entries from the remove the SERVER CACHE
// 	v2, err := clientDataStore.Retrive(keys1[2])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[2])
// 	}
// 	v3, err := clientDataStore.Retrive(keys1[3])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[3])
// 	}
// 	if v2 == nil || v3 == nil {
// 		t.Error("Error: Failed to retrive entries from server cache !")
// 	}

// 	// Retrive two entries from the remove the SERVER STORAGE
// 	v4, err := clientDataStore.Retrive(keys1[4])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[4])
// 	}
// 	v5, err := clientDataStore.Retrive(keys1[5])
// 	if err != nil {
// 		t.Errorf("Retrive Error: %v path=%v !", err, keys1[5])
// 	}
// 	if v4 == nil || v5 == nil {
// 		t.Error("Error: Failed to retrive entries from server storage !")
// 	}
// 	//readonlyClientProxy
// }
