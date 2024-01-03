package eu

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	ccurlintf "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/eu/cache"

	eucommon "github.com/arcology-network/eu/common"
	ethcommon "github.com/ethereum/go-ethereum/common"

	// eucommon "github.com/arcology-network/eu/common"
	execution "github.com/arcology-network/eu/execution"
	apicontainer "github.com/arcology-network/vm-adaptor/api/container"
	apicumulative "github.com/arcology-network/vm-adaptor/api/cumulative"
	apimultiprocess "github.com/arcology-network/vm-adaptor/api/multiprocess"
	apiio "github.com/arcology-network/vm-adaptor/api/runtime"
	adaptorintf "github.com/arcology-network/vm-adaptor/interface"
)

type APIRouter struct {
	logs       []adaptorintf.ILog
	depth      uint8
	serialNums [4]uint64 // sub-process/container/element/uuid generator,

	schedule interface{}
	eu       adaptorintf.EU

	handlerDict map[[20]byte]adaptorintf.ApiCallHandler // APIs under the atomic namespace

	localCache *cache.WriteCache
	dataReader ccurlintf.ReadOnlyDataStore

	execResult *eucommon.Result
}

func NewAPIRouter(cache *cache.WriteCache) *APIRouter {
	api := &APIRouter{
		eu:         nil,
		localCache: cache,
		// filter:      *cache.NewWriteCacheFilter(cache),
		handlerDict: make(map[[20]byte]adaptorintf.ApiCallHandler),
		depth:       0,
		execResult:  &eucommon.Result{},
		serialNums:  [4]uint64{},
	}

	handlers := []adaptorintf.ApiCallHandler{
		apiio.NewIoHandlers(api),
		apimultiprocess.NewMultiprocessHandler(
			api,
			common.To[*execution.JobSequence, adaptorintf.JobSequence]([]*execution.JobSequence{}),
			&execution.Generation{}),
		apicontainer.NewBaseHandlers(api, nil),
		apicumulative.NewU256CumulativeHandler(api),
		// cumulativei256.NewInt256CumulativeHandlers(api),
		apiio.NewRuntimeHandlers(api),
	}

	for i, v := range handlers {
		if _, ok := api.handlerDict[(handlers)[i].Address()]; ok {
			panic("Error: Duplicate handler addresses found!! " + fmt.Sprint((handlers)[i].Address()))
		}
		api.handlerDict[(handlers)[i].Address()] = v
	}
	return api
}

func (this *APIRouter) New(localCache interface{}, schedule interface{}) adaptorintf.EthApiRouter {
	api := NewAPIRouter(localCache.(*cache.WriteCache))
	api.depth = this.depth + 1
	return api
}

func (this *APIRouter) WriteCache() interface{} { return this.localCache }

// func (this *APIRouter) DataReader() interface{} { return this.dataReader }

func (this *APIRouter) CheckRuntimeConstrains() bool { // Execeeds the max recursion depth or the max sub processes
	return this.Depth() < eucommon.MAX_RECURSIION_DEPTH &&
		atomic.AddUint64(&eucommon.TotalSubProcesses, 1) <= eucommon.MAX_VM_INSTANCES
}

func (this *APIRouter) DecrementDepth() uint8 {
	if this.depth > 0 {
		this.depth--
	}
	return this.depth
}

func (this *APIRouter) Depth() uint8                { return this.depth }
func (this *APIRouter) Coinbase() ethcommon.Address { return this.eu.Coinbase() }
func (this *APIRouter) Origin() ethcommon.Address   { return this.eu.Origin() }

func (this *APIRouter) SetSchedule(schedule interface{}) { this.schedule = schedule }
func (this *APIRouter) Schedule() interface{}            { return this.schedule }

func (this *APIRouter) HandlerDict() map[[20]byte]adaptorintf.ApiCallHandler { return this.handlerDict }

func (this *APIRouter) VM() interface{} {
	return common.IfThenDo1st(this.eu != nil, func() interface{} { return this.eu.VM() }, nil)
}

func (this *APIRouter) GetEU() interface{}   { return this.eu }
func (this *APIRouter) SetEU(eu interface{}) { this.eu = eu.(adaptorintf.EU) }

func (this *APIRouter) SetReadOnlyDataSource(readOnlyDataSource interface{}) {
	this.dataReader = readOnlyDataSource.(ccurlintf.ReadOnlyDataStore)
}

func (this *APIRouter) GetSerialNum(idx int) uint64 {
	v := this.serialNums[idx]
	this.serialNums[idx]++
	return v
}

func (this *APIRouter) Pid() [32]byte {
	return this.eu.TxHash()
}

func (this *APIRouter) ElementUID() []byte {
	instanceID := this.Pid()
	serial := strconv.Itoa(int(this.GetSerialNum(eucommon.ELEMENT_ID)))
	return []byte(append(instanceID[:8], []byte(serial)...))
}

// Generate an UUID based on transaction hash and the counter
func (this *APIRouter) UUID() []byte {
	id := codec.Bytes32(this.Pid()).UUID(this.GetSerialNum(eucommon.UUID))
	return id[:8]
}

func (this *APIRouter) AddLog(key, value string) {
	this.logs = append(this.logs, &eucommon.ExecutionLog{
		Key:   key,
		Value: value,
	})
}

func (this *APIRouter) GetLogs() []adaptorintf.ILog {
	return this.logs
}

func (this *APIRouter) ClearLogs() {
	this.logs = this.logs[:0]
}

func (this *APIRouter) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, blockhash ethcommon.Hash) (bool, []byte, bool, int64) {
	if handler, ok := this.handlerDict[callee]; ok {
		result, successful, fees := handler.Call(
			ethcommon.Address(codec.Bytes20(caller).Clone().(codec.Bytes20)),
			ethcommon.Address(codec.Bytes20(callee).Clone().(codec.Bytes20)),
			common.Clone(input),
			origin,
			nonce,
		)
		return true, result, successful, fees
	}
	return false, []byte{}, true, 0 // not an Arcology call, used 0 gas
}
