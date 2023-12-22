package tests

import (
	"errors"
	"math"
	"math/big"

	commontypes "github.com/arcology-network/common-lib/types"
	concurrenturl "github.com/arcology-network/concurrenturl"
	"github.com/arcology-network/concurrenturl/commutative"
	"github.com/arcology-network/concurrenturl/interfaces"
	ccurlstorage "github.com/arcology-network/concurrenturl/storage"
	"github.com/arcology-network/eu"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	evmcore "github.com/ethereum/go-ethereum/core"
	evmcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/params"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	execution "github.com/arcology-network/eu"
	eucommon "github.com/arcology-network/eu/common"
	ccapi "github.com/arcology-network/vm-adaptor/api"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
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

func MainTestConfig() *eucommon.Config {
	vmConfig := vm.Config{}
	cfg := &eucommon.Config{
		ChainConfig: params.MainnetChainConfig,
		VMConfig:    &vmConfig,
		BlockNumber: big.NewInt(0),
		ParentHash:  evmcommon.Hash{},
		Time:        big.NewInt(0),
		Coinbase:    &adaptorcommon.Coinbase,
		GasLimit:    math.MaxUint64, // Should come from the message
		Difficulty:  big.NewInt(0),
	}
	cfg.Chain = new(eucommon.DummyChain)
	return cfg
}

// Choose which data source to use
func chooseDataStore() interfaces.Datastore {
	return ccurlstorage.NewParallelEthMemDataStore() // Eth trie datastore
	// return ccurlstorage.NewLevelDBDataStore("./leveldb") // Eth trie datastore
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(0, 1), cachedstorage.NewMemDB(), encoder, decoder)
	// return cachedstorage.NewDataStore(nil, cachedstorage.NewCachePolicy(1000000, 1), cachedstorage.NewMemDB(), encoder, decoder)
}

func NewTestEU() (adaptorcommon.EUInterface, *eucommon.Config, interfaces.Datastore, *concurrenturl.ConcurrentUrl, []interfaces.Univalue) {
	datastore := chooseDataStore()
	datastore.Inject(ccurlcommon.ETH10_ACCOUNT_PREFIX, commutative.NewPath())

	url := concurrenturl.NewConcurrentUrl(datastore)
	// if len(args) > 0 {
	// 	url = args[0].(*concurrenturl.ConcurrentUrl)
	// }
	api := ccapi.NewAPI(url)

	statedb := eth.NewImplStateDB(api)
	statedb.PrepareFormer(evmcommon.Hash{}, evmcommon.Hash{}, 0)
	statedb.CreateAccount(adaptorcommon.Coinbase)

	statedb.CreateAccount(adaptorcommon.Alice)
	statedb.AddBalance(adaptorcommon.Alice, new(big.Int).SetUint64(1e18))

	statedb.CreateAccount(adaptorcommon.Bob)
	statedb.AddBalance(adaptorcommon.Bob, new(big.Int).SetUint64(1e18))

	// statedb.CreateAccount(adaptorcommon.RUNTIME_HANDLER)
	// statedb.AddBalance(adaptorcommon.RUNTIME_HANDLER, new(big.Int).SetUint64(1e18))

	_, transitions := api.StateFilter().ByType()
	// indexer.Univalues(transitionsFiltered).Print()

	// fmt.Println("\n" + adaptorcommon.FormatTransitions(transitions))

	// Deploy.
	url = concurrenturl.NewConcurrentUrl(datastore)
	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{0})
	api = ccapi.NewAPI(url)
	statedb = eth.NewImplStateDB(api)

	config := MainTestConfig()
	config.Coinbase = &adaptorcommon.Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	return eu.NewEU(config.ChainConfig, *config.VMConfig, statedb, api), config, datastore, url, transitions
}

func DeployThenInvoke(targetPath, contractFile, version, contractName, funcName string, inputData []byte, checkNonce bool) (error, *execution.EU, *evmcoretypes.Receipt) {
	eu, contractAddress, ccurl, err := AliceDeploy(targetPath, contractFile, version, contractName)
	if err != nil {
		return err, nil, nil
	}

	if len(funcName) == 0 {
		return err, eu, nil
	}
	return AliceCall(eu, *contractAddress, funcName, ccurl), eu, nil
}

func AliceDeploy(targetPath, contractFile, compilerVersion, contract string) (*execution.EU, *evmcommon.Address, *concurrenturl.ConcurrentUrl, error) {
	executor, config, db, url, _ := NewTestEU()

	code, err := compiler.CompileContracts(targetPath, contractFile, compilerVersion, contract, true)
	if err != nil || len(code) == 0 {
		return nil, nil, nil, errors.New("Error: Failed to generate the byte code")
	}

	// ================================== Deploy the contract ==================================
	msg := core.NewMessage(adaptorcommon.Alice, nil, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, true)
	stdMsg := &adaptorcommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := executor.(*eu.EU).Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
	_, transitions := executor.(*eu.EU).Api().StateFilter().ByType()

	if receipt.Status != 1 || err != nil || execResult.Err != nil {
		return nil, nil, nil, errors.New("Error: Deployment failed!!!")
	}

	contractAddress := receipt.ContractAddress
	url = concurrenturl.NewConcurrentUrl(db)
	url.Import(transitions)
	url.Sort()
	url.Commit([]uint32{1})
	return executor.(*eu.EU), &contractAddress, url, nil
}

func AliceCall(eu *execution.EU, contractAddress evmcommon.Address, funcName string, ccurl *concurrenturl.ConcurrentUrl) error {
	config := MainTestConfig()
	config.Coinbase = &adaptorcommon.Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	api := ccapi.NewAPI(ccurl)
	statedb := eth.NewImplStateDB(api)
	execution.NewEU(config.ChainConfig, *config.VMConfig, statedb, api)

	data := crypto.Keccak256([]byte(funcName))[:4]
	msg := core.NewMessage(adaptorcommon.Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	stdMsg := &adaptorcommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
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

// func AliceCall(eu *execution.EU, contractAddress evmcommon.Address, funcName string, ccurl *concurrenturl.ConcurrentUrl) error {
// 	api := ccapi.NewAPI(ccurl)
// 	eu.SetApi(api)

// 	data := crypto.Keccak256([]byte(funcName))[:4]
// 	msg := core.NewMessage(adaptorcommon.Alice, &contractAddress, 0, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
// 	stdMsg := &execution.StandardMessage{
// 		ID:     1,
// 		TxHash: [32]byte{1, 1, 1},
// 		Native: &msg, // Build the message
// 		Source: commontypes.TX_SOURCE_LOCAL,
// 	}

// 	config := MainTestConfig()
// 	config.Coinbase = &adaptorcommon.Coinbase
// 	config.BlockNumber = new(big.Int).SetUint64(10000000)
// 	config.Time = new(big.Int).SetUint64(10000000)
// 	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
// 	// _, transitions : eu.Api().StateFilter().ByType()

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

func DepolyContract(eu *execution.EU, config *eucommon.Config, code string, funcName string, inputData []byte, nonce uint64, checkNonce bool) (error, *eucommon.Config, *execution.EU, *evmcoretypes.Receipt) {
	msg := core.NewMessage(adaptorcommon.Alice, nil, nonce, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), evmcommon.Hex2Bytes(code), nil, false)
	stdMsg := &adaptorcommon.StandardMessage{
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

	_, transitionsFiltered := eu.Api().StateFilter().ByType()
	ccurl := eu.Api().Ccurl()
	ccurl.Import(transitionsFiltered)
	ccurl.Sort()
	ccurl.Commit([]uint32{1})

	return nil, config, eu, receipt
}

func CallContract(eu *execution.EU, contractAddress common.Address, inputData []byte, nonceIncrement uint64, checkNonce bool) (error, *execution.EU, *evmcore.ExecutionResult, *evmcoretypes.Receipt) {
	// data := crypto.Keccak256([]byte(funcName))[:4]
	// inputData = append(data, inputData...)

	msg := core.NewMessage(adaptorcommon.Alice, &contractAddress, 10+nonceIncrement, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), inputData, nil, false)
	stdMsg := &adaptorcommon.StandardMessage{
		ID:     1,
		TxHash: [32]byte{1, 1, 1},
		Native: &msg, // Build the message
		Source: commontypes.TX_SOURCE_LOCAL,
	}

	config := MainTestConfig()
	config.Coinbase = &adaptorcommon.Coinbase
	config.BlockNumber = new(big.Int).SetUint64(10000000)
	config.Time = new(big.Int).SetUint64(10000000)

	var execResult *evmcore.ExecutionResult
	receipt, execResult, err := eu.Run(stdMsg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(*stdMsg.Native)) // Execute it
	// _, transitions := eu.Api().StateFilter().ByType()

	// msg = core.NewMessage(adaptorcommon.Alice, &contractAddress, 1, new(big.Int).SetUint64(0), 1e15, new(big.Int).SetUint64(1), data, nil, false)
	// receipt, execResult, _ := eu.Run(evmcommon.BytesToHash([]byte{1, 1, 1}), 1, &msg, execution.NewEVMBlockContext(config), execution.NewEVMTxContext(msg))
	// _, transitions = eu.Api().StateFilter().ByType()

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
