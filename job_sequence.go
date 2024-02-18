// Package execution provides functionality for executing job sequences.
package execution

import (
	"crypto/sha256"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	"github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/importer"
	"github.com/arcology-network/concurrenturl/univalue"
	cache "github.com/arcology-network/eu/cache"
	"github.com/arcology-network/eu/execution"

	ccurlintf "github.com/arcology-network/concurrenturl/interfaces"
	eucommon "github.com/arcology-network/eu/common"
	eth "github.com/arcology-network/vm-adaptor/eth"
	intf "github.com/arcology-network/vm-adaptor/interface"
	evmcommon "github.com/ethereum/go-ethereum/common"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	ethparams "github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// JobSequence represents a sequence of jobs to be executed.
type JobSequence struct {
	ID           uint32   // group id
	PreTxs       []uint32 ``
	StdMsgs      []*eucommon.StandardMessage
	Results      []*execution.Result
	ApiRouter    intf.EthApiRouter
	RecordBuffer []*univalue.Univalue
}

func (*JobSequence) T() *JobSequence { return &JobSequence{} }

// New creates a new JobSequence with the given ID and API router.
func (*JobSequence) New(id uint32, apiRouter intf.EthApiRouter) *JobSequence {
	return &JobSequence{
		ID:        id,
		ApiRouter: apiRouter,
	}
}

// NewFromCall creates a new JobSequence from the given call.
func (*JobSequence) NewFromCall(evmMsg *evmcore.Message, api intf.EthApiRouter) *JobSequence {
	newJobSeq := new(JobSequence).New(uint32(api.GetSerialNum(eucommon.SUB_PROCESS)), api)

	return newJobSeq.AppendMsg(&eucommon.StandardMessage{
		ID:     uint64(newJobSeq.GetID()),
		Native: evmMsg,
		TxHash: newJobSeq.DeriveNewHash(api.GetEU().(interface{ TxHash() [32]byte }).TxHash()),
	})
}

// GetID returns the ID of the JobSequence.
func (this *JobSequence) GetID() uint32 { return this.ID }
func (this *JobSequence) AppendMsg(msg interface{}) *JobSequence {
	this.StdMsgs = append(this.StdMsgs, msg.(*eucommon.StandardMessage))
	return this
}

// DeriveNewHash derives a new hash based on the original hash and the JobSequence ID.
func (this *JobSequence) DeriveNewHash(original [32]byte) [32]byte {
	return sha256.Sum256(array.Flatten([][]byte{
		codec.Bytes32(original).Encode(),
		codec.Uint32(this.ID).Encode(),
	}))
}

// Length returns the number of standard messages in the JobSequence.
func (this *JobSequence) Length() int { return len(this.StdMsgs) }

// SetNonceOffset sets the nonce offset for the given address. This is used to avoid conflicts when
// deploying contracts in different child threads. This happens when the child threads create by the multiprocessor
// are trying to deploy contracts at the same address.

//						 main thread (nonce = n)
//	     	 		          |
//						+---------+---------+
//						|                   |
//					child thread 0   child thread 1
//					(nonce = n)      (nonce = n)
//
// Both child threads are trying to deploy contracts starting with the nonce value n + 1. This will cause a conflict.
// So the solution is to give different nonce offsets to different child threads, so they can deploy their contracts at different addresses.
// This should only be used for transactions spawned by the multiprocessor. The external transactions should not use this.

// Run executes the job sequence and returns the results. nonceOffset is used to calculate the nonce of the transaction, in
// case there is a contract deployment in the sequence.
func (this *JobSequence) Run(config *execution.Config, mainApi intf.EthApiRouter, threadId uint64) ([]uint32, []*univalue.Univalue) {
	this.Results = make([]*execution.Result, len(this.StdMsgs))
	this.ApiRouter = mainApi.New(cache.NewWriteCache(mainApi.WriteCache().(*cache.WriteCache)), mainApi.GetDeployer(), this.ApiRouter.GetSchedule())

	for i, msg := range this.StdMsgs {
		// Create a new write cache for the message.
		pendingApi := this.ApiRouter.New((cache.NewWriteCache(this.ApiRouter.WriteCache().(*cache.WriteCache))), mainApi.GetDeployer(), this.ApiRouter.GetSchedule())

		// The api router always increments the depth, every time a new write cache is created from another one. But this isn't the case for
		// executing a sequence of messages. So we need to decrement it here.
		pendingApi.DecrementDepth()
		this.Results[i] = this.execute(msg, config, pendingApi)                                          // Execute the message and store the result.
		this.ApiRouter.WriteCache().(*cache.WriteCache).AddTransitions(this.Results[i].RawStateAccesses) // Merge the write cache of the pendingApi into the mainApi.
	}

	accessRecords := univalue.Univalues(this.ApiRouter.WriteCache().(*cache.WriteCache).Export()).To(indexer.IPAccess{})
	return array.Fill(make([]uint32, len(accessRecords)), this.ID), accessRecords
}

// GetClearedTransition returns the cleared transitions of the JobSequence.
func (this *JobSequence) GetClearedTransition() []*univalue.Univalue {
	if idx, _ := array.FindFirstIf(this.Results, func(v *execution.Result) bool { return v.Err != nil }); idx < 0 {
		return this.ApiRouter.WriteCache().(*cache.WriteCache).Export()
	}

	trans := array.Concate(this.Results,
		func(v *execution.Result) []*univalue.Univalue {
			return v.Transitions()
		},
	)
	return trans
}

// FlagConflict flags the JobSequence as conflicting.
func (this *JobSequence) FlagConflict(dict map[uint32]uint64, err error) {
	first, _ := array.FindFirstIf(this.Results, func(r *execution.Result) bool {
		_, ok := (dict)[r.TxIndex]
		return ok
	})

	for i := first; i < len(this.Results); i++ {
		this.Results[i].Err = err
	}
}

// execute executes a standard message and returns the result.
func (this *JobSequence) execute(StdMsg *eucommon.StandardMessage, config *execution.Config, api intf.EthApiRouter) *execution.Result {
	statedb := eth.NewImplStateDB(api)
	statedb.PrepareFormer(StdMsg.TxHash, [32]byte{}, uint32(StdMsg.ID))

	eu := NewEU(
		config.ChainConfig,
		vm.Config{},
		statedb,
		api,
	)

	receipt, evmResult, prechkErr :=
		eu.Run(
			StdMsg,
			execution.NewEVMBlockContext(config),
			execution.NewEVMTxContext(*StdMsg.Native),
		)

	return (&execution.Result{
		TxIndex:          uint32(StdMsg.ID),
		TxHash:           common.IfThenDo1st(receipt != nil, func() evmcommon.Hash { return receipt.TxHash }, evmcommon.Hash{}),
		RawStateAccesses: cache.NewWriteCacheFilter(api.WriteCache()).ToBuffer(),
		Err:              common.IfThenDo1st(prechkErr == nil, func() error { return evmResult.Err }, prechkErr),
		From:             StdMsg.Native.From,
		Coinbase:         *config.Coinbase,
		Receipt:          receipt,
		EvmResult:        evmResult,
		StdMsg:           StdMsg,
	}).Postprocess()
}

// CalcualteRefund calculates the refund amount for the JobSequence.
func (this *JobSequence) CalcualteRefund() uint64 {
	amount := uint64(0)
	for _, v := range this.ApiRouter.WriteCache().(*cache.WriteCache).Cache() {
		typed := v.Value().(ccurlintf.Type)
		amount += common.IfThen(
			!v.Preexist(),
			(uint64(typed.Size())/32)*uint64(v.Writes())*ethparams.SstoreSetGas,
			(uint64(typed.Size())/32)*uint64(v.Writes()),
		)
	}
	return amount
}

// RefundTo refunds the specified amount from the payer to the recipient.
func (this *JobSequence) RefundTo(payer, recipent *univalue.Univalue, amount uint64) (uint64, error) {
	credit := commutative.NewU256Delta(uint256.NewInt(amount), true).(*commutative.U256)
	if _, _, _, _, err := recipent.Value().(ccurlintf.Type).Set(credit, nil); err != nil {
		return 0, err
	}
	recipent.IncrementDeltaWrites(1)

	debit := commutative.NewU256Delta(uint256.NewInt(amount), false).(*commutative.U256)
	if _, _, _, _, err := payer.Value().(ccurlintf.Type).Set(debit, nil); err != nil {
		return 0, err
	}
	payer.IncrementDeltaWrites(1)

	return amount, nil
}
