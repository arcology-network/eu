package common

import (
	// "github.com/arcology-network/common-lib/codec"

	"encoding/hex"
	"fmt"
	"strings"

	slice "github.com/arcology-network/common-lib/exp/slice"
	stgcommon "github.com/arcology-network/common-lib/types/storage/common"
	"github.com/arcology-network/common-lib/types/storage/univalue"
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
	RawStateAccesses []*univalue.Univalue
	immuned          []*univalue.Univalue // Won't be affect by conflicts
	Receipt          *evmTypes.Receipt
	EvmResult        *evmcore.ExecutionResult
	StdMsg           *StandardMessage
	Err              error
}

// GenGasTransition generates the gas transition for the sender from the sender's balance change. If the sender's balance deduction is
// greater than the gas used, it means a transfer has happened, the gas transition contains both the transfer and the gas deduction.
// In this case, the transition is split into two parts, one for the transfer and the other for the gas deduction. The transfer part is
// affected by conflicts, while the gas deduction part is not.
func (this *Result) GenGasTransition(rawTransition *univalue.Univalue, gasDelta *uint256.Int, isCredit bool) *univalue.Univalue {
	balanceTransition := rawTransition.Clone().(*univalue.Univalue)
	if diff := balanceTransition.Value().(stgcommon.Type).Delta().(uint256.Int); diff.Cmp(gasDelta) >= 0 {
		// transfer := diff.Sub(diff.Clone(), (*uint256.Int)(gasDelta))                            // balance - gas
		// (balanceTransition).Value().(stgcommon.Type).SetDelta((*codec.Uint256)(transfer)) // Set the transfer, Won't change the initial value.
		// (balanceTransition).Value().(stgcommon.Type).SetDeltaSign(false)
		//
		newGasTransition := balanceTransition.Clone().(*univalue.Univalue)
		newGasTransition.Value().(stgcommon.Type).SetDelta(*gasDelta)
		newGasTransition.Value().(stgcommon.Type).SetDeltaSign(isCredit)
		newGasTransition.Property.SetPersistent(true)
		return newGasTransition
	}
	return nil
}

func (this *Result) Postprocess() *Result {
	if len(this.RawStateAccesses) == 0 {
		return this
	}

	// Calculate the gas used in wei for the sender.
	gasUsedInWei := new(uint256.Int).Mul(uint256.NewInt(this.Receipt.GasUsed), uint256.NewInt(this.StdMsg.Native.GasPrice.Uint64()))

	// Find the sender's balance from the state accesses.
	_, senderBalance := slice.FindFirstIf(this.RawStateAccesses, func(_ int, v *univalue.Univalue) bool {
		return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.From[:]))
	})

	// Find the coinbase's balance from the state accesse records.
	_, coinbaseBalance := slice.FindFirstIf(this.RawStateAccesses, func(_ int, v *univalue.Univalue) bool {
		return v != nil && strings.HasSuffix(*v.GetPath(), "/balance") && strings.Contains(*v.GetPath(), hex.EncodeToString(this.Coinbase[:]))
	})

	// sender blance and coinbase balance changes should never be nil in a normal transaction. Because the sender needs to pay for the gas anyway,
	// the system will always check the sender's balance before executing the transaction and deduct the gas fee from the sender's balance.
	// The only case is where the sender balance and coinbase balance changes are nil is when the transaction is initiated by a cross chain relayer.
	// The relayer derives the transaction from a transition receipt on a different chain, and wraps it as a new transaction.
	// In this case, the relayer isn't the real sender of the transaction to pay for the gas.
	if senderBalance != nil && coinbaseBalance != nil {
		// Find the sender's balance from the state accesses and generate a transition corresponding to the gas deduction part.
		if senderGasDebit := this.GenGasTransition(*senderBalance, gasUsedInWei, false); senderGasDebit != nil {
			this.immuned = append(this.immuned, senderGasDebit)
		}

		if *(*senderBalance).GetPath() != *(*coinbaseBalance).GetPath() {
			if coinbaseGasCredit := this.GenGasTransition(*coinbaseBalance, gasUsedInWei, true); coinbaseGasCredit != nil {
				this.immuned = append(this.immuned, coinbaseGasCredit)
			}
		}
	}

	slice.Foreach(this.RawStateAccesses, func(_ int, v **univalue.Univalue) {
		if v != nil {
			return
		}

		path := (*v).GetPath()
		if strings.HasSuffix(*path, "/nonce") && strings.Contains(*path, hex.EncodeToString(this.From[:])) {
			(*v).Property.SetPersistent(true) // Won't be affect by conflicts
		}
	})

	this.RawStateAccesses = this.Transitions() // Return all the successful transitions
	return this
}

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
