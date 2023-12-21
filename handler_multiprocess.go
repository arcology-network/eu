package eu

import (
	"math"
	"sync/atomic"

	"github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	ethcommon "github.com/ethereum/go-ethereum/common"
	evmcore "github.com/ethereum/go-ethereum/core"
	"github.com/holiman/uint256"

	eucommon "github.com/arcology-network/eu/common"
	"github.com/arcology-network/vm-adaptor/abi"

	api "github.com/arcology-network/vm-adaptor/api"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
)

// APIs under the concurrency namespace
type MultiprocessHandlers struct {
	*api.BaseHandlers
	erros   []error
	jobseqs []*JobSequence
}

func NewMultiprocessHandlers(ethApiRouter adaptorcommon.EthApiRouter) *MultiprocessHandlers {
	handler := &MultiprocessHandlers{
		erros:   []error{},
		jobseqs: []*JobSequence{},
	}
	handler.BaseHandlers = api.NewBaseHandlers(ethApiRouter, handler)
	return handler
}

func (this *MultiprocessHandlers) Address() [20]byte { return adaptorcommon.MULTIPROCESS_HANDLER }

func (this *MultiprocessHandlers) Run(caller [20]byte, input []byte) ([]byte, bool, int64) {
	if atomic.AddUint64(&adaptorcommon.TotalSubProcesses, 1); !this.Api().CheckRuntimeConstrains() {
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
	length = common.Min(adaptorcommon.MAX_VM_INSTANCES, length)

	if !successful {
		return []byte{}, successful, fee
	}

	generation := NewGeneration(0, threads, []*JobSequence{})
	fees := make([]int64, length)
	this.erros = make([]error, length)
	this.jobseqs = make([]*JobSequence, length)

	for i := uint64(0); i < length; i++ {
		funCall, successful, fee := this.GetByIndex(path, uint64(i))
		if fees[i] = fee; successful {
			this.jobseqs[i], this.erros[i] = this.toJobSeq(funCall)
		}
		generation.Add(this.jobseqs[i])
	}
	transitions := generation.Run(this.Api())

	// Sub processes may have been spawned during the eucommon, recheck it.
	if !this.Api().CheckRuntimeConstrains() {
		return []byte{}, false, fee
	}

	// Unify tx IDs c
	mainTxID := uint32(this.Api().GetEU().(*EU).Message().ID)
	common.Foreach(transitions, func(v *interfaces.Univalue, _ int) { (*v).SetTx(mainTxID) })

	this.Api().Ccurl().WriteCache().AddTransitions(transitions) // Merge the write cache to the main cache
	return []byte{}, true, common.Sum[int64](fees)
}

func (this *MultiprocessHandlers) toJobSeq(input []byte) (*JobSequence, error) {
	gasLimit, value, calleeAddr, funCall, err := abi.Parse4(input,
		uint64(0), 1, 32,
		uint256.NewInt(0), 1, 32,
		[20]byte{}, 1, 32,
		[]byte{}, 2, math.MaxInt64)

	if err != nil {
		return nil, err
	}

	newJobSeq := &JobSequence{
		ID:        uint32(this.BaseHandlers.Api().GetSerialNum(adaptorcommon.SUB_PROCESS)),
		ApiRouter: this.BaseHandlers.Api(),
	}

	transfer := value.ToBig()
	addr := ethcommon.Address(calleeAddr)
	evmMsg := evmcore.NewMessage( // Build the message
		this.BaseHandlers.Api().Origin(),
		&addr,
		0,
		transfer, // Amount to transfer
		gasLimit,
		this.BaseHandlers.Api().GetEU().(*EU).Message().Native.GasPrice, // gas price
		funCall,
		nil,
		false, // Don't checking nonce
	)

	stdMsg := &eucommon.StandardMessage{
		ID:     uint64(newJobSeq.ID), // this is the problem !!!!
		Native: &evmMsg,
		TxHash: newJobSeq.DeriveNewHash(this.BaseHandlers.Api().GetEU().(*EU).Message().TxHash),
	}

	newJobSeq.StdMsgs = []*eucommon.StandardMessage{stdMsg}
	return newJobSeq, nil
}
