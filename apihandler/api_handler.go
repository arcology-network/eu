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

package apihandler

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/mempool"
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	ethcommon "github.com/ethereum/go-ethereum/common"

	apicontainer "github.com/arcology-network/eu/apihandler/container"
	apicumulative "github.com/arcology-network/eu/apihandler/cumulative"
	apimultiprocess "github.com/arcology-network/eu/apihandler/multiprocess"
	apiruntime "github.com/arcology-network/eu/apihandler/runtime"
	intf "github.com/arcology-network/eu/interface"
)

type APIHandler struct {
	deployer   [20]byte // For transactions, the msg.sender, for sub-processes, the Multiprocessor's address
	logs       []intf.ILog
	depth      uint8
	serialNums [4]uint64 // sub-process/container/element/uuid generator,

	schedule interface{}
	eu       interface{}

	handlerDict map[[20]byte]intf.ApiCallHandler // APIs under the atomic namespace

	writeCachePool *mempool.Mempool[*tempcache.WriteCache]
	localCache     *tempcache.WriteCache // The private tempcache for the current APIHandler

	auxDict map[string]interface{} // Auxiliary data generated during the execution of the APIHandler

	execResult *eucommon.Result

	executionSubsidy uint64
}

func NewAPIHandler(writeCachePool *mempool.Mempool[*tempcache.WriteCache]) *APIHandler {
	api := &APIHandler{
		writeCachePool: writeCachePool,
		eu:             nil,
		localCache:     writeCachePool.New(),
		auxDict:        make(map[string]interface{}),
		handlerDict:    make(map[[20]byte]intf.ApiCallHandler),
		depth:          0,
		execResult:     &eucommon.Result{},
		serialNums:     [4]uint64{},
	}

	handlers := []intf.ApiCallHandler{
		apiruntime.NewIoHandlers(api),
		apimultiprocess.NewMultiprocessHandler(api),
		apicontainer.NewBaseHandlers(api),
		apicumulative.NewU256CumulativeHandler(api),
		// cumulativei256.NewInt256CumulativeHandlers(api),
		apiruntime.NewRuntimeHandlers(api),
	}

	for i, v := range handlers {
		if _, ok := api.handlerDict[(handlers)[i].Address()]; ok {
			panic("Error: Duplicate handler addresses found!! " + fmt.Sprint((handlers)[i].Address()))
		}
		api.handlerDict[(handlers)[i].Address()] = v
	}
	return api
}

// Initliaze a new APIHandler from an existing writeCache. This is different from the NewAPIHandler() function in that it does not create a new writeCache.
func (this *APIHandler) New(writeCachePool interface{}, localCache interface{}, deployer ethcommon.Address, schedule interface{}) intf.EthApiRouter {
	// localCache := writeCachePool.(*mempool.Mempool[*tempcache.WriteCache]).New()
	api := NewAPIHandler(this.writeCachePool)
	api.SetDeployer(deployer)
	// api.writeCachePool = writeCachePool.(*mempool.Mempool[*tempcache.WriteCache])
	api.writeCachePool = this.writeCachePool
	api.localCache = localCache.(*tempcache.WriteCache)
	api.depth = this.depth + 1
	api.deployer = deployer
	api.schedule = schedule
	api.auxDict = make(map[string]interface{})
	return api
}

// The Cascade() function creates a new APIHandler whose writeCache uses the parent APIHandler's writeCache as the
// read-only data store.  writecache -> parent's writecache -> backend datastore
func (this *APIHandler) Cascade() intf.EthApiRouter {
	api := NewAPIHandler(this.writeCachePool)
	api.SetDeployer(this.deployer)
	api.depth = this.depth + 1
	api.schedule = this.schedule
	api.auxDict = make(map[string]interface{})

	// writeCache := this.writeCachePool.New() // Get a new write tempcache from the shared write tempcache pool.
	writeCache := tempcache.NewWriteCache(this.localCache, 32, 1)

	// Use the current write tempcache as the read-only data store for the replicated APIHandler
	return api.SetWriteCache(writeCache.SetReadOnlyBackend(this.localCache))
}

func (this *APIHandler) AuxDict() map[string]interface{} { return this.auxDict }
func (this *APIHandler) WriteCachePool() interface{}     { return this.writeCachePool }

func (this *APIHandler) GetDeployer() ethcommon.Address         { return this.deployer }
func (this *APIHandler) SetDeployer(deployer ethcommon.Address) { this.deployer = deployer }

func (this *APIHandler) GetEU() interface{}   { return this.eu }
func (this *APIHandler) SetEU(eu interface{}) { this.eu = eu }

func (this *APIHandler) GetSchedule() interface{}         { return this.schedule }
func (this *APIHandler) SetSchedule(schedule interface{}) { this.schedule = schedule }

func (this *APIHandler) WriteCache() interface{} { return this.localCache }
func (this *APIHandler) SetWriteCache(writeCache interface{}) intf.EthApiRouter {
	this.localCache = writeCache.(*tempcache.WriteCache)
	return this
}

func (this *APIHandler) CheckRuntimeConstrains() bool { // Execeeds the max recursion depth or the max sub processes
	return this.Depth() < eucommon.MAX_RECURSIION_DEPTH &&
		atomic.AddUint64(&eucommon.TotalSubProcesses, 1) <= eucommon.MAX_TOTAL_VM_INSTANCES
}

func (this *APIHandler) DecrementDepth() uint8 {
	if this.depth > 0 {
		this.depth--
	}
	return this.depth
}

func (this *APIHandler) Depth() uint8 { return this.depth }

func (this *APIHandler) Coinbase() ethcommon.Address {
	return this.eu.(interface{ Coinbase() [20]byte }).Coinbase()
}

func (this *APIHandler) Origin() ethcommon.Address {
	if this.eu == nil {
		return [20]byte{}
	}
	return this.eu.(interface{ Origin() [20]byte }).Origin()
}

func (this *APIHandler) HandlerDict() map[[20]byte]intf.ApiCallHandler {
	return this.handlerDict
}

func (this *APIHandler) VM() interface{} {
	return common.IfThenDo1st(this.eu != nil, func() interface{} { return this.eu.(interface{ VM() interface{} }).VM() }, nil)
}

func (this *APIHandler) GetSerialNum(idx int) uint64 {
	v := this.serialNums[idx]
	this.serialNums[idx]++
	return v
}

func (this *APIHandler) Pid() [32]byte {
	return this.eu.(interface{ TxHash() [32]byte }).TxHash()
}

func (this *APIHandler) ElementUID() []byte {
	instanceID := this.Pid()
	serial := strconv.Itoa(int(this.GetSerialNum(eucommon.ELEMENT_ID)))
	return []byte(append(instanceID[:8], []byte(serial)...))
}

// Generate an UUID based on transaction hash and the counter
func (this *APIHandler) UUID() []byte {
	id := codec.Bytes32(this.Pid()).UUID(this.GetSerialNum(eucommon.UUID))
	return id[:8]
}

func (this *APIHandler) AddLog(key, value string) {
	this.logs = append(this.logs, &eucommon.ExecutionLog{
		Key:   key,
		Value: value,
	})
}

func (this *APIHandler) GetLogs() []intf.ILog {
	return this.logs
}

func (this *APIHandler) ClearLogs() {
	this.logs = this.logs[:0]
}

func (this *APIHandler) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, blockhash ethcommon.Hash, isReadOnly bool) (bool, []byte, bool, int64) {
	if handler, ok := this.handlerDict[callee]; ok {
		result, successful, fees := handler.Call(
			ethcommon.Address(codec.Bytes20(caller).Clone().(codec.Bytes20)),
			ethcommon.Address(codec.Bytes20(callee).Clone().(codec.Bytes20)),
			slice.Clone(input),
			origin,
			nonce,
			isReadOnly,
		)
		return true, result, successful, fees
	}
	return false, []byte{}, true, 0 // not an Arcology call, used 0 gas
}

// Exeuction subsidy is the amount of gas that the parallel TXs pay to offset the cost of execution of the deferred TX.
func (this *APIHandler) GetExecutionSubsidy() uint64        { return this.executionSubsidy }
func (this *APIHandler) SetExecutionSubsidy(subsidy uint64) { this.executionSubsidy = subsidy }
