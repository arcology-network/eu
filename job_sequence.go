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
	"github.com/arcology-network/vm-adaptor/eth"
	intf "github.com/arcology-network/vm-adaptor/interface"

	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/vm"
	evmparams "github.com/ethereum/go-ethereum/params"
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

// GetID returns the ID of the JobSequence.
func (this *JobSequence) GetID() uint32 { return this.ID }
func (this *JobSequence) AppendMsg(msg interface{}) {
	this.StdMsgs = append(this.StdMsgs, msg.(*eucommon.StandardMessage))
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

// Run executes the job sequence and returns the results.
func (this *JobSequence) Run(config *execution.Config, mainApi intf.EthApiRouter) ([]uint32, []*univalue.Univalue) {
	this.Results = make([]*execution.Result, len(this.StdMsgs))
	this.ApiRouter = mainApi.New(cache.NewWriteCache(mainApi.WriteCache().(*cache.WriteCache)), this.ApiRouter.Schedule()) // cascade the write caches

	for i, msg := range this.StdMsgs {
		pendingApi := this.ApiRouter.New((cache.NewWriteCache(this.ApiRouter.WriteCache().(*cache.WriteCache))), this.ApiRouter.Schedule())
		pendingApi.DecrementDepth()

		this.Results[i] = this.execute(msg, config, pendingApi)
		this.ApiRouter.WriteCache().(*cache.WriteCache).AddTransitions(this.Results[i].RawStateAccesses)
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
			(uint64(typed.Size())/32)*uint64(v.Writes())*evmparams.SstoreSetGas,
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
