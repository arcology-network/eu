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
	"math"
	"math/big"
	"strconv"
	"sync/atomic"

	"github.com/arcology-network/common-lib/codec"
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/mempool"
	"github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"

	softdeltaset "github.com/arcology-network/common-lib/exp/softdeltaset"
	"github.com/holiman/uint256"

	stgcommon "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/noncommutative"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
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

	auxDict map[string]any // Auxiliary data generated during the execution of the APIHandler

	// Temporarily holds gas info between the buyGas() and refundGas()
	// payer *gas.PrepayerInfo
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

		// payer: &gas.PrepayerInfo{}, // Initialize the gas prepayer lookup
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
func (this *APIHandler) New(writeCachePool any, localCache any, deployer ethcommon.Address, schedule any) intf.EthApiRouter {
	// localCache := writeCachePool.(*mempool.Mempool[*cache.WriteCache]).New()
	api := NewAPIHandler(this.writeCachePool)
	api.SetDeployer(deployer)
	// api.writeCachePool = writeCachePool.(*mempool.Mempool[*cache.WriteCache])
	api.writeCachePool = this.writeCachePool
	api.localCache = localCache.(*cache.WriteCache)
	api.depth = this.depth + 1
	api.deployer = deployer
	api.schedule = schedule
	api.auxDict = make(map[string]any)

	// api.gasPrepayer = gasPayer.(*gas.GasPrepayer) // Use the same gas prepayer as the parent APIHandler
	// api.payer = &gas.PrepayerInfo{} // Initialize the gas prepayer lookup
	return api
}

// The Cascade() function creates a new APIHandler whose writeCache uses the parent APIHandler's writeCache as the
// read-only data store.  writecache -> parent's writecache -> backend datastore
func (this *APIHandler) Cascade() intf.EthApiRouter {
	api := NewAPIHandler(this.writeCachePool)
	api.SetDeployer(this.deployer)
	api.depth = this.depth + 1
	api.schedule = this.schedule
	api.auxDict = make(map[string]any)

	// writeCache := this.writeCachePool.New() // Get a new write cache from the shared write cache pool.
	writeCache := cache.NewWriteCache(this.localCache, 32, 1)

	// Use the current write cache as the read-only data store for the replicated APIHandler
	return api.SetWriteCache(writeCache.SetReadOnlyBackend(this.localCache))
}

func (this *APIHandler) AuxDict() map[string]any { return this.auxDict }
func (this *APIHandler) WriteCachePool() any     { return this.writeCachePool }

func (this *APIHandler) GetDeployer() ethcommon.Address         { return this.deployer }
func (this *APIHandler) SetDeployer(deployer ethcommon.Address) { this.deployer = deployer }

func (this *APIHandler) GetEU() any   { return this.eu }
func (this *APIHandler) SetEU(eu any) { this.eu = eu }

func (this *APIHandler) GetSchedule() any         { return this.schedule }
func (this *APIHandler) SetSchedule(schedule any) { this.schedule = schedule }

func (this *APIHandler) WriteCache() any { return this.localCache }
func (this *APIHandler) SetWriteCache(writeCache any) intf.EthApiRouter {
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

func (this *APIHandler) VM() any {
	return common.IfThenDo1st(this.eu != nil, func() any { return this.eu.(interface{ VM() any }).VM() }, nil)
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

	// Get the required prepayment amount.
	callee, funcSign := *job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data)
	prepaidGasAmount := this.GetRequiredAmount(callee, funcSign)
	if prepaidGasAmount == nil {
		return 0, true // No prepaid gas found info found, nothing to do.
	}

	// Check if the prepaid gas amount is enough to pay for the deferred execution of the job.
	PrepaidGas := prepaidGasAmount.(uint64) // Set
	if PrepaidGas > *gasRemaining {
		return 0, false // Not enough gas remaining to pay for the deferred execution.
	}

	// Write the sender address, so it can be used later.
	txID, cache := this.GetTxContext()
	targetPath := stgcommon.PrepayersPath(callee, funcSign) + hexutil.Encode(job.StdMsg.Native.From[:]) + ":" + strconv.FormatUint(txID, 10)

	// Store the prepaid gas price.
	price := noncommutative.Bigint(*job.StdMsg.Native.GasPrice)
	if _, err := cache.Write(txID, targetPath, &price); err != nil {
		return 0, false
	}

	// Decrement the gas remaining and initial gas by the prepaid gas amount.
	*initGas -= PrepaidGas      // The initial gas after prepayment deduction.
	*gasRemaining -= PrepaidGas // Subtract the prepaid gas from the gas remaining.

	// For quick access later.
	job.GasRemaining = gasRemaining
	job.InitialGas = initGas
	job.PrepaidGas = PrepaidGas
	return 0, true
}

// Add the prepaid gas to top up job's remaining gas for the deferred execution.
func (this *APIHandler) UsePrepaidGas(gasRemaining *uint64) bool {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil || !this.HasDeferred() || !job.StdMsg.IsDeferred {
		return false // Only available for deferred execution jobs.
	}

	// info := (&gas.PrepayerInfo{}).FromJob(job) // Get the prepayer info from the job.
	// lookup := this.GetPrepayerLookup()
	// _, totalPrepaid := lookup.SumPrepaidGas(info.UID()) // Sum the prepaid gas for the job.
	txID, cache := this.GetTxContext()
	callee, funcSign := *job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data)
	payers, _, _ := cache.Read(txID, stgcommon.PrepayersPath(callee, funcSign), commutative.Path{})

	totalPayer := payers.(*softdeltaset.DeltaSet[string]).Length() // Total payers
	*gasRemaining += totalPayer * job.PrepaidGas                   //  All the jobs share the same PrepaidGas amount, so we can use it directly.
	return true
}

// This function refunds the prepaid gas when the prepaid gas is more than the gas used by the job.
func (this *APIHandler) RefundPrepaidGas(gasLeft *uint64) bool {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	if job.StdMsg.Native.To == nil || job.StdMsg.Native.Data == nil || !this.HasDeferred() { //
		return false
	}

	// A parallel TX and it is failed.
	if !job.StdMsg.IsDeferred && job.Err != nil {
		txID, cache := this.GetTxContext()
		callee, funcSign := *job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data)

		key := stgcommon.PrepayersPath(callee, funcSign) + hexutil.Encode(job.StdMsg.Native.From[:]) + ":" + strconv.FormatUint(txID, 10)
		if _, err := cache.Write(txID, key, nil); err != nil {
			return false
		}

		payer := hexutil.Encode(job.StdMsg.Native.From[:]) + ":" + strconv.FormatUint(txID, 10)
		payerPath, _ := common.GetParentPath(stgcommon.PrepayersPath(callee, funcSign) + hexutil.Encode(job.StdMsg.Native.From[:]))
		v, _, _ := cache.Read(txID, payerPath+payer, new(noncommutative.Bigint)) // Get the gas price
		gasPrice := v.(big.Int)

		refundPerPayer := job.PrepaidGas * uint64(gasPrice.Uint64()) // All the job shares the same PrepaidGas amount, so we can use it directly.
		remaining := uint256.NewInt(refundPerPayer)
		remaining = remaining.Mul(remaining, uint256.MustFromBig(&gasPrice))

		// // Credit the gas back to the payer's account.
		payerAddr := ethcommon.HexToAddress(payer) // Not need to remove the suffix. The function does it internally.
		this.eu.(interface{ Statedb() vm.StateDB }).Statedb().AddBalance(payerAddr, remaining)
	}

	// Not the same as IsDeferred.
	// if ok := this.HasDeferred(); !ok {
	// 	return false // The job is a deferred TX, no need to refund the prepaid gas.
	// }

	// This only gets executed in a parallel TX.
	// Based on the job's execution result, we either refund the prepaid gas or write
	// the prepayer info to the prepayer register, so it can be used for the deferred execution.
	// We place the code here instead of the PrepayGas() function because by the time we reach the prepay gas
	// function, the job hasn't been executed yet, so we don't know if the job is successful or not.
	// Without the execution result, we cannot determine if we should refund the prepaid gas or not.
	callee, funcSign := *job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data)
	if !job.StdMsg.IsDeferred {
		txID, cache := this.GetTxContext()
		if job.Err != nil { // Parallel Execution failed, refund the prepaid gas directly.
			*gasLeft += job.PrepaidGas

			// Remove the payer info because the job failed. No need to get the gas price from the storage anymore.
			targetPath := stgcommon.PrepayersPath(callee, funcSign) + hexutil.Encode(job.StdMsg.Native.From[:]) + ":" + strconv.FormatUint(txID, 10)
			if _, err := cache.Write(txID, targetPath, nil); err != nil {
				return false
			}
		}
		return true
	}

	// In the deferred TX, refund the gas back to the prepayers based on the precentage of gas left, regardless of the job's success.
	txID, cache := this.GetTxContext()

	payerPath, _ := common.GetParentPath(stgcommon.PrepayersPath(callee, funcSign) + hexutil.Encode(job.StdMsg.Native.From[:]))
	v, _, _ := cache.Read(txID, payerPath, new(commutative.Path)) // Just to ensure the path exists.
	payers := v.(*softdeltaset.DeltaSet[string]).Elements()

	// Calculate the gas left after the job execution.
	totalPrepaid := uint64(len(payers)) * job.PrepaidGas // All the job shared the same prepaid gas amount
	originalGasRemaining := float64(*job.GasRemaining) * float64(*gasLeft) / float64(*job.GasRemaining+totalPrepaid)

	// Calculate the refund per payer based on the gas left and the number of payers.
	refundPerPayer := uint64(math.Round(float64(*gasLeft)-originalGasRemaining) / float64(len(payers)))

	// Minus the prepaid portion from the gas left.
	(*gasLeft) -= (*gasLeft) - uint64(math.Round(originalGasRemaining))

	// 	// Refund the prepaid gas portion back to each prepayer.
	for _, payer := range payers {
		v, _, _ := cache.Read(txID, payerPath+payer, new(noncommutative.Bigint)) // Just to ensure the path exists.
		gasPrice := v.(big.Int)

		remaining := uint256.NewInt(refundPerPayer)
		remaining = remaining.Mul(remaining, uint256.MustFromBig(&gasPrice))

		// Credit the gas back to the payer's account.
		payerAddr := ethcommon.HexToAddress(payer) // Not need to remove the suffix. The function does it internally.
		this.eu.(interface{ Statedb() vm.StateDB }).Statedb().AddBalance(payerAddr, remaining)

		// Remove the payer info.
		if _, err := cache.Write(txID, stgcommon.PrepayersPath(callee, funcSign)+payer, nil); err != nil {
			return false
		}
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

// If the the address and the function signature of the job is deferrable, then it
// returns true and the required prepaid gas amount.
func (this *APIHandler) HasDeferred() bool {
	job := this.eu.(interface{ Job() *eucommon.Job }).Job()
	requiredAmount := this.GetRequiredAmount(*job.StdMsg.Native.To, codec.Bytes4{}.FromBytes(job.StdMsg.Native.Data[:]))
	return requiredAmount != nil
}

func (this *APIHandler) GetTxContext() (uint64, *cache.WriteCache) {
	return this.GetEU().(interface{ ID() uint64 }).ID(),
		this.WriteCache().(*cache.WriteCache)
}

// Get the required prepayment amount.
func (this *APIHandler) GetRequiredAmount(addr [20]byte, funcSign [4]byte) any {
	txID, cache := this.GetTxContext()
	payerAmountPath := stgcommon.RequiredPrepaymentPath(addr, funcSign)
	prepaidGasAmount, _, _ := cache.Read(txID, payerAmountPath, new(commutative.Uint64))
	return prepaidGasAmount
}
