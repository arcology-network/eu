package committertest

// "github.com/arcology-network/concurrent-evm/core/rawdb"

// func TestLevelResetRead(t *testing.T) {
// 	dir := "./mydb"
// 	testkey := []byte("abc-123")
// 	testval := []byte("ok")

// 	leveldb, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "temp", false)
// 	if err != nil {
// 		fmt.Printf("err1:%v\n", err)
// 		return
// 	}
// 	leveldb.Put(testkey, testval)

// 	leveldb.Close()

// 	leveldb1, err := rawdb.NewLevelDBDatabase(dir, 256, 16, "temp", false)
// 	if err != nil {
// 		fmt.Printf("err2:%v\n", err)
// 		return
// 	}

// 	val, err := leveldb1.Get(testkey)

// 	if err != nil || len(val) != 2 {
// 		fmt.Printf("err3:%v\n", err)
// 		return
// 	}

// 	if !bytes.Equal(val, testval) {
// 		fmt.Printf("retrive val is:%v\n", string(val))
// 		return
// 	}

// }
// func TestStateStoreResetRead(t *testing.T) {
// 	dbpath := "./myStore"
// 	sstore := statestore.NewStateStore(stgproxy.NewLevelDBStoreProxy(dbpath))
// 	alice := AliceAccount()
// 	// sstore:= statestore.NewStateStore(store)

// 	if _, err := adaptorcommon.CreateNewAccount(stgcomm.SYSTEM, alice, sstore.WriteCache); err != nil { // NewAccount account structure {
// 		t.Error(err)
// 	}
// 	acctTrans := univalue.Univalues(slice.Clone(sstore.Export(univalue.Sorter))).To(univalue.IPTransition{})

// 	committer := stgcommitter.NewStateCommitter(sstore.ReadOnlyStore(), sstore.GetWriters())
// 	committer.Import(acctTrans)
// 	committer.Precommit([]uint32{stgcomm.SYSTEM})
// 	committer.Commit(stgcomm.SYSTEM)

// 	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), noncommutative.NewBytes([]byte{1, 2, 3})); err != nil {
// 		t.Error(err)
// 	}
// 	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), noncommutative.NewBytes([]byte{2, 2, 3})); err != nil {
// 		t.Error(err)
// 	}
// 	if _, err := sstore.Write(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), noncommutative.NewBytes([]byte{199, 45, 67})); err != nil {
// 		t.Error(err)
// 	}
// 	acctTrans = univalue.Univalues(slice.Clone(sstore.Export(univalue.Sorter))).To(univalue.IPTransition{})

// 	// committer.Import(acctTrans)
// 	committer = stgcommitter.NewStateCommitter(sstore.ReadOnlyStore(), sstore.GetWriters())
// 	committer.Import(acctTrans)
// 	committer.Precommit([]uint32{1})
// 	committer.Commit(2)

// 	// sstore1 := statestore.NewStateStore(stgproxy.NewLevelDBStoreProxy(dbpath))

// 	// outV, _, _ := sstore1.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
// 	// if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
// 	// 	t.Error("Error: The path should exist", outV)
// 	// }

// 	// outV, _, _ = sstore1.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(1), new(noncommutative.Bytes))
// 	// if outV == nil || !bytes.Equal(outV.([]byte), []byte{2, 2, 3}) {
// 	// 	t.Error("Error: The path should exist", outV)
// 	// }

// 	// outV, _, _ = sstore1.Read(1, "blcc://eth1.0/account/"+alice+"/storage/container/"+RandomKey(0), new(noncommutative.Bytes))
// 	// if outV == nil || !bytes.Equal(outV.([]byte), []byte{199, 45, 67}) {
// 	// 	t.Error("Error: The path should exist", outV)
// 	// }

// }

// func TestStateStoreResetRead1(t *testing.T) {
// 	dbpath := "./myStore"
// 	sstore := statestore.NewStateStore(stgproxy.NewLevelDBStoreProxy(dbpath))
// 	alice := AliceAccount()

// 	outV, _, _ := sstore.Read(1, "blcc://eth1.0/account/"+alice+"/storage/native/"+RandomKey(0), new(noncommutative.Bytes))
// 	if outV == nil || !bytes.Equal(outV.([]byte), []byte{1, 2, 3}) {
// 		t.Error("Error: The path should exist", outV)
// 	}
// }
