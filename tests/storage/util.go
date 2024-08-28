/*
 *   Copyright (c) 2023 Arcology Network

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

// Generate a random account, testing only
package stgtest

import (
	"fmt"
	"log"
	"math/rand"
	"testing"
	"time"

	slice "github.com/arcology-network/common-lib/exp/slice"
	eth "github.com/arcology-network/eu/eth"
	statestore "github.com/arcology-network/storage-committer"
	interfaces "github.com/arcology-network/storage-committer/common"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	opadapter "github.com/arcology-network/storage-committer/op"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	"github.com/arcology-network/storage-committer/storage/proxy"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	"github.com/arcology-network/storage-committer/type/univalue"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	rlp "github.com/ethereum/go-ethereum/rlp"
	"golang.org/x/crypto/sha3"
	"golang.org/x/exp/constraints"
)

func RandomAccount() string {
	var letters = []byte("abcdef0123456789")
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	addr := hexutil.Encode(b)
	return addr
}

func AliceAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 10)
	return hexutil.Encode(b)
}

func BobAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 11)
	return hexutil.Encode(b)
}

func CarolAccount() string {
	b := make([]byte, 20)
	slice.Fill(b, 12)
	return hexutil.Encode(b)
}

func RandomAccounts(n int) []string {
	accounts := make([]string, n)
	for i := range accounts {
		b := sha3.Sum256([]byte(fmt.Sprintf("%v", rand.Intn(1000000))))
		accounts[i] = hexutil.Encode(b[:20])
	}
	return accounts
}

func rlpEncoder(args ...interface{}) []byte {
	encoded, err := rlp.EncodeToBytes(args)
	if err != nil {
		log.Fatal("Error encoding data:", err)
	}
	return encoded
}

func RandomKey[T constraints.Integer](seed T) string {
	buf := sha3.Sum256([]byte(fmt.Sprint(seed)))
	return hexutil.Encode(buf[:20])
}

func RandomKeys[T constraints.Integer](s0, s1 T) []string {
	keys := make([]string, s1-s0)
	for i := range keys {
		keys[i] = RandomKey(s0 + T(i))
	}
	return keys
}

// Initiate the input new accounts in the tempcache
func NewAcountsInCache(writeCache *tempcache.WriteCache, accounts ...string) {
	// sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	// writeCache := sstore.WriteCache
	for i := range accounts {
		if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, accounts[i], writeCache); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}
	}
}

func NewWriteCacheWithAcounts(store interfaces.ReadOnlyStore, accounts ...string) *tempcache.WriteCache {
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	for i := range accounts {
		if _, err := eth.CreateNewAccount(stgcommcommon.SYSTEM, accounts[i], writeCache); err != nil { // NewAccount account structure {
			fmt.Println(err)
		}
	}
	return writeCache
}

func verifierEthMerkle(roothash [32]byte, acct string, key string, store interfaces.ReadOnlyStore, t *testing.T) {
	// roothash := store.(*stgproxy.StorageProxy).EthStore().Root()                               // Get the proof provider by a root hash.
	ethdb := store.(*stgproxy.StorageProxy).EthStore().EthDB()                       // Get the proof provider by a root hash.
	provider, err := ethstg.NewMerkleProofCache(2, ethdb).GetProofProvider(roothash) // Initiate the proof tempcache, maximum 2 blocks
	if err != nil {
		t.Fatal(err)
	}

	// Verify Bob's stgproxy for a big int value.
	bobAddr := ethcommon.BytesToAddress([]byte(hexutil.MustDecode(acct)))
	accountResult, err := provider.GetProof(bobAddr, []string{key})
	if err := accountResult.Validate(provider.Root()); err != nil {
		t.Error(err)
	}

	opProof := opadapter.Convertible(*accountResult).New() // Convert Bob's proof to OP format and verify.
	if err := opProof.Verify(provider.Root()); err != nil {
		t.Error(err)
	}
}

// It's mainly used for TESTING purpose.
func FlushToStore(sstore *statestore.StateStore) interfaces.ReadOnlyStore {
	acctTrans := univalue.Univalues(slice.Clone(sstore.Export(univalue.Sorter))).To(univalue.IPTransition{})
	txs := slice.Transform(acctTrans, func(_ int, v *univalue.Univalue) uint32 {
		return v.GetTx()
	})

	committer := stgcommitter.NewStateCommitter(sstore, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit(txs) // Write all the transitions to the store
	committer.Commit(10)
	sstore.Clear()
	return sstore
}

// It's mainly used for TESTING purpose.
func FlushGeneration(sstore *statestore.StateStore) []uint32 {
	acctTrans := univalue.Univalues(slice.Clone(sstore.Export(univalue.Sorter))).To(univalue.IPTransition{})
	txs := slice.Transform(acctTrans, func(_ int, v *univalue.Univalue) uint32 {
		return v.GetTx()
	})

	committer := stgcommitter.NewStateCommitter(sstore, sstore.GetWriters())
	committer.Import(acctTrans)
	committer.Precommit(txs) // Write all the transitions to the store
	return txs
}
