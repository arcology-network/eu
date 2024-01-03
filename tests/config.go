package tests

import (
	"errors"
	"math"
	"math/big"

	commontypes "github.com/arcology-network/common-lib/types"
	concurrenturl "github.com/arcology-network/concurrenturl"
	"github.com/arcology-network/concurrenturl/commutative"
	ccurlintf "github.com/arcology-network/concurrenturl/interfaces"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/concurrenturl/univalue"
	"github.com/arcology-network/eu"
	"github.com/arcology-network/eu/cache"
	eucommon "github.com/arcology-network/eu/common"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	evmcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/eu/execution"
	"github.com/arcology-network/vm-adaptor/compiler"
	"github.com/arcology-network/vm-adaptor/eth"
)

const (
	ROOT_PATH   = "./tmp/filedb/"
	BACKUP_PATH = "./tmp/filedb-back/"
)

var (
	encoder = ccurlstorage.Rlp{}.Encode
	decoder = ccurlstorage.Rlp{}.Decode
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
	return ccurlstorage.NewParallelEthMemDataStore() // Eth trie datastore
	// return ccurlstorage.NewLevelDBDataStore("./leveldb") // Eth trie datastore
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), encoder, decoder)
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), cachedstorage.NewMemDB(), encoder, decoder)
}

func NewTestEU() (*execution.EU, *execution.Config, ccurlintf.Datastore, *concurrenturl.StorageCommitter, []*univalue.Univalue) {
	datastore := chooseDataStore()
	datastore.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	localCache := cache.NewWriteCache(datastore)
	// if len(args) > 0 {
	// 	url = args[0].(*concurrenturl.StorageCommitter )
	// }
	api := eu.NewAPIRouter(localCache)

	statedb := eth.NewImplStateDB(api)
	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
	statedb.CreateAccount(Coinbase)

	statedb.CreateAccount(Alice)
	statedb.AddBalance(Alice, new(big.Int).SetUint64(1e18))

	statedb.CreateAccount(Bob)
	statedb.AddBalance(Bob, new(big.Int).SetUint64(1e18))

	// statedb.CreateAccount(eucommon.RUNTIME_HANDLER)
	// statedb.AddBalance(eucommon.RUNTIME_HANDLER, new(big.Int).SetUint64(1e18))

	// _, transitions := api.WriteCacheFilter().ByType()
	_, transitions := cache.NewWriteCacheFilter(api.WriteCache()).ByType()
	// indexer.Univalues(transitionsFiltered).Print()

	// fmt.Println("\n" + eucommon.FormatTransitions(transitions))

	// Deploy.
	url := concurrenturl.NewStorageCommitter(datastore)
	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{0})
	api = eu.NewAPIRouter(localCache)
	statedb = eth.NewImplStateDB(api)

	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	return execution.NewEU(config.ChainConfig, *config.VMConfig, statedb, api), config, datastore, url, transitions
}

func DeployThenInvoke(targetPath, contractFile, version, contractName, funcName string, inputData []byte, checkNonce bool) (error, *execution.EU, *evmcoretypes.Receipt) {
	eu, contractAddress, db, err := AliceDeploy(targetPath, contractFile, version, contractName)
	if err != nil {
		return err, nil, nil
	}

	if len(funcName) == 0 {
		return err, eu, nil
	}
	return AliceCall(eu, *contractAddress, funcName, db), eu, nil
}

func AliceDeploy(targetPath, contractFile, compilerVersion, contract string) (*execution.EU, *evmcommon.Address, ccurlintf.Datastore, error) {
	eu, config, db, url, _ := NewTestEU()

	code, err := compiler.CompileContracts(targetPath, contractFile, compilerVersion, contract, true)
	if err != nil || len(code) == 0 {
		return nil, nil, nil, errors.New("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	stdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
	// _, transitions := eu.Api().WriteCacheFilter().ByType()
	_, transitions := cache.NewWriteCacheFilter(eu.Api().WriteCache()).ByType()

	if receipt.Status != 1 || err != nil || execResult.Err != nil {
		return nil, nil, nil, errors.New("Error: Deployment failed!!!")
	}

	contractAddress := receipt.ContractAddress
	url = concurrenturl.NewStorageCommitter(db)
	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{1})
	return eu, &contractAddress, db, nil
}

func AliceCall(executor *execution.EU, contractAddress evmcommon.Address, funcName string, datastore ccurlintf.Datastore) error {
	config := MainTestConfig()
	config.Coinbase = &Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	localCache := cache.NewWriteCache(datastore)
	api := eu.NewAPIRouter(localCache)
	statedb := eth.NewImplStateDB(api)
	execution.NewEU(config.ChainConfig, *config.VMConfig, statedb, api)

	data := crypto.Keccak256([]byte(funcName))[:4]
	msg := core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	stdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := executor.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
	if err != nil {
		return (err)
	}

	if execResult != nil && execResult.Err != nil {
		return (execResult.Err)
	}

	if receipt.Status != 1 || err != nil {
		return errors.New("Error: Failed to call!!!")
	}
	return nil
}

// func AliceCall(eu *execution.EU, contractAddress evmcommon.Address, funcName string, ccurl *concurrenturl.StorageCommitter ) error {
// 	api := eu.NewAPIRouter(ccurl)
// 	eu.SetApi(api)

// 	data := crypto.Keccak256([]byte(funcName))[:4]
// 	msg := core.NewMessage(Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
// 	stdMsg := &execution.StandardMessage{
// 		ID:     1,
// 		TxHash: [32]byte{1, 1, 1},
// 		Native: &msg, // Build the message
// 		Source: commontypes.TX_SOURCE_LOCAL,
// 	}

// 	config := MainTestConfig()
// 	config.Coinbase = &Coinbase
// 	config.BlockNumber = new(big.Int).SetUint64(10000000)
// 	config.Time = new(big.Int).SetUint64(10000000)
// 	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
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

func DepolyContract(eu *execution.EU, ccurl *concurrenturl.StorageCommitter, config *execution.Config, code string, funcName string, inputData []byte, nonce uint64, checkNonce bool) (error, *execution.Config, *execution.EU, *evmcoretypes.Receipt) {
	msg := core.NewMessage(Alice, nil, nonce, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	stdMsg := &eucommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, _, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it

	if err != nil || receipt.Status != 1 {
		errmsg := ""
		if err != nil {
			errmsg = err.Error()
		}
		return errors.New("Error: Deployment failed!!!" + errmsg), config, eu, nil
	}

	_, transitionsFiltered := cache.NewWriteCacheFilter(eu.Api().WriteCache()).ByType()
	// ccurl := eu.Api().Ccurl()
	ccurl.Import(transitionsFiltered)
	ccurl.Sort()
	ccurl.Commit([]uint32{1})

	return nil, config, eu, receipt
}

func CallContract(eu *execution.EU, contractAddress common.Address, inputData []byte, nonceIncrement uint64, checkNonce bool) (error, *execution.EU, *evmcore.ExecutionResult, *evmcoretypes.Receipt) {
	// data := crypto.Keccak256([]byte(funcName))[:4]
	// inputData = append(data, inputData...)

	msg := core.NewMessage(Alice, &contractAddress, 10+nonceIncrement, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), inputData, nil, false)
	stdMsg := &eucommon.StandardMessage{
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
	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
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
