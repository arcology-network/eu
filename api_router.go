package eu

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl"
	eucommon "github.com/arcology-network/eu/common"
	ethcommon "github.com/ethereum/go-ethereum/common"

	// eucommon "github.com/arcology-network/eu/common"
	execution "github.com/arcology-network/eu/execution"
	apihandler "github.com/arcology-network/vm-adaptor/api"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
	intf "github.com/arcology-network/vm-adaptor/interface"
)

type API struct {
	logs  []intf.ILog
	depth uint8

	serialNums [4]uint64 // sub-process/container/element/uuid generator,

	schedule interface{}
	eu       intf.EU
	// reserved interface{}

	handlerDict map[[20]byte]intf.ApiCallHandler // APIs under the atomic namespace
	ccurl       *concurrenturl.ConcurrentUrl

	execResult *eucommon.Result

	filter intf.StateFilter
}

func NewAPI(ccurl *concurrenturl.ConcurrentUrl) *API {
	api := &API{
		eu:          nil,
		ccurl:       ccurl,
		handlerDict: make(map[[20]byte]intf.ApiCallHandler),
		depth:       0,
		execResult:  &eucommon.Result{},
		serialNums:  [4]uint64{},
	}
	api.filter = NewExportFilter(api)

	handlers := []intf.ApiCallHandler{
		apihandler.NewIoHandlers(api),
		apihandler.NewMultiprocessHandlers(
			api,
			common.To[*execution.JobSequence, intf.JobSequence]([]*execution.JobSequence{}),
			&execution.Generation{}),
		apihandler.NewBaseHandlers(api, nil),
		apihandler.NewU256CumulativeHandlers(api),
		// cumulativei256.NewInt256CumulativeHandlers(api),
		apihandler.NewRuntimeHandlers(api),
	}

	for i, v := range handlers {
		if _, ok := api.handlerDict[(handlers)[i].Address()]; ok {
			panic("Error: Duplicate handler addresses found!! " + fmt.Sprint((handlers)[i].Address()))
		}
		api.handlerDict[(handlers)[i].Address()] = v
	}

	// api.ccurl.NewAccount(
	// 	ccurlcommon.SYSTEM,
	// 	hex.EncodeToString(codec.Bytes20(runtime.NewHandler(api).Address()).Encode()),
	// )
	return api
}

func (this *API) New(ccurl *concurrenturl.ConcurrentUrl, schedule interface{}) intf.EthApiRouter {
	api := NewAPI(ccurl)
	api.depth = this.depth + 1
	return api
}

func (this *API) CheckRuntimeConstrains() bool { // Execeeds the max recursion depth or the max sub processes
	return this.Depth() < adaptorcommon.MAX_RECURSIION_DEPTH &&
		atomic.AddUint64(&adaptorcommon.TotalSubProcesses, 1) <= adaptorcommon.MAX_VM_INSTANCES
}

func (this *API) StateFilter() intf.StateFilter { return this.filter }

func (this *API) DecrementDepth() uint8 {
	if this.depth > 0 {
		this.depth--
	}
	return this.depth
}

func (this *API) Depth() uint8                { return this.depth }
func (this *API) Coinbase() ethcommon.Address { return this.eu.Coinbase() }
func (this *API) Origin() ethcommon.Address   { return this.eu.Origin() }

func (this *API) SetSchedule(schedule interface{}) { this.schedule = schedule }
func (this *API) Schedule() interface{}            { return this.schedule }

func (this *API) HandlerDict() map[[20]byte]intf.ApiCallHandler { return this.handlerDict }

func (this *API) VM() interface{} {
	return common.IfThenDo1st(this.eu != nil, func() interface{} { return this.eu.VM() }, nil)
}

func (this *API) GetEU() interface{}   { return this.eu }
func (this *API) SetEU(eu interface{}) { this.eu = eu.(intf.EU) }

func (this *API) Ccurl() *concurrenturl.ConcurrentUrl            { return this.ccurl }
func (this *API) SetCcurl(newCcurl *concurrenturl.ConcurrentUrl) { this.ccurl = newCcurl }

func (this *API) GetSerialNum(idx int) uint64 {
	v := this.serialNums[idx]
	this.serialNums[idx]++
	return v
}

func (this *API) Pid() [32]byte {
	return this.eu.TxHash()
}

func (this *API) ElementUID() []byte {
	instanceID := this.Pid()
	serial := strconv.Itoa(int(this.GetSerialNum(adaptorcommon.ELEMENT_ID)))
	return []byte(append(instanceID[:8], []byte(serial)...))
}

// Generate an UUID based on transaction hash and the counter
func (this *API) UUID() []byte {
	id := codec.Bytes32(this.Pid()).UUID(this.GetSerialNum(adaptorcommon.UUID))
	return id[:8]
}

func (this *API) AddLog(key, value string) {
	this.logs = append(this.logs, &eucommon.ExecutionLog{
		Key:   key,
		Value: value,
	})
}

func (this *API) GetLogs() []intf.ILog {
	return this.logs
}

func (this *API) ClearLogs() {
	this.logs = this.logs[:0]
}

func (this *API) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, blockhash ethcommon.Hash) (bool, []byte, bool, int64) {
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
