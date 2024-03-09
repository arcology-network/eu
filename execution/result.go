package execution

import (
	// "github.com/arcology-network/common-lib/codec"

	"encoding/hex"
	"fmt"
	"strings"

	slice "github.com/arcology-network/common-lib/exp/slice"
	eucommon "github.com/arcology-network/eu/common"
	ccurlintf "github.com/arcology-network/storage-committer/interfaces"
	"github.com/arcology-network/storage-committer/univalue"
	evmcore "github.com/ethereum/go-ethereum/core"
	ethcoretypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/holiman/uint256"
)

type Result struct {
	GroupID          uint32 // == Group ID
	TxIndex          uint32
	TxHash           [32]byte
	From             [20]byte
	Coinbase         [20]byte
	RawStateAccesses []*univalue.Univalue
	immuned          []*univalue.Univalue //These transitions will take effect if the execution fails.
	Receipt          *ethcoretypes.Receipt
	EvmResult        *evmcore.ExecutionResult
	StdMsg           *eucommon.StandardMessage
	Err              error
}

// The tx sender has to pay the tx fees regardless the execution status. This function deducts the gas fee from the sender's balance
// change and generates a new transition for that.
func (this *Result) GenGasTransition(balanceTransition *univalue.Univalue, gasDelta *uint256.Int, isCredit bool) *univalue.Univalue {
	totalDelta := balanceTransition.Value().(ccurlintf.Type).Delta().(uint256.Int)
	if totalDelta.Cmp(gasDelta) == 0 { // Balance change == gas fee paid.
		balanceTransition.Property.SetPersistent(true) // Won't be affect by conflicts
		return balanceTransition
	}

	// Separate the gas fee from the balance change and generate a new transition for that.
	gasTransition := balanceTransition.Clone().(*univalue.Univalue)
	gasTransition.Value().(ccurlintf.Type).SetDelta(*gasDelta)    // Set the gas fee.
	gasTransition.Value().(ccurlintf.Type).SetDeltaSign(isCredit) // Negative for the sender, positive for the coinbase.
	gasTransition.Property.SetPersistent(true)
	return gasTransition
}

func (this *Result) Postprocess() *Result {
	if len(this.RawStateAccesses) == 0 {
		return this
	}

	// The sender isn't the coinbase.
	if this.From != this.Coinbase {
		_, senderBalance := slice.FindFirstIf(this.RawStateAccesses, func(v *univalue.Univalue) bool { //It includes the gas fee and possible transfers.
			return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.From[:]))
		})

		_, coinbaseBalance := slice.FindFirstIf(this.RawStateAccesses, func(v *univalue.Univalue) bool {
			return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.Coinbase[:]))
		})

		// Usually, neither the sender balance nor the coinbase balance can't be nil unless the transaction
		// is a L1->L2 transaction derived from a transaction receipt and the network is in a L2 setup.
		if senderBalance != nil && coinbaseBalance != nil {
			// Separate the gas fee from the balance change and generate a new transition for that. It will be immune to the execution status.
			gasUsedInWei := new(uint256.Int).Mul(uint256.NewInt(this.Receipt.GasUsed), uint256.NewInt(this.StdMsg.Native.GasPrice.Uint64()))
			if senderGasDebit := this.GenGasTransition(*senderBalance, gasUsedInWei, false); senderGasDebit != nil {
				this.immuned = append(this.immuned, senderGasDebit)
			}

			if coinbaseGasCredit := this.GenGasTransition(*coinbaseBalance, gasUsedInWei, true); coinbaseGasCredit != nil {
				this.immuned = append(this.immuned, coinbaseGasCredit)
			}
		}
	}

	_, senderNonce := slice.FindFirstIf(this.RawStateAccesses, func(v *univalue.Univalue) bool {
		return strings.HasSuffix(*v.GetPath(), "/nonce") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.From[:]))
	})

	if senderNonce != nil {
		(*senderNonce).Property.SetPersistent(true)       // Won't be affect by conflicts either
		this.immuned = append(this.immuned, *senderNonce) // Add the nonce transition to the immune list even if the execution is unsuccessful.
	}
	this.RawStateAccesses = this.Transitions() // Return all the successful transitions
	return this
}

// If the execution is unsuccessful, only keep the transitions that are immune to failures.
func (this *Result) Transitions() []*univalue.Univalue {
	if this.Err != nil {
		return this.immuned //.MoveIf(&this.RawStateAccesses, func(v *univalue.Univalue) bool { return v.Persistent() })
	}
	return this.RawStateAccesses
}

func (this *Result) Print() {
	// fmt.Println("GroupID: ", this.GroupID)
	fmt.Println("TxIndex: ", this.TxIndex)
	fmt.Println("TxHash: ", this.TxHash)
	fmt.Println()
	univalue.Univalues(this.RawStateAccesses).Print()
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
