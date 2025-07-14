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
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"sync/atomic"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/mempool"
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"
	"github.com/arcology-network/eu/gas"
	"github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/holiman/uint256"

	stgcommon "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"

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

	schedule any
	eu       any

	handlerDict map[[20]byte]intf.ApiCallHandler // APIs under the atomic namespace

	writeCachePool *mempool.Mempool[*cache.WriteCache]
	localCache     *cache.WriteCache // The private cache for the current APIHandler

	auxDict map[string]interface{} // Auxiliary data generated during the execution of the APIHandler

	payer *gas.PrepayerInfo // Pay for the deferred execution gas.
}

func NewAPIHandler(writeCachePool *mempool.Mempool[*cache.WriteCache]) *APIHandler {
	api := &APIHandler{
		writeCachePool: writeCachePool,
		eu:             nil,
		localCache:     writeCachePool.New(),
		auxDict:        make(map[string]any),
		handlerDict:    make(map[[20]byte]intf.ApiCallHandler),
		depth:          0,
		serialNums:     [4]uint64{},

		payer: &gas.PrepayerInfo{}, // Initialize the gas prepayer lookup
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
func (this *APIHandler) New(writeCachePool interface{}, localCache interface{}, deployer ethcommon.Address, schedule any) intf.EthApiRouter {
	// localCache := writeCachePool.(*mempool.Mempool[*cache.WriteCache]).New()
	api := NewAPIHandler(this.writeCachePool)
	api.SetDeployer(deployer)
	// api.writeCachePool = writeCachePool.(*mempool.Mempool[*cache.WriteCache])
	api.writeCachePool = this.writeCachePool
	api.localCache = localCache.(*cache.WriteCache)
	api.depth = this.depth + 1
	api.deployer = deployer
	api.schedule = schedule
	api.auxDict = make(map[string]interface{})

	// api.gasPrepayer = gasPayer.(*gas.GasPrepayer) // Use the same gas prepayer as the parent APIHandler
	api.payer = &gas.PrepayerInfo{} // Initialize the gas prepayer lookup
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

	// writeCache := this.writeCachePool.New() // Get a new write cache from the shared write cache pool.
	writeCache := cache.NewWriteCache(this.localCache, 32, 1)

	// Use the current write cache as the read-only data store for the replicated APIHandler
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
	this.localCache = writeCache.(*cache.WriteCache)
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
	pid := this.Pid()
	instanceID := common.TrimTrail(pid[:8], 0) // Trim the trailing zeros to make it a valid UUID
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
		return true, result, successful, eucommon.GAS_CALL_API + fees
	}
	return false, []byte{}, true, 0 // not an Arcology call, used 0 gas
}

// For runtime caller to get the job information for the current call.
func (this *APIHandler) Job() any {
	return this.eu.(interface{ Job() *eucommon.Job }).Job()
}

// Either prepay the gas for the deferred execution of the job, or use the prepaid gas to pay for the deferred execution of the job.
func (this *APIHandler) PrepayGas(initGas *uint64, gasRemaining *uint64) (uint64, bool) {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil {
		return 0, true // Deployment / simple transfer TX, no need to prepay gas.
	}

	if ok, _ := this.IsDeferrable(); !ok {
		return 0, true // The job is deferred, no need to refund the prepaid gas.
	}

	// If the job is not deferred, we don't need to prepay the gas.
	txID := this.GetEU().(interface{ ID() uint64 }).ID()
	storage := this.WriteCache().(*cache.WriteCache)

	// Get the required prepaid gas amount.
	payerAmountPath := stgcommon.RequiredPrepaidGasAmountPath(*job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data))
	prepaidGasAmount, _, _ := storage.Read(txID, payerAmountPath, new(commutative.Int64))
	if prepaidGasAmount == nil || prepaidGasAmount.(int64) == 0 {
		return 0, true // No prepaid gas found info found, nothing to do.
	}

	// Check if the prepaid gas amount is enough to pay for the deferred execution of the job.
	PrepaidGas := uint64(prepaidGasAmount.(int64)) // Set
	if PrepaidGas > *gasRemaining {
		return 0, false // Not enough gas remaining to pay for the deferred execution.
	}

	// Log the gas info before prepaying the gas.
	this.payer.GasRemaining = *gasRemaining // Set the gas remaining for the job from the EVM
	this.payer.InitialGas = *initGas        // Set the initial gas for the job from the EVM
	this.payer.PrepayedAmount = PrepaidGas  // Set the gas amount that is already prepaid for the job.

	// Decrement the gas remaining and initial gas by the prepaid gas amount.
	*initGas -= PrepaidGas      // Subtract the prepaid gas from the initial gas.
	*gasRemaining -= PrepaidGas // Subtract the prepaid gas from the gas remaining.
	return 0, true
}

// Add the prepaid gas to top up job's remaining gas for the deferred execution.
func (this *APIHandler) UsePrepaidGas(gasRemaining *uint64) bool {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil || !job.StdMsg.IsDeferred {
		return false // Only available for deferred execution jobs.
	}

	info := (&gas.PrepayerInfo{}).FromJob(job) // Get the prepayer info from the job.
	lookup := this.PrepayerRegister()
	_, totalPrepaid := lookup.SumPrepaidGas(info.UID()) // Sum the prepaid gas for the job.
	*gasRemaining += totalPrepaid                       // Add the prepaid gas to the gas remaining for execution.
	return true
}

// This function refunds the prepaid gas when the prepaid gas is more than the gas used by the job.
func (this *APIHandler) RefundPrepaidGas(gasLeft *uint64) bool {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil || !job.StdMsg.IsDeferred { // Only available for deferred execution jobs.
		return false
	}

	// Not the same as IsDeferred.
	if ok, _ := this.IsDeferrable(); !ok {
		return false // The job is deferred, no need to refund the prepaid gas.
	}

	// This part is only executed at the end of the parallel TX.
	// Based on the job's execution result, we either refund the prepaid gas or write
	// the prepayer info to the prepayer register, so it can be used for the deferred execution.
	// We place the code here instead of the PrepayGas() function because by the time we reach the prepay gas
	// function, the job hasn't been executed yet, so we don't know if the job is successful or not.
	// Without the execution result, we cannot determine if we should refund the prepaid gas or not.
	if !job.StdMsg.IsDeferred {
		if job.Err != nil {
			*gasLeft += this.payer.PrepayedAmount // Parallel Execution failed, hasn't touched the deferred yet, refund the prepaid gas directly.
			return true
		}

		cache := this.WriteCache().(*cache.WriteCache)
		txID := this.GetEU().(interface{ ID() uint64 }).ID()
		path := stgcommon.PrepayersPath() + job.StdMsg.AddrAndSignature() + "/" + hex.EncodeToString(job.StdMsg.TxHash[:])
		this.payer.Successful = job.Successful()
		_, err := cache.Write(txID, path, noncommutative.NewBytes(this.payer.Encode())) // Write the prepayer info to the prepayer register.
		return err == nil
	}

	// In the deferred TX, refund the gas back to the prepayers based on the precentage of gas left, regardless of the job's success.
	lookup := this.PrepayerRegister()
	info := (&gas.PrepayerInfo{}).FromJob(job)
	successfulPayers, totalPrepaid := lookup.SumPrepaidGas(info.UID())

	// // Calculate the gas left after the job execution.
	originalGasRemaining := float64(job.GasRemaining) * float64(*gasLeft) / float64(job.GasRemaining+totalPrepaid)

	// Calculate the refund per payer based on the gas left and the number of payers.
	refundPerPayer := uint64(math.Round(float64(*gasLeft)-originalGasRemaining) / float64(len(successfulPayers)))

	// 	// Minus the prepaid portion from the gas left.
	(*gasLeft) -= (*gasLeft) - uint64(math.Round(originalGasRemaining))

	// 	// Refund the prepaid gas portion back to each prepayer.
	for _, payer := range successfulPayers {
		remaining := uint256.NewInt(refundPerPayer)
		remaining = remaining.Mul(remaining, uint256.MustFromBig(payer.GasPrice))

		// Credit the gas back to the payer's account.
		this.eu.(interface{ StateDB() vm.StateDB }).StateDB().AddBalance(payer.From, remaining)
	}
	return false
}

// This function sets the execution error for the current call 1instead of getting the error from the receipt.
func (this *APIHandler) SetExecutionErr(err error) {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil {
		return // Only available for deferred execution jobs.
	}
}

// If the the address and the function signature of the job is deferrable, then it returns true and the required prepaid gas amount.
func (this *APIHandler) IsDeferrable() (bool, uint64) {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	cache := this.WriteCache().(*cache.WriteCache)
	txID := this.GetEU().(interface{ ID() uint64 }).ID()

	funcSign := codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data[:])
	RequiredPrepaidGasAmountPath := stgcommon.RequiredPrepaidGasAmountPath(*job.StdMsg.Native.To, funcSign) // Generate the sub path for the prepaid gas amount.
	requiredAmount, _, _ := cache.Read(txID, RequiredPrepaidGasAmountPath, new(noncommutative.Int64))
	if requiredAmount == nil || requiredAmount.(int64) == 0 {
		return false, 0 // No prepaid gas found info found, nothing to do.
	}
	return true, uint64(requiredAmount.(int64))
}

func (this *APIHandler) PrepayerRegister() *gas.GasPrepayerLookup {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	info := (&gas.PrepayerInfo{}).FromJob(job) // Get the prepayer info from the job.
	txID := this.GetEU().(interface{ ID() uint64 }).ID()
	payerRegister := stgcommon.PrepayersPath() + info.UID() + "/" //+ hex.EncodeToString(job.StdMsg.TxHash[:])
	cache := this.WriteCache().(*cache.WriteCache)

	addrFuncPayers, _, _ := cache.Read(txID, payerRegister, new(commutative.Path))
	prepayers := make([]*gas.PrepayerInfo, addrFuncPayers.(*deltaset.DeltaSet[string]).Length())
	for i := 0; i < len(prepayers); i++ { // Iterate through the prepayer register to find the prepayer info.
		payerKey, _ := addrFuncPayers.(*commutative.Path).GetByIndex(uint64(i))
		buffer, _, _ := cache.Read(txID, payerRegister+payerKey, new(noncommutative.Bytes))
		prepayers[i] = new(gas.PrepayerInfo).Decode(buffer.(*noncommutative.Bytes).Value().([]byte)).(*gas.PrepayerInfo) // Decode the prepayer info.
	}
	return gas.NewGasPrepayerLookup(prepayers)
}
