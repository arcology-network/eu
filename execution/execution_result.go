package execution

import (
	// "github.com/arcology-network/common-lib/codec"

	"encoding/hex"
	"fmt"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	ccurlinterfaces "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/univalue"
	adaptorcommon "github.com/arcology-network/vm-adaptor/common"
	evmcore "github.com/ethereum/go-ethereum/core"
	evmTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

type Result struct {
	GroupID          uint32 // == Group ID
	TxIndex          uint32
	TxHash           [32]byte
	From             [20]byte
	Coinbase         [20]byte
	rawStateAccesses []ccurlinterfaces.Univalue
	immuned          []ccurlinterfaces.Univalue
	Receipt          *evmTypes.Receipt
	EvmResult        *evmcore.ExecutionResult
	stdMsg           *adaptorcommon.StandardMessage
	Err              error
}

// The tx sender has to pay the tx fees regardless the execution status.
func (this *Result) GenGasTransition(rawTransition ccurlinterfaces.Univalue, gasDelta *uint256.Int, isCredit bool) ccurlinterfaces.Univalue {
	balanceTransition := rawTransition.Clone().(ccurlinterfaces.Univalue)
	if diff := balanceTransition.Value().(ccurlinterfaces.Type).Delta().(uint256.Int); diff.Cmp(gasDelta) >= 0 {
		// transfer := diff.Sub(diff.Clone(), (*uint256.Int)(gasDelta))                            // balance - gas
		// (balanceTransition).Value().(ccurlinterfaces.Type).SetDelta((*codec.Uint256)(transfer)) // Set the transfer, Won't change the initial value.
		// (balanceTransition).Value().(ccurlinterfaces.Type).SetDeltaSign(false)
		//
		newGasTransition := balanceTransition.Clone().(ccurlinterfaces.Univalue)
		newGasTransition.Value().(ccurlinterfaces.Type).SetDelta(*gasDelta)
		newGasTransition.Value().(ccurlinterfaces.Type).SetDeltaSign(isCredit)
		newGasTransition.GetUnimeta().(*univalue.Unimeta).SetPersistent(true)
		return newGasTransition
	}
	return nil
}

func (this *Result) Postprocess() *Result {
	if len(this.rawStateAccesses) == 0 {
		return this
	}

	_, senderBalance := common.FindFirstIf(this.rawStateAccesses, func(v ccurlinterfaces.Univalue) bool {
		return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.From[:]))
	})

	gasUsedInWei := uint256.NewInt(1).Mul(uint256.NewInt(this.Receipt.GasUsed), uint256.NewInt(this.stdMsg.Native.GasPrice.Uint64()))
	if senderGasDebit := this.GenGasTransition(*senderBalance, gasUsedInWei, false); senderGasDebit != nil {
		this.immuned = append(this.immuned, senderGasDebit)
	}

	_, coinbaseBalance := common.FindFirstIf(this.rawStateAccesses, func(v ccurlinterfaces.Univalue) bool {
		return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.Coinbase[:]))
	})

	if *(*senderBalance).GetPath() != *(*coinbaseBalance).GetPath() {
		if coinbaseGasCredit := this.GenGasTransition(*coinbaseBalance, gasUsedInWei, true); coinbaseGasCredit != nil {
			this.immuned = append(this.immuned, coinbaseGasCredit)
		}
	}

	common.Foreach(this.rawStateAccesses, func(v *ccurlinterfaces.Univalue, _ int) {
		if v != nil {
			return
		}

		path := (*v).GetPath()
		if strings.HasSuffix(*path, "/nonce") && strings.Contains(*path, hex.EncodeToString(this.From[:])) {
			(*v).GetUnimeta().(*univalue.Unimeta).SetPersistent(true) // Won't be affect by conflicts
		}
	})

	this.rawStateAccesses = this.Transitions() // Return all the successful transitions
	return this
}

func (this *Result) Transitions() []ccurlinterfaces.Univalue {
	if this.Err != nil {
		return this.immuned //.MoveIf(&this.rawStateAccesses, func(v ccurlinterfaces.Univalue) bool { return v.Persistent() })
	}
	return this.rawStateAccesses
}

func (this *Result) Print() {
	// fmt.Println("GroupID: ", this.GroupID)
	fmt.Println("TxIndex: ", this.TxIndex)
	fmt.Println("TxHash: ", this.TxHash)
	fmt.Println()
	indexer.Univalues(this.rawStateAccesses).Print()
	fmt.Println("Error: ", this.Err)
}

type Results []*Result

func (this Results) Print() {
	fmt.Println("Execution Results: ")
	for _, v := range this {
		v.Print()
		fmt.Println()
	}
}
