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

package multiprocessor

import (
	"errors"
	"math"
	"math/big"
	"sync/atomic"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/slice"
	univalue "github.com/arcology-network/storage-committer/type/univalue"

	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	evmcommon "github.com/ethereum/go-ethereum/common"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"

	"github.com/arcology-network/eu/abi"

	eu "github.com/arcology-network/eu"
	eucommon "github.com/arcology-network/eu/common"

	basecontainer "github.com/arcology-network/eu/apihandler/container"
	intf "github.com/arcology-network/eu/interface"
)

// APIs under the concurrency namespace
type MultiprocessHandler struct {
	*basecontainer.BaseHandlers
}

func NewMultiprocessHandler(ethApiRouter intf.EthApiRouter) *MultiprocessHandler {
	handler := &MultiprocessHandler{}
	handler.BaseHandlers = basecontainer.NewBaseHandlers(ethApiRouter, handler.Run, &eu.Generation{})
	return handler
}

func (this *MultiprocessHandler) Address() [20]byte { return eucommon.MULTIPROCESS_HANDLER }

func (this *MultiprocessHandler) Run(caller, callee [20]byte, input []byte, args ...interface{}) ([]byte, bool, int64) {
	if atomic.AddUint64(&eucommon.TotalSubProcesses, 1); !this.Api().CheckRuntimeConstrains() {
		return []byte{}, false, 0
	}

	input, err := abi.DecodeTo(input, 0, []byte{}, 2, math.MaxInt64)
	if err != nil {
		return []byte{}, false, 0
	}

	numThreads, err := abi.DecodeTo(input, 0, uint64(1), 1, 8)
	if err != nil {
		return []byte{}, false, 0
	}
	threads := common.Min(common.Max(uint8(numThreads), 1), math.MaxUint8) // [1, 255]

	path := this.Connector().Key(caller)
	length, successful, fee := this.Length(path)
	length = common.Min(eucommon.MAX_VM_INSTANCES, length)
	if !successful {
		return []byte{}, successful, fee
	}

	// Initialize a new generation
	fees := make([]int64, length)
	erros := make([]error, length)
	ethMsgs := make([]*evmcore.Message, length)

	slice.Foreach(ethMsgs, func(i int, _ **evmcore.Message) {
		funCall, successful, fee := this.GetByIndex(path, uint64(i)) // Get the function call data and the fee.
		fees[i] = fee

		if !successful { // Assign the fee to the fees array
			ethMsgs[i], erros[i] = nil, errors.New("Error: Failed to get the function call data")
		}
		// Convert the function call data to an ethereum message for execution.
		ethMsgs[i], erros[i] = this.WrapEthMsg(caller, funCall)
	})

	// Generate the configuration for the sub processes based on the current block context.
	subConfig := eucommon.NewConfigFromBlockContext(this.Api().GetEU().(interface{ VM() interface{} }).VM().(*vm.EVM).Context)
	newGen := eu.NewGenerationFromMsgs(0, threads, ethMsgs, this.Api())
	transitions := newGen.Execute(subConfig, this.Api()) // Run the job sequences in parallel.

	// Unify tx IDs
	mainTxID := uint64(this.Api().GetEU().(interface{ ID() uint64 }).ID())
	slice.Foreach(transitions, func(_ int, v **univalue.Univalue) { (*v).SetTx(mainTxID) })
	this.Api().WriteCache().(*tempcache.WriteCache).Insert(transitions) // Merge the write tempcache to the main tempcache

	// Prepare the return values to return to the caller.
	returnValues := make([][]byte, length)
	successes := make([]bool, length)
	totalSubGasUsed := uint64(0)
	for i, seq := range newGen.JobSeqs() {
		// only one job per sequence for multiprocessing
		successes[i] = seq.Results[0].Receipt.Status == 1 // Check if the transaction was successful
		returnValues[i] = seq.Results[0].EvmResult.Return()
		totalSubGasUsed += uint64(seq.Results[0].Receipt.GasUsed) // Get the gas used by the transaction

		// Append the sub logs to the main thread
		for _, log := range seq.Results[0].Receipt.Logs {
			this.Api().VM().(*vm.EVM).StateDB.AddLog(log)
		}
	}

	// Add the gas used by the sub processes to the main thread, the state is updated by transitions.
	// The receipt has to be processed separately.
	this.Api().VM().(*vm.EVM).ArcologyNetworkAPIs.CallContext.Contract.Gas += totalSubGasUsed
	// Sub processes may have been spawned during the execution, recheck it.
	if !this.Api().CheckRuntimeConstrains() {
		return []byte{}, false, fee
	}

	// Prepare the return values to return for the caller.
	encodedReturnedData, err := EncodeCallReturns(returnValues, successes)
	if err != nil {
		return []byte{}, false, slice.Sum[int64, int64](fees)
	}
	return encodedReturnedData, true, slice.Sum[int64, int64](fees)
}

// toEthMsgs converts the input byte slice into a list of ethereum messages.
func (this *MultiprocessHandler) WrapEthMsg(caller [20]byte, input []byte) (*evmcore.Message, error) {
	gasLimit, value, calleeAddr, funCall, err := abi.Parse4(input,
		uint64(0), 1, 32,
		uint256.NewInt(0), 1, 32,
		[20]byte{}, 1, 32,
		[]byte{}, 2, math.MaxInt64)

	if err != nil {
		return nil, err
	}

	transfer := value.ToBig()
	addr := evmcommon.Address(calleeAddr)
	msg := evmcore.NewMessage( // Build the message
		this.Api().Origin(), // Where the gas comes from, cannot use the caller here.
		&addr,
		0,
		transfer, // Amount to transfer
		gasLimit,
		this.Api().GetEU().(interface{ GasPrice() *big.Int }).GasPrice(), // gas price
		funCall,
		nil,
		false, // Don't checking nonce
	)
	return &msg, nil
}
