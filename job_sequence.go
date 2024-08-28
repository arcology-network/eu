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

// Package execution provides functionality for executing job sequences.
package execution

import (
	"crypto/sha256"

	"github.com/arcology-network/common-lib/codec"
	"github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	slice "github.com/arcology-network/common-lib/exp/slice"
	commontype "github.com/arcology-network/common-lib/types"
	eucommon "github.com/arcology-network/eu/common"
	eth "github.com/arcology-network/eu/eth"
	intf "github.com/arcology-network/eu/interface"
	stgcommon "github.com/arcology-network/storage-committer/common"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	"github.com/arcology-network/storage-committer/type/commutative"
	univalue "github.com/arcology-network/storage-committer/type/univalue"

	evmcommon "github.com/ethereum/go-ethereum/common"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/holiman/uint256"
)

// JobSequence represents a sequence of jobs to be executed.
type JobSequence struct {
	ID           uint32 // group id
	PreTxs       []uint32
	StdMsgs      []*commontype.StandardMessage
	Results      []*eucommon.Result
	SeqAPI       intf.EthApiRouter
	RecordBuffer []*univalue.Univalue
}

func NewJobSequence(seqID uint32, tx []uint64, evmMsgs []*evmcore.Message, txHash [32]byte, api intf.EthApiRouter) *JobSequence {
	newJobSeq := &JobSequence{
		ID:     seqID,
		SeqAPI: api,
	}

	for i, evmMsg := range evmMsgs {
		newJobSeq.AppendMsg(&commontype.StandardMessage{
			ID:     tx[i],
			Native: evmMsg,
			TxHash: txHash,
		})
	}
	return newJobSeq
}

func (*JobSequence) T() *JobSequence { return &JobSequence{} }

// New creates a new JobSequence with the given ID and API router.
func (*JobSequence) New(id uint32, apiRouter intf.EthApiRouter) *JobSequence {
	return &JobSequence{
		ID:     id,
		SeqAPI: apiRouter,
	}
}

// NewFromCall creates a new JobSequence from the given call.
func (*JobSequence) NewFromCall(evmMsg *evmcore.Message, baseTxHash [32]byte, api intf.EthApiRouter) *JobSequence {
	newJobSeq := new(JobSequence).New(uint32(api.GetSerialNum(eucommon.SUB_PROCESS)), api)

	return newJobSeq.AppendMsg(&commontype.StandardMessage{
		ID:     uint64(newJobSeq.GetID()),
		Native: evmMsg,
		TxHash: newJobSeq.DeriveNewHash(baseTxHash), //api.GetEU().(interface{ TxHash() [32]byte }).TxHash()
	})
}

// GetID returns the ID of the JobSequence.
func (this *JobSequence) GetID() uint32 { return this.ID }
func (this *JobSequence) AppendMsg(msg interface{}) *JobSequence {
	this.StdMsgs = append(this.StdMsgs, msg.(*commontype.StandardMessage))
	return this
}

// DeriveNewHash derives a pseudo-random transaction hash from the given original hash and the JobSequence ID.
// It is used to help uniquely identify transactions spawned by the multiprocessor in conflict detection and resolution.
func (this *JobSequence) DeriveNewHash(original [32]byte) [32]byte {
	return sha256.Sum256(slice.Flatten([][]byte{
		codec.Bytes32(original).Encode(),
		codec.Uint32(this.ID).Encode(),
	}))
}

// Length returns the number of standard messages in the JobSequence.
func (this *JobSequence) Length() int { return len(this.StdMsgs) }

// Run executes the job sequence and returns the results. nonceOffset is used to calculate the nonce of the transaction, in
// case there is a contract deployment in the sequence.
func (this *JobSequence) Run(config *eucommon.Config, seqAPI intf.EthApiRouter, threadId uint64) ([]uint32, []*univalue.Univalue) {
	this.SeqAPI = seqAPI //.Cascade() // Create a new write tempcache for the sequence with the main router as the data source.
	this.SeqAPI.DecrementDepth()

	// Only one transaction in the sequence, no need to create a new router.
	this.Results = make([]*eucommon.Result, len(this.StdMsgs))
	if len(this.StdMsgs) == 1 {
		this.Results[0] = this.execute(this.StdMsgs[0], config, this.SeqAPI.Cascade())
		return slice.Fill(make([]uint32, len(this.Results[0].RawStateAccesses)), this.ID), this.Results[0].RawStateAccesses
	}

	for i, msg := range this.StdMsgs {
		txApi := this.SeqAPI.Cascade() // A new router whose writeCache uses the parent APIHandler's writeCache as the data source.
		txApi.DecrementDepth()         // The api router always increments the depth.  So we need to decrement it here.

		this.Results[i] = this.execute(msg, config, txApi) // Execute the message and store the result.

		// the line below modifies the tempcache in the major api as well.
		this.SeqAPI.WriteCache().(*tempcache.WriteCache).Insert(this.Results[i].RawStateAccesses) // Merge the txApi write tempcache back into the api router.
		mapi.Merge(txApi.AuxDict(), this.SeqAPI.AuxDict())                                        // The tx may generate new aux data, so merge it into the main api router.
	}

	// Get acumulated state access records from all the transactions in the sequence.
	accmulatedAccessRecords := univalue.Univalues(this.SeqAPI.WriteCache().(*tempcache.WriteCache).Export()).To(univalue.IPAccess{})
	return slice.Fill(make([]uint32, len(accmulatedAccessRecords)), this.ID), accmulatedAccessRecords
}

// GetClearedTransition returns the cleared transitions of the JobSequence.
func (this *JobSequence) GetClearedTransition() []*univalue.Univalue {
	// if idx, _ := slice.FindFirstIf(this.Results, func(v *Result) bool { return v.Err != nil }); idx < 0 {
	// 	return this.SeqAPI.WriteCache().(*tempcache.WriteCache).Export()
	// }

	trans := slice.Concate(this.Results,
		func(v *eucommon.Result) []*univalue.Univalue {
			return v.Transitions()
		},
	)

	uniqueDict := make(map[string]*univalue.Univalue)
	for _, v := range trans {
		uniqueDict[*v.GetPath()] = v
	}

	uniqueTrans := mapi.Values(uniqueDict)
	return univalue.Univalues(uniqueTrans).SortByKey()
}

// FlagConflict flags the transitions after the first conflicting transaction.
func (this *JobSequence) FlagConflict(dict map[uint32]uint64, err error) {
	// Get the first index of the first conflict transaction.
	// All the transitions after this index aren't usuable any more.
	first, _ := slice.FindFirstIf(this.Results, func(_ int, r *eucommon.Result) bool {
		_, ok := (dict)[r.TxIndex]
		return ok
	})

	// The results of the transactions after the first conflict transaction are flagged as conflicting as well.
	// Because they are potentially affected by the conflict by using the conflicting state.
	for i := first; i < len(this.Results); i++ {
		this.Results[i].Err = err
	}
}

// execute executes a standard message and returns the result.
func (this *JobSequence) execute(StdMsg *commontype.StandardMessage, config *eucommon.Config, api intf.EthApiRouter) *eucommon.Result {
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
			eucommon.NewEVMBlockContext(config),
			eucommon.NewEVMTxContext(*StdMsg.Native),
		)

	return (&eucommon.Result{
		TxIndex:          uint32(StdMsg.ID),
		TxHash:           common.IfThenDo1st(receipt != nil, func() evmcommon.Hash { return receipt.TxHash }, evmcommon.Hash{}),
		RawStateAccesses: tempcache.NewWriteCacheFilter(api.WriteCache()).ToBuffer(),
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
	// for _, v := range *this.SeqAPI.WriteCache().(*tempcache.WriteCache).Cache() {
	// 	typed := v.Value().(stgtype.Type)
	// 	amount += common.IfThen(
	// 		!v.Preexist(),
	// 		(uint64(typed.Size())/32)*uint64(v.Writes())*ethparams.SstoreSetGas,
	// 		(uint64(typed.Size())/32)*uint64(v.Writes()),
	// 	)
	// }
	return amount
}

// RefundTo refunds the specified amount from the payer to the recipient.
func (this *JobSequence) RefundTo(payer, recipent *univalue.Univalue, amount uint64) (uint64, error) {
	credit := commutative.NewU256Delta(uint256.NewInt(amount), true).(*commutative.U256)
	if _, _, _, _, err := recipent.Value().(stgcommon.Type).Set(credit, nil); err != nil {
		return 0, err
	}
	recipent.IncrementDeltaWrites(1)

	debit := commutative.NewU256Delta(uint256.NewInt(amount), false).(*commutative.U256)
	if _, _, _, _, err := payer.Value().(stgcommon.Type).Set(debit, nil); err != nil {
		return 0, err
	}
	payer.IncrementDeltaWrites(1)
	return amount, nil
}
