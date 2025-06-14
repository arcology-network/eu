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

package execution

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"
	slice "github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"
	intf "github.com/arcology-network/eu/interface"
	scheduler "github.com/arcology-network/scheduler"
	arbitrator "github.com/arcology-network/scheduler/arbitrator"
	stgcommon "github.com/arcology-network/storage-committer/common"
	univalue "github.com/arcology-network/storage-committer/type/univalue"
	evmcore "github.com/ethereum/go-ethereum/core"
)

// APIs under the concurrency namespace
type Generation struct {
	ID          uint64
	numThreads  uint8
	jobSeqs     []*JobSequence // para jobSeqs
	occurrences *map[string]int
}

func (*Generation) OccurrenceDict(jobSeqs []*JobSequence) *map[string]int {
	occurrences := map[string]int{}
	for _, seq := range jobSeqs {
		for _, msg := range seq.StdMsgs {
			occurrences[scheduler.ToKey(msg)]++ // Only count the first one if found
			break
		}
	}
	return &occurrences
}

func NewGeneration(id uint64, numThreads uint8, jobSeqs []*JobSequence) *Generation {
	gen := &Generation{
		ID:         id,
		numThreads: numThreads,
		jobSeqs:    jobSeqs,
	}
	gen.occurrences = gen.OccurrenceDict(jobSeqs)
	return gen
}

// This function is used for Multiprocessor execution ONLY !!!.
// This function converts a list of raw calls to a list of parallel job sequences. One job sequence is created for each caller.
// If there are N callers, there will be N job sequences. There sequences will be later added to a generation and executed in parallel.
func NewGenerationFromMsgs(id uint64, numThreads uint8, evmMsgs []*evmcore.Message, api intf.EthApiRouter) *Generation {
	gen := NewGeneration(id, uint8(len(evmMsgs)), []*JobSequence{})
	slice.Foreach(evmMsgs, func(i int, msg **evmcore.Message) {
		gen.Add(new(JobSequence).NewFromCall(*msg, api.GetEU().(interface{ TxHash() [32]byte }).TxHash(), api))
	})
	gen.occurrences = gen.OccurrenceDict(gen.jobSeqs)
	api.SetSchedule(gen.occurrences)
	return gen
}

func (this *Generation) Length() uint64     { return uint64(len(this.jobSeqs)) }
func (this *Generation) JobT() *JobSequence { return &JobSequence{} }
func (this *Generation) JobSeqs() []*JobSequence {
	return slice.To[*JobSequence, *JobSequence](this.jobSeqs)
}

func (this *Generation) At(idx uint64) *JobSequence {
	return common.IfThenDo1st(idx < uint64(len(this.jobSeqs)), func() *JobSequence { return this.jobSeqs[idx] }, nil)
}

func (*Generation) New(id uint64, numThreads uint8, jobSeqs []*JobSequence) *Generation {
	return NewGeneration(id, numThreads, slice.To[*JobSequence, *JobSequence](jobSeqs))
}

func (this *Generation) Add(job *JobSequence) bool {
	this.jobSeqs = append(this.jobSeqs, job)
	return true
}

// The run function executes the job sequences in parallel and returns the results in a single slice.
// The blockAPI is used to access the state data. For external transaction execution, the blockAPI has
// all the state data from the last block. For the spawned transaction execution, the blockAPI has the state data
// of it parent thread up to the point of the point of the thread creation. The child thread uses the state data of the parent
// thread to create a state snapshot for itself. Eventually, all the state changes generated by the child threads will be
// merged back into the parent thread.
//
// But when a child thread is trying to deploy a contract, it needs to increment the nonce of the caller contract and
// the nonce is a global counter for the account. Since there is no inter-thread communication, the child will increment
// the nonce of the parent thread by itself independently. Different child threads may deploy their contracts at the same address.

// This isn't a problem for the external transaction execution, the conflict detector will find it out and revert the transactions.
// But for the spawned transaction execution, sometimes we need to deploy some temporary contracts to do their jobs, and certainly we
// don't want to cause any conflict. That is why we need to give different nonceOffset to different child threads, so they can deploy
// their contracts at different addresses.

func (this *Generation) Execute(execCoinbase interface{}, blockAPI intf.EthApiRouter) []*univalue.Univalue {
	config := execCoinbase.(*eucommon.Config)

	seqIDs := make([][]uint64, len(this.jobSeqs))
	records := make([][]*univalue.Univalue, len(this.jobSeqs))

	// Execute the job sequences in parallel. All the access records from the same sequence share
	// the same sequence ID. The sequence ID is used to detect the conflicts between different sequences.
	slice.ParallelForeach(this.jobSeqs, int(this.numThreads), func(i int, _ **JobSequence) {
		seqIDs[i], records[i] = this.jobSeqs[i].Run(config, blockAPI.Cascade(), uint64(i))
	})

	conflictInfo := this.Detect(seqIDs, records)
	txDict, seqDict, _ := conflictInfo.ToDict()

	// Mark the conflicts in the job sequences.
	cleanTrans := slice.Concate(this.jobSeqs, func(seq *JobSequence) []*univalue.Univalue {
		if _, ok := seqDict[(*seq).ID]; ok { // Check if the sequence ID is in the conflict list.
			(*seq).FlagConflict(txDict, errors.New(stgcommon.WARN_ACCESS_CONFLICT))
		}
		return (*seq).GetClearedTransition() // Return the conflict-free transitions
	})
	return cleanTrans
}

// There needs to be a sequence ID for each transaction in the sequence, not just the transaction ID because
// multiple transactions may be in the same sequence and they may have the same transaction ID.
func (*Generation) Detect(seqIDs [][]uint64, records [][]*univalue.Univalue) arbitrator.Conflicts {
	if len(records) == 1 {
		return arbitrator.Conflicts{}
	}

	return arbitrator.Conflicts(arbitrator.NewArbitrator().InsertAndDetect(slice.Flatten(seqIDs), slice.Flatten(records)))
}

func (this *Generation) Clear() uint64 {
	length := len(this.jobSeqs)
	this.jobSeqs = this.jobSeqs[:0]
	return uint64(length)
}
