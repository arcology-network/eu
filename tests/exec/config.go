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

package exectest

import (
	"errors"
	"math"
	"math/big"
	"path/filepath"

	commonlibcommon "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/mempool"

	commontype "github.com/arcology-network/common-lib/types"
	statestore "github.com/arcology-network/storage-committer"
	stgcommon "github.com/arcology-network/storage-committer/common"
	stgcomm "github.com/arcology-network/storage-committer/storage/committer"
	ethstg "github.com/arcology-network/storage-committer/storage/ethstorage"
	"github.com/arcology-network/storage-committer/storage/proxy"
	storage "github.com/arcology-network/storage-committer/storage/proxy"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	evmcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/holiman/uint256"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	eu "github.com/arcology-network/eu"
	apihandler "github.com/arcology-network/eu/apihandler"
	adaptorcommon "github.com/arcology-network/eu/common"
	"github.com/arcology-network/eu/compiler"
	ethimpl "github.com/arcology-network/eu/eth"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	encoder = ethstg.Rlp{}.Encode
	decoder = ethstg.Rlp{}.Decode
)

func MainTestConfig() *adaptorcommon.Config {
	vmConfig := vm.Config{}
	cfg := &adaptorcommon.Config{
		ChainConfig: params.MainnetChainConfig,
		VMConfig:    &vmConfig,
		BlockNumber: big.NewInt(0),
		ParentHash:  evmcommon.Hash{},
		Time:        big.NewInt(0),
		Coinbase:    &Coinbase,
		GasLimit:    math.MaxUint64, // Should come from the message
		Difficulty:  big.NewInt(0),
	}
	cfg.Chain = new(adaptorcommon.DummyChain)
	return cfg
}

// Choose which data source to use
func chooseDataStore() stgcommon.ReadOnlyStore {
	// return ethstg.NewParallelEthMemDataStore() // Eth trie datastore
	return storage.NewMemDBStoreProxy() // Eth trie datastore
	// return ethstg.NewLevelDBDataStore("./leveldb") // Eth trie datastore
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), encoder, decoder)
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), cachedstorage.NewMemDB(), encoder, decoder)
}

func NewTestEU(coinbase evmcommon.Address, genesisAccts ...evmcommon.Address) *TestEu {
	datastore := chooseDataStore()
	// datastore.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	sstore := statestore.NewStateStore(datastore.(*proxy.StorageProxy))

	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(sstore.WriteCache, 32, 1) // Generation writecache
	}, func(tempcache *tempcache.WriteCache) { tempcache.Clear() }))

	statedb := ethimpl.NewImplStateDB(api)
	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
	statedb.CreateAccount(Coinbase)

	for i := 0; i < len(genesisAccts); i++ {
		statedb.CreateAccount(genesisAccts[i])
		v, _ := uint256.FromBig(new(big.Int).SetUint64(1e18))
		statedb.AddBalance(genesisAccts[i], v)
	}

	// Apply the transitions to the storage.
	_, transitions := tempcache.NewWriteCacheFilter(api.WriteCache()).ByType()

	store := statestore.NewStateStore(datastore.(*proxy.StorageProxy))
	committer := stgcomm.NewStateCommitter(datastore, store.GetWriters())
	committer.Import(transitions)
	committer.Precommit([]uint64{0})
	committer.Commit(20)

	// Init a new API
	api = apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(sstore, 32, 1)
	}, func(tempcache *tempcache.WriteCache) { tempcache.Clear() }))

	statedb = ethimpl.NewImplStateDB(api)

	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	return &TestEu{
		eu:          eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api),
		config:      config,
		store:       sstore,
		committer:   committer,
		transitions: transitions,
	}
}

// func CreateNewEu(oinbase evmcommon.Address, blockNum uint64) *eu.EU {
// 	config := MainTestConfig()
// 	config.Coinbase = &Coinbase
// 	config.BlockNumber = new(big.Int).SetUint64(blockNum)
// 	config.Time = new(big.Int).SetUint64(10000000)
// 	return eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api)
// }

func ConfigChain(coinbase evmcommon.Address, blockNum uint64) {
	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(blockNum)
	config.Time = new(big.Int).SetUint64(10000000)
}

func DeployThenInvoke(targetPath, contractFile, version, contractName, funcName string, inputData []byte, checkNonce bool, args ...uint64) (*evmcore.ExecutionResult, error, *eu.EU, *evmcoretypes.Receipt) {
	if !commonlibcommon.FileExists(filepath.Join(targetPath, contractFile)) {
		return nil, errors.New("Error: The contract is not found!!!"), nil, nil
	}

	eu, contractAddress, db, _, err := AliceDeploy(targetPath, contractFile, version, contractName)
	if err != nil {
		return nil, err, nil, nil
	}

	if len(funcName) == 0 {
		return nil, err, eu, nil
	}

	amount := uint64(0)
	if len(args) > 0 {
		amount = args[0]
	}
	result, err := AliceCall(eu, *contractAddress, funcName, db, amount)
	return result, err, eu, nil
}

func CreateEthMsg(from evmcommon.Address, to evmcommon.Address, nonce, value, gasLimit, gasPrice uint64, data []byte, checkNonce bool, tx uint64) core.Message {
	return core.NewMessage(
		Alice,
		&to,
		nonce,
		new(big.Int).SetUint64(value),
		gasLimit,
		new(big.Int).SetUint64(gasPrice),
		data,
		nil,
		true,
	)
}

func AliceDeploy(targetPath, contractFile, compilerVersion, contract string) (*eu.EU, *evmcommon.Address, stgcommon.ReadOnlyStore, []byte, error) {
	code, err := compiler.CompileContracts(targetPath, contractFile, compilerVersion, contract, true)
	if err != nil {
		return nil, nil, nil, []byte{}, err
	}

	if len(code) == 0 {
		return nil, nil, nil, []byte{}, errors.New("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	StdMsg := &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	testEu := NewTestEU(Coinbase, Alice, Bob)

	receipt, execResult, err := testEu.eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(testEu.config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	_, transitions := tempcache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

	// fmt.Print(v)
	if receipt.Status != 1 || err != nil || execResult.Err != nil {
		return nil, nil, nil, []byte{}, errors.New("Error: Deployment failed!!!")
	}

	statestore := testEu.store.(*statestore.StateStore)
	contractAddress := receipt.ContractAddress
	testEu.committer = stgcomm.NewStateCommitter(statestore.Store(), statestore.GetWriters())
	testEu.committer.Import(transitions)
	testEu.committer.Precommit([]uint64{1})
	testEu.committer.Commit(20)

	// testEu.eu.Api().WriteCache().(*tempcache.WriteCache).Clear()

	return testEu.eu, &contractAddress, testEu.store, evmcommon.Hex2Bytes(code), nil
}

func AliceCall(executor *eu.EU, contractAddress evmcommon.Address, funcName string, datastore stgcommon.ReadOnlyStore, amount uint64) (*core.ExecutionResult, error) {
	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	executor.Api().WriteCache().(*tempcache.WriteCache).Clear()

	// localCache := tempcache.NewWriteCache(datastore, 32, 1)
	api := apihandler.NewAPIHandler(mempool.NewMempool[*tempcache.WriteCache](16, 1, func() *tempcache.WriteCache {
		return tempcache.NewWriteCache(datastore, 32, 1)
	}, func(tempcache *tempcache.WriteCache) { tempcache.Clear() }))

	statedb := ethimpl.NewImplStateDB(api)
	eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api)

	data := crypto.Keccak256([]byte(funcName))[:4]
	msg := core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(amount), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	StdMsg := &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := executor.Run(StdMsg, adaptorcommon.NewEVMBlockContext(config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	if err != nil {
		return execResult, err
	}

	if execResult != nil && execResult.Err != nil {
		return execResult, (execResult.Err)
	}

	if receipt.Status != 1 || err != nil {
		return execResult, errors.New("Error: Failed to call!!!")
	}
	return execResult, nil
}

func DepolyContract(eu *eu.EU, committer *stgcomm.StateCommitter, config *adaptorcommon.Config, code string, funcName string, inputData []byte, nonce uint64, checkNonce bool) (error, *adaptorcommon.Config, *eu.EU, *evmcoretypes.Receipt) {
	msg := core.NewMessage(Alice, nil, nonce, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	StdMsg := &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	receipt, _, err := eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it

	if err != nil || receipt.Status != 1 {
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		return errors.New("Error: Deployment failed!!!" + errmsg), config, eu, nil
	}

	_, transitionsFiltered := tempcache.NewWriteCacheFilter(eu.Api().WriteCache()).ByType()
	// committer := eu.Api().Ccurl()
	committer.Import(transitionsFiltered)
	committer.Precommit([]uint64{1})
	committer.Commit(20)
	return nil, config, eu, receipt
}

func CallContract(eu *eu.EU, contractAddress common.Address, inputData []byte, nonceIncrement uint64, checkNonce bool) (error, *eu.EU, *evmcore.ExecutionResult, *evmcoretypes.Receipt) {
	// data := crypto.Keccak256([]byte(funcName))[:4]
	// inputData = append(data, inputData...)

	msg := core.NewMessage(Alice, &contractAddress, 10+nonceIncrement, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), inputData, nil, false)
	StdMsg := &commontype.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontype.TX_SOURCE_LOCAL,
	}

	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	var execResult *evmcore.ExecutionResult
	receipt, execResult, err := eu.Run(StdMsg, adaptorcommon.NewEVMBlockContext(config), adaptorcommon.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()

	// msg = core.NewMessage(Alice, &contractAddress, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	// receipt, execResult, _ := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, adaptorcommon.NewEVMBlockContext(config), adaptorcommon.NewEVMTxContext(msg))
	// _, transitions = eu.Api().WriteCacheFilter().ByType()

	if err != nil {
		return nil, nil, execResult, receipt
	}

	if receipt.Status != 1 {
		return execResult.Err, eu, execResult, receipt
	}

	if execResult != nil && execResult.Err != nil {
		return execResult.Err, eu, execResult, receipt
	}
	return nil, eu, execResult, receipt
}
