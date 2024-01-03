package execution

import (
	"errors"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/eu/execution"

	arbitrator "github.com/arcology-network/concurrenturl/arbitrator"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/univalue"
	intf "github.com/arcology-network/vm-adaptor/interface"
)

// APIs under the concurrency namespace
type Generation struct {
	ID         uint32
	numThreads uint8
	jobSeqs    []*JobSequence // para jobSeqs
}

func NewGeneration(id uint32, numThreads uint8, jobSeqs []*JobSequence) *Generation {
	return &Generation{
		ID:         id,
		numThreads: numThreads,
		jobSeqs:    jobSeqs,
	}
}

// func (this *Generation) BranchID() uint32 { return this.branchID }
func (this *Generation) Length() uint64     { return uint64(len(this.jobSeqs)) }
func (this *Generation) JobT() *JobSequence { return &JobSequence{} }
func (this *Generation) JobSeqs() []*JobSequence {
	return common.To[*JobSequence, *JobSequence](this.jobSeqs)
}

func (this *Generation) At(idx uint64) *JobSequence {
	return common.IfThenDo1st(idx < uint64(len(this.jobSeqs)), func() *JobSequence { return this.jobSeqs[idx] }, nil)
}

func (this *Generation) New(id uint32, numThreads uint8, jobSeqs []*JobSequence) *Generation {
	return NewGeneration(id, numThreads, common.To[*JobSequence, *JobSequence](jobSeqs))
}

func (this *Generation) Add(job *JobSequence) bool {
	this.jobSeqs = append(this.jobSeqs, job)
	return true
}

func (this *Generation) Run(parentApiRouter intf.EthApiRouter) []*univalue.Univalue {
	config := execution.NewConfig().SetCoinbase(parentApiRouter.Coinbase())

	groupIDs := make([][]uint32, len(this.jobSeqs))
	records := make([][]*univalue.Univalue, len(this.jobSeqs))
	// t0 := time.Now()
	worker := func(start, end, idx int, args ...interface{}) {
		// for i := 0; i < len(this.jobSeqs); i++ {
		for i := start; i < end; i++ {
			groupIDs[i], records[i] = this.jobSeqs[i].Run(config, parentApiRouter)
			//	indexer.Univalues(records[i]).Sort(groupIDs[i]) // Debugging only
		}
	}
	common.ParallelWorker(len(this.jobSeqs), int(this.numThreads), worker)
	// fmt.Println(time.Since(t0))

	txDict, groupDict, _ := this.Detect(groupIDs, records).ToDict()
	return common.Concate(this.jobSeqs, func(seq *JobSequence) []*univalue.Univalue {
		if _, ok := (*groupDict)[(*seq).ID]; ok {
			(*seq).FlagConflict(txDict, errors.New(ccurlcommon.WARN_ACCESS_CONFLICT))
		}
		return (*seq).GetClearedTransition()
	})
}

func (*Generation) Detect(groupIDs [][]uint32, records [][]*univalue.Univalue) arbitrator.Conflicts {
	if len(records) == 1 {
		return arbitrator.Conflicts{}
	}
	return arbitrator.Conflicts((&arbitrator.Arbitrator{}).Detect(common.Flatten(groupIDs), common.Flatten(records)))
}

func (this *Generation) Clear() uint64 {
	length := len(this.jobSeqs)
	this.jobSeqs = this.jobSeqs[:0]
	return uint64(length)
}
