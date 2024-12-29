package exectest

import (
	"bytes"
	"testing"

	"github.com/arcology-network/common-lib/exp/mempool"
	eth "github.com/arcology-network/eu/eth"
	stgcomm "github.com/arcology-network/storage-committer/storage/committer"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	evmcommon "github.com/ethereum/go-ethereum/common"

	apihandler "github.com/arcology-network/eu/apihandler"
)

func TestStateDBV2GetNonexistBalance(t *testing.T) {
	// db := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), ccurlstorage.Rlp{}.Encode, ccurlstorage.Rlp{}.Decode)
	db := chooseDataStore()
	// db.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	// localCache := tempcache.NewWriteCache(datastore, 32, 1)
	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(db, 32, 1)
	}, func(tempcache *tempcache.WriteCache) { tempcache.Reset() }))

	account := evmcommon.BytesToAddress([]byte{201, 202, 203, 204, 205})
	ethStatedb := eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 1)
	ethStatedb.CreateAccount(account)
	_, transitions := api.WriteCache().(*tempcache.WriteCache).ExportAll()
	// fmt.Println("\n" + euCommon.FormatTransitions(transitions))
	committer := stgcomm.NewStateCommitter(db, nil)
	committer.Import(transitions)
	committer.Precommit([]uint64{1})
	committer.Commit(20)

	committer = stgcomm.NewStateCommitter(db, nil)
	ethStatedb = eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 2)
	balance := ethStatedb.GetBalance(account)
	if balance == nil || !balance.IsZero() {
		t.Fail()
	}
}

func TestStateDBV2GetNonexistCode(t *testing.T) {
	// db := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), ccurlstorage.Rlp{}.Encode, ccurlstorage.Rlp{}.Decode)
	db := chooseDataStore()
	// db.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	// localCache := tempcache.NewWriteCache(db, 32, 1)
	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(db, 32, 1)
	}, func(tempcache *tempcache.WriteCache) { tempcache.Reset() }))

	account := evmcommon.BytesToAddress([]byte{201, 202, 203, 204, 205}) // a random address, there should be no code.
	ethStatedb := eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 1)
	ethStatedb.CreateAccount(account)
	_, transitions := api.WriteCache().(*tempcache.WriteCache).ExportAll()
	// fmt.Println("\n" + euCommon.FormatTransitions(transitions))

	committer := stgcomm.NewStateCommitter(db, nil)
	committer.Import(transitions)
	committer.Precommit([]uint64{1})
	committer.Commit(20)

	committer = stgcomm.NewStateCommitter(db, nil)
	ethStatedb = eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 2)
	code := ethStatedb.GetCode(account)
	if len(code) != 0 {
		t.Error("The code length should be 0")
	}
}

func TestStateDBV2GetNonexistStorageState(t *testing.T) {
	// db := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), ccurlstorage.Rlp{}.Encode, ccurlstorage.Rlp{}.Decode)
	db := chooseDataStore()
	// meta := commutative.NewPath()
	// db.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, meta)

	// localCache := tempcache.NewWriteCache(db, 32, 1)
	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(db, 32, 1)
	}, func(tempcache *tempcache.WriteCache) { tempcache.Reset() }))

	account := evmcommon.BytesToAddress([]byte{201, 202, 203, 204, 205})
	ethStatedb := eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 1)
	ethStatedb.CreateAccount(account)
	_, transitions := api.WriteCache().(*tempcache.WriteCache).ExportAll()
	// fmt.Println("\n" + euCommon.FormatTransitions(transitions))
	committer := stgcomm.NewStateCommitter(db, nil)
	committer.Import(transitions)
	committer.Precommit([]uint64{1})
	committer.Commit(20)

	committer = stgcomm.NewStateCommitter(db, nil)
	ethStatedb = eth.NewImplStateDB(api)
	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 2)
	state := ethStatedb.GetState(account, evmcommon.Hash{})
	if !bytes.Equal(state.Bytes(), evmcommon.Hash{}.Bytes()) {
		t.Fail()
	}
}

// func TestEthStateDBInterfaces(t *testing.T) {
// 	// db := cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), ccurlstorage.Rlp{}.Encode, ccurlstorage.Rlp{}.Decode)
// 	db := chooseDataStore()
// 	meta := commutative.NewPath()
// 	db.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, meta)
// 	committer := stgcomm.NewStateCommitter(db)

// 	// localCache := tempcache.NewWriteCache(db, 32, 1)
// 	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
// 		return tempcache.NewWriteCache(db, 32, 1)
// 	}, func(tempcache *tempcache.WriteCache) { tempcache.Clear() }))

// 	account := evmcommon.BytesToAddress([]byte{201, 202, 203, 204, 205})
// 	ethStatedb := eth.NewImplStateDB(api)
// 	ethStatedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 1)
// 	ethStatedb.CreateAccount(account)
// 	_, transitions := api.WriteCache().(*tempcache.WriteCache).ExportAll()
// 	// fmt.Println("\n" + euCommon.FormatTransitions(transitions))
// 	committer.Import(transitions)
// 	committer.Precommit([]uint64{1})
// 	committer.Commit(20)

// 	committer = stgcomm.NewStateCommitter(db)
// 	ethStatedb = eth.NewImplStateDB(api)

// 	alice, bob := evmcommon.Address{}, evmcommon.Address{}
// 	slice.Fill(alice[:], 1)
// 	slice.Fill(bob[:], 2)

// 	ethStatedb.CreateAccount(alice)
// 	ethStatedb.CreateAccount(bob)

// 	ethStatedb.SetBalance(alice, big.NewInt(1111))
// 	ethStatedb.SetBalance(bob, big.NewInt(2222))

// 	if ethStatedb.GetBalance(alice).Cmp(big.NewInt(1111)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	if ethStatedb.GetBalance(bob).Cmp(big.NewInt(2222)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	ethStatedb.SubBalance(alice, big.NewInt(11))
// 	ethStatedb.SubBalance(bob, big.NewInt(22))

// 	if ethStatedb.GetBalance(alice).Cmp(big.NewInt(1100)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	if ethStatedb.GetBalance(bob).Cmp(big.NewInt(2200)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	if ethStatedb.PeekBalance(alice).Cmp(big.NewInt(1100)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	if ethStatedb.PeekBalance(bob).Cmp(big.NewInt(2200)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	ethStatedb.AddBalance(alice, big.NewInt(10))
// 	ethStatedb.AddBalance(bob, big.NewInt(11))

// 	if ethStatedb.GetBalance(alice).Cmp(big.NewInt(1110)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	if ethStatedb.GetBalance(bob).Cmp(big.NewInt(2211)) != 0 {
// 		t.Error("Wrong balance!")
// 	}

// 	ethStatedb.SetNonce(alice, uint64(11))
// 	ethStatedb.SetNonce(bob, uint64(22))

// 	if ethStatedb.GetNonce(alice) != uint64(1) {
// 		t.Error("Wrong Nonce!", ethStatedb.GetNonce(alice))
// 	}

// 	if ethStatedb.GetNonce(bob) != uint64(1) {
// 		t.Error("Wrong Nonce!", ethStatedb.GetNonce(bob))
// 	}

// 	ethStatedb.SetCode(alice, []byte{1, 2, 3, 4})
// 	ethStatedb.SetCode(bob, []byte{4, 5, 6, 7})

// 	if !bytes.Equal(ethStatedb.GetCode(alice), []byte{1, 2, 3, 4}) {
// 		t.Error("Wrong code!")
// 	}

// 	if !bytes.Equal(ethStatedb.GetCode(bob), []byte{4, 5, 6, 7}) {
// 		t.Error("Wrong code!")
// 	}

// 	// if _, err := committer.Write(1, "blcc://eth1.0/account/"+string(alice[:])+"/storage/container/ctrn-0/", noncommutative.NewString("path")); err != nil {
// 	// 	t.Error(err)
// 	// }

// 	localCache := api.WriteCache().(*tempcache.WriteCache)

// 	if _, err := localCache.Write(1, "blcc://eth1.0/account/"+string(alice[:])+"/storage/container/ctrn-0/elem-000", noncommutative.NewString("123")); err == nil {
// 		t.Error(err)
// 	}

// 	if _, err := localCache.Write(1, "blcc://eth1.0/account/"+string(alice[:])+"/storage/container/ctrn-0/elem-001", noncommutative.NewString("456")); err == nil {
// 		t.Error(err)
// 	}

// 	// Try to read an nonexistent entry from an nonexistent path, should fail !
// 	if value, _, _ := localCache.Read(1, "blcc://eth1.0/account/"+string(alice[:])+"/storage/container/ctrn-0/elem-000", nil); value != nil {
// 		t.Error("Error: Shouldn't be not found")
// 	}

// 	// try again
// 	if value, _, _ := localCache.Read(1, "blcc://eth1.0/account/"+string(alice[:])+"/storage/container/ctrn-0/elem-000", nil); value != nil {
// 		t.Error("Error: Shouldn't be not found")
// 	}
// }
