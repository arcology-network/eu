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

package tests

// func TestGeneration(t *testing.T) {
// 	store := storage.NewHybirdStore()
// 	api := apihandler.NewAPIHandler(mempool.NewMempool[*cache.WriteCache](16, 1, func() *cache.WriteCache {
// 		return cache.NewWriteCache(store, 32, 1)
// 	}, func(cache *cache.WriteCache) { cache.Reset() }))

// 	currentPath, _ := os.Getwd()
// 	code, err := compiler.CompileContracts(path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/"), "examples/visit-counter/VisitCounter.sol", "0.8.19", "VisitCounter", true)
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	msg := CreateEthMsg(Alice, Bob, 0, 1000, 1e15, 1, evmcommon.Hex2Bytes(code), false, 0)
// 	generation := eu.NewGenerationFromMsgs(0, 12, []*evmcore.Message{&msg}, api)
// 	generation.Execute(api)
// }
