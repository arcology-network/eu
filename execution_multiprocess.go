/*
 *   Copyright (c) 2025 Arcology Network

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

package eu

import (
	"crypto/sha256"

	"github.com/arcology-network/common-lib/codec"
	slice "github.com/arcology-network/common-lib/exp/slice"
	libcommon "github.com/arcology-network/common-lib/types"
	eucommon "github.com/arcology-network/eu/common"
	intf "github.com/arcology-network/eu/interface"
	workload "github.com/arcology-network/scheduler/workload"
	evmcore "github.com/ethereum/go-ethereum/core"
)

// This function is used for `Multiprocessor execution ONLY !!!`.
//
// This function converts a list of raw calls to a list of parallel job sequences. One job sequence is created for each caller.
// If there are N callers, there will be N job sequences. There sequences will be later added to a generation and executed in parallel.
func NewGenerationFromMsgs(id uint64, evmMsgs []*evmcore.Message, api intf.EthApiRouter) *workload.Generation {
	gen := workload.NewGeneration(id, uint32(len(evmMsgs)), []*workload.JobSequence{})
	slice.Foreach(evmMsgs, func(i int, msg **evmcore.Message) {
		gen.Add(NewFromEthMsg(*msg, api.GetEU().(interface{ TxHash() [32]byte }).TxHash(), api))
	})
	gen.Occurrences = gen.OccurrenceDict(gen.JobSeqs)
	api.SetSchedule(gen.Occurrences)
	return gen
}

// NewFromCall creates a new JobSequence from an ETH message with a `derived transaction hash`.
// This function is used for Multiprocessor execution ONLY !!!.
func NewFromEthMsg(evmMsg *evmcore.Message, baseTxHash [32]byte, api intf.EthApiRouter) *workload.JobSequence {
	newJobSeq := workload.JobSequence{ID: uint64(api.GetSerialNum(eucommon.SUB_PROCESS))}
	stdMsg := (&libcommon.StandardMessage{
		ID:     uint64(newJobSeq.ID),
		Native: evmMsg,
		TxHash: DeriveNewHash(baseTxHash, newJobSeq.ID), //api.GetEU().(interface{ TxHash() [32]byte }).TxHash()
	})

	// Append the new job to the job sequence
	newJobSeq.Jobs = append(newJobSeq.Jobs, &workload.Job{
		ID:     stdMsg.ID,
		StdMsg: stdMsg,
		Result: &workload.Result{},
	})
	return &newJobSeq
}

func DeriveNewHash(original [32]byte, seed uint64) [32]byte {
	return sha256.Sum256(slice.Flatten([][]byte{
		codec.Bytes32(original).Encode(),
		codec.Uint64(seed).Encode(),
	}))
}
