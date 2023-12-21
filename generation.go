package eu

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"

	// evmeu "github.com/arcology-network/vm-adaptor"
	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ccurlinterfaces "github.com/arcology-network/concurrenturl/interfaces"
	eucommon "github.com/arcology-network/eu/common"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
)

// APIs under the concurrency namespace
type Generation struct {
	ID         uint32
	numThreads uint8
	jobs       []*JobSequence // para jobs
}

func NewGeneration(id uint32, numThreads uint8, jobs []*JobSequence) *Generation {
	return &Generation{
		ID:         id,
		numThreads: numThreads,
		jobs:       jobs,
	}
}

// func (this *Generation) BranchID() uint32 { return this.branchID }
func (this *Generation) Length() uint64       { return uint64(len(this.jobs)) }
func (this *Generation) Jobs() []*JobSequence { return this.jobs }

func (this *Generation) At(idx uint64) *JobSequence {
	return common.IfThenDo1st(idx < uint64(len(this.jobs)), func() *JobSequence { return this.jobs[idx] }, nil)
}

func (this *Generation) Add(job *JobSequence) bool {
	this.jobs = append(this.jobs, job)
	return true
}

func (this *Generation) Run(parentApiRouter adaptorcommon.EthApiRouter) []interfaces.Univalue {
	config := eucommon.NewConfig().SetCoinbase(parentApiRouter.Coinbase())

	groupIDs := make([][]uint32, len(this.jobs))
	records := make([][]ccurlinterfaces.Univalue, len(this.jobs))
	// t0 := time.Now()
	worker := func(start, end, idx int, args ...interface{}) {
		// for i := 0; i < len(this.jobs); i++ {
		for i := start; i < end; i++ {
			groupIDs[i], records[i] = this.jobs[i].Run(config, parentApiRouter)
			//	indexer.Univalues(records[i]).Sort(groupIDs[i]) // Debugging only
		}
	}
	common.ParallelWorker(len(this.jobs), int(this.numThreads), worker)
	// fmt.Println(time.Since(t0))

	txDict, groupDict, _ := this.Detect(groupIDs, records).ToDict()
	return common.Concate(this.jobs, func(seq *JobSequence) []interfaces.Univalue {
		if _, ok := (*groupDict)[(*seq).ID]; ok {
			(*seq).FlagConflict(txDict, errors.New(ccurlcommon.WARN_ACCESS_CONFLICT))
		}
		return (*seq).GetClearedTransition()
	})
}

func (*Generation) Detect(groupIDs [][]uint32, records [][]interfaces.Univalue) arbitrator.Conflicts {
	if len(records) == 1 {
		return arbitrator.Conflicts{}
	}
	return arbitrator.Conflicts((&arbitrator.Arbitrator{}).Detect(common.Flatten(groupIDs), common.Flatten(records)))
}

func (this *Generation) Clear() uint64 {
	length := len(this.jobs)
	this.jobs = this.jobs[:0]
	return uint64(length)
}
