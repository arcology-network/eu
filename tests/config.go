package tests

import (
	"errors"
	"math"
	"math/big"
	"path/filepath"

	commonlibcommon "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/mempool"
	commontypes "github.com/arcology-network/common-lib/types"
	"github.com/arcology-network/eu/cache"
	eucommon "github.com/arcology-network/eu/common"
	concurrenturl "github.com/arcology-network/storage-committer"
	"github.com/arcology-network/storage-committer/commutative"
	ccurlintf "github.com/arcology-network/storage-committer/interfaces"
	stgcommstorage "github.com/arcology-network/storage-committer/storage"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	evmcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	eu "github.com/arcology-network/eu"
	execution "github.com/arcology-network/eu/execution"
	apihandler "github.com/arcology-network/evm-adaptor/apihandler"
	"github.com/arcology-network/evm-adaptor/compiler"
	"github.com/arcology-network/evm-adaptor/eth"
	ccurlcommon "github.com/arcology-network/storage-committer/common"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	encoder = stgcommstorage.Rlp{}.Encode
	decoder = stgcommstorage.Rlp{}.Decode
)

func MainTestConfig() *execution.Config {
	vmConfig := vm.Config{}
	cfg := &execution.Config{
		ChainConfig: params.MainnetChainConfig,
		VMConfig:    &vmConfig,
		BlockNumber: big.NewInt(0),
		ParentHash:  evmcommon.Hash{},
		Time:        big.NewInt(0),
		Coinbase:    &Coinbase,
		GasLimit:    math.MaxUint64, // Should come from the message
		Difficulty:  big.NewInt(0),
	}
	cfg.Chain = new(execution.DummyChain)
	return cfg
}

// Choose which data source to use
func chooseDataStore() ccurlintf.Datastore {
	return stgcommstorage.NewParallelEthMemDataStore() // Eth trie datastore
	// return stgcommstorage.NewLevelDBDataStore("./leveldb") // Eth trie datastore
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), encoder, decoder)
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), cachedstorage.NewMemDB(), encoder, decoder)
}

// func NewTestEU(coinbase evmcommon.Address, genesisAccts ...evmcommon.Address) (*eu.EU, *execution.Config, ccurlintf.Datastore, *concurrenturl.StateCommitter, []*univalue.Univalue) {
func NewTestEU(coinbase evmcommon.Address, genesisAccts ...evmcommon.Address) *TestEu {
	datastore := chooseDataStore()
	datastore.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	api := apihandler.NewAPIHandler(mempool.NewMempool[*cache.WriteCache](16, 1, func() *cache.WriteCache {
		return cache.NewWriteCache(datastore, 32, 1)
	}, (&cache.WriteCache{}).Reset))

	statedb := eth.NewImplStateDB(api)
	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
	statedb.CreateAccount(Coinbase)

	for i := 0; i < len(genesisAccts); i++ {
		statedb.CreateAccount(genesisAccts[i])
		statedb.AddBalance(genesisAccts[i], new(big.Int).SetUint64(1e18))
	}
	_, transitions := cache.NewWriteCacheFilter(api.WriteCache()).ByType()

	// Deploy.
	committer := concurrenturl.NewStorageCommitter(datastore)
	committer.Import(transitions)
	committer.Sort()
	committer.Precommit([]uint32{0})
	committer.Commit()

	api = apihandler.NewAPIHandler(mempool.NewMempool[*cache.WriteCache](16, 1, func() *cache.WriteCache {
		return cache.NewWriteCache(datastore, 32, 1)
	}, (&cache.WriteCache{}).Reset))

	statedb = eth.NewImplStateDB(api)

	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	return &TestEu{
		eu:          eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api),
		config:      config,
		db:          datastore,
		committer:   committer,
		transitions: transitions,
	}
	// return eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api), config, datastore, committer, transitions
}

func DeployThenInvoke(targetPath, contractFile, version, contractName, funcName string, inputData []byte, checkNonce bool) (*evmcore.ExecutionResult, error, *eu.EU, *evmcoretypes.Receipt) {
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
	result, err := AliceCall(eu, *contractAddress, funcName, db)
	return result, err, eu, nil
}

func AliceDeploy(targetPath, contractFile, compilerVersion, contract string) (*eu.EU, *evmcommon.Address, ccurlintf.Datastore, []byte, error) {
	code, err := compiler.CompileContracts(targetPath, contractFile, compilerVersion, contract, true)
	if err != nil {
		return nil, nil, nil, []byte{}, err
	}

	if len(code) == 0 {
		return nil, nil, nil, []byte{}, errors.New("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	StdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	testEu := NewTestEU(Coinbase, Alice, Bob)

	receipt, execResult, err := testEu.eu.Run(StdMsg, execution.NewEVMBlockContext(testEu.config), execution.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()
	_, transitions := cache.NewWriteCacheFilter(testEu.eu.Api().WriteCache()).ByType()

	if receipt.Status != 1 || err != nil || execResult.Err != nil {
		return nil, nil, nil, []byte{}, errors.New("Error: Deployment failed!!!")
	}

	contractAddress := receipt.ContractAddress
	testEu.committer = concurrenturl.NewStorageCommitter(testEu.db)
	testEu.committer.Import(transitions)
	testEu.committer.Sort()
	testEu.committer.Precommit([]uint32{1})
	testEu.committer.Commit()

	return testEu.eu, &contractAddress, testEu.db, evmcommon.Hex2Bytes(code), nil
}

func AliceCall(executor *eu.EU, contractAddress evmcommon.Address, funcName string, datastore ccurlintf.Datastore) (*core.ExecutionResult, error) {
	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	// localCache := cache.NewWriteCache(datastore, 32, 1)
	api := apihandler.NewAPIHandler(mempool.NewMempool[*cache.WriteCache](16, 1, func() *cache.WriteCache {
		return cache.NewWriteCache(datastore, 32, 1)
	}, (&cache.WriteCache{}).Reset))

	statedb := eth.NewImplStateDB(api)
	eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api)

	data := crypto.Keccak256([]byte(funcName))[:4]
	msg := core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	StdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := executor.Run(StdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*StdMsg.Native)) // Execute it
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

// func AliceCall(eu *execution.EU, contractAddress evmcommon.Address, funcName string, committer *concurrenturl.StorageCommitter ) error {
// 	api := eu.NewAPIHandler(committer)
// 	eu.SetApi(api)

// 	data := crypto.Keccak256([]byte(funcName))[:4]
// 	msg := core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
// 	StdMsg := &execution.StandardMessage{
// 		ID:     1,
// 		TxHash: [32]byte{1, 1, 1},
// 		Native: &msg, // Build the message
// 		Source: commontypes.TX_SOURCE_LOCAL,
// 	}

// 	config := MainTestConfig()
// 	config.Coinbase = &Coinbase
// 	config.BlockNumber = new(big.Int).SetUint64(10000000)
// 	config.Time = new(big.Int).SetUint64(10000000)
// 	receipt, execResult, err := eu.Run(StdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*StdMsg.Native)) // Execute it
// 	// _, transitions : eu.Api().WriteCacheFilter().ByType()

// 	if err != nil {
// 		return (err)
// 	}

// 	if execResult != nil && execResult.Err != nil {
// 		return (execResult.Err)
// 	}

// 	if receipt.Status != 1 || err != nil {
// 		return errors.New("Error: Failed to call!!!")
// 	}
// 	return nil
// }

func DepolyContract(eu *eu.EU, committer *concurrenturl.StateCommitter, config *execution.Config, code string, funcName string, inputData []byte, nonce uint64, checkNonce bool) (error, *execution.Config, *eu.EU, *evmcoretypes.Receipt) {
	msg := core.NewMessage(Alice, nil, nonce, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	StdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, _, err := eu.Run(StdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*StdMsg.Native)) // Execute it

	if err != nil || receipt.Status != 1 {
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		return errors.New("Error: Deployment failed!!!" + errmsg), config, eu, nil
	}

	_, transitionsFiltered := cache.NewWriteCacheFilter(eu.Api().WriteCache()).ByType()
	// committer := eu.Api().Ccurl()
	committer.Import(transitionsFiltered)
	committer.Sort()
	committer.Precommit([]uint32{1})
	committer.Commit()
	return nil, config, eu, receipt
}

func CallContract(eu *eu.EU, contractAddress common.Address, inputData []byte, nonceIncrement uint64, checkNonce bool) (error, *eu.EU, *evmcore.ExecutionResult, *evmcoretypes.Receipt) {
	// data := crypto.Keccak256([]byte(funcName))[:4]
	// inputData = append(data, inputData...)

	msg := core.NewMessage(Alice, &contractAddress, 10+nonceIncrement, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), inputData, nil, false)
	StdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	var execResult *evmcore.ExecutionResult
	receipt, execResult, err := eu.Run(StdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*StdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()

	// msg = core.NewMessage(Alice, &contractAddress, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	// receipt, execResult, _ := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg))
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
