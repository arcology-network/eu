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

package eth

import (
	intf "github.com/arcology-network/eu/interface"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/ethereum/go-ethereum/common"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	evmtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/params"
	uint256 "github.com/holiman/uint256"
)

// Arcology implementation of Eth ImplStateDB interfaces.
type ImplStateDB struct {
	refund           uint64
	txHash           evmcommon.Hash
	tid              uint64 // tx id
	logs             map[evmcommon.Hash][]*evmtypes.Log
	transientStorage transientStorage
	api              intf.EthApiRouter
}

func NewImplStateDB(api intf.EthApiRouter) *ImplStateDB {
	return &ImplStateDB{
		logs:             make(map[evmcommon.Hash][]*evmtypes.Log),
		api:              api,
		transientStorage: newTransientStorage(),
	}
}

func (this *ImplStateDB) CreateAccount(addr evmcommon.Address) {
	createAccount(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
}

func (this *ImplStateDB) AddBalance(addr evmcommon.Address, amount *uint256.Int) {
	this.updateBalance(addr, amount, true) // POSITIVE
}

func (this *ImplStateDB) SubBalance(addr evmcommon.Address, amount *uint256.Int) {
	this.updateBalance(addr, amount, false) // NEGATIVE
}

// A helper function to update the balance of an account, not part of the original interface
// Its Ethereum counterpart is in the (s *stateObject) setBalance(amount *uint256.Int) function.
func (this *ImplStateDB) updateBalance(addr evmcommon.Address, amount *uint256.Int, isPositive bool) {
	if !this.Exist(addr) {
		createAccount(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
	}

	if amount.IsZero() {
		return
	}

	delta := commutative.NewU256Delta(amount, isPositive) // Create a delta
	if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.tid, getBalancePath(this.api.WriteCache().(*tempcache.WriteCache), addr), delta); err != nil {
		panic("Error: Failed to updateBalance() with delta")
	}
}

func (this *ImplStateDB) GetBalance(addr evmcommon.Address) *uint256.Int {
	if !this.Exist(addr) {
		return uint256.NewInt(0)
	}

	value, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.tid, getBalancePath(this.api.WriteCache().(*tempcache.WriteCache), addr), new(commutative.U256))
	v := value.(uint256.Int)
	return &v // v.(*commutative.U256).Value().(*big.Int)
}

func (this *ImplStateDB) PeekBalance(addr evmcommon.Address) *uint256.Int {
	if !this.Exist(addr) {
		return uint256.NewInt(0)
	}

	value, _ := this.api.WriteCache().(*tempcache.WriteCache).Peek(getBalancePath(this.api.WriteCache().(*tempcache.WriteCache), addr), new(commutative.U256))
	v := value.(uint256.Int)
	return &v
}

// NOTE: No function should ever call this function, except the initializer.
// The balance is always updated with a delta and the actual balance is calculated
// when it is read or at commit time !!!
func (this *ImplStateDB) SetBalance(addr evmcommon.Address, amount *uint256.Int) {
	this.updateBalance(addr, amount, true)
}

func (this *ImplStateDB) GetNonce(addr evmcommon.Address) uint64 {
	if !this.Exist(addr) {
		return 0
	}

	nonce, _ := this.api.WriteCache().(*tempcache.WriteCache).Peek(getNoncePath(this.api.WriteCache().(*tempcache.WriteCache), addr), new(commutative.Uint64))
	return nonce.(uint64) + this.CalculateNonceOffset(addr, nonce.(uint64)) // Add the nonce offset
}

func (this *ImplStateDB) SetNonce(addr evmcommon.Address, nonce uint64) {
	if !this.Exist(addr) {
		createAccount(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
	}
	// fmt.Println("SetNonce:", addr, ":", nonce)

	// This original implementation will set the nonce to the given value, but here we just write the nonce delta, which is 1 to the tempcache, becuase the nonce increment is always 1
	// This is Arcology's way to handle the nonce, and the actual nonce will be calculated when it is read or at commit time.
	if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.tid, getNoncePath(this.api.WriteCache().(*tempcache.WriteCache), addr), commutative.NewUint64Delta(1)); err != nil {
		panic(err)
	}
}

func (this *ImplStateDB) GetCodeHash(addr evmcommon.Address) evmcommon.Hash {
	code := this.GetCode(addr)
	if len(code) == 0 {
		return evmcommon.Hash{}
	} else {
		return crypto.Keccak256Hash(code)
	}
}

func (this *ImplStateDB) GetCode(addr evmcommon.Address) []byte {
	if !this.Exist(addr) {
		return nil
	}

	value, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.tid, getCodePath(this.api.WriteCache().(*tempcache.WriteCache), addr), new(noncommutative.Bytes))
	if value == nil {
		return []byte{}
	}
	return value.([]byte)

}

func (this *ImplStateDB) SetCode(addr evmcommon.Address, code []byte) {
	if !this.Exist(addr) {
		createAccount(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
	}

	if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.tid, getCodePath(this.api.WriteCache().(*tempcache.WriteCache), addr), noncommutative.NewBytes(code)); err != nil {
		panic(err)
	}
}

func (this *ImplStateDB) SelfDestruct(addr evmcommon.Address)           { return }
func (this *ImplStateDB) HasSelfDestructed(addr evmcommon.Address) bool { return false }
func (this *ImplStateDB) Selfdestruct6780(common.Address)               {}

func (this *ImplStateDB) GetCodeSize(addr evmcommon.Address) int                          { return len(this.GetCode(addr)) }
func (this *ImplStateDB) AddRefund(amount uint64)                                         { this.refund += amount }
func (this *ImplStateDB) SubRefund(amount uint64)                                         { this.refund -= amount }
func (this *ImplStateDB) GetRefund() uint64                                               { return this.refund }
func (this *ImplStateDB) RevertToSnapshot(id int)                                         {}
func (this *ImplStateDB) Snapshot() int                                                   { return 0 }
func (this *ImplStateDB) AddPreimage(hash evmcommon.Hash, preimage []byte)                {}
func (this *ImplStateDB) AddAddressToAccessList(addr evmcommon.Address)                   {} // Do nothing.
func (this *ImplStateDB) AddSlotToAccessList(addr evmcommon.Address, slot evmcommon.Hash) {}

// func (this *ImplStateDB) Set(eac EthAccountCache, esc EthStorageCache)                    {} // TODO

// GetCommittedState retrieves the value associated with the specific key
// without any mutations caused in the current execution.
func (this *ImplStateDB) GetCommittedState(addr evmcommon.Address, key evmcommon.Hash) evmcommon.Hash {
	if value, _ := this.api.WriteCache().(*tempcache.WriteCache).ReadCommitted(this.tid, getStorageKeyPath(this.api, addr, key), new(noncommutative.Bytes)); value != nil {
		// v, _, _ := value.(interfaces.Type).Get()
		return evmcommon.BytesToHash(value.([]byte))
	}
	return evmcommon.Hash{}
}

func (this *ImplStateDB) GetState(addr evmcommon.Address, key evmcommon.Hash) evmcommon.Hash {
	if value, _, _ := this.api.WriteCache().(*tempcache.WriteCache).Read(this.tid, getStorageKeyPath(this.api, addr, key), new(noncommutative.Bytes)); value != nil {
		return evmcommon.BytesToHash(value.([]byte))
	}
	return evmcommon.Hash{}
}

func (this *ImplStateDB) SetState(addr evmcommon.Address, key, value evmcommon.Hash) {
	if !this.Exist(addr) {
		createAccount(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
	}

	path := getStorageKeyPath(this.api, addr, key)
	if _, err := this.api.WriteCache().(*tempcache.WriteCache).Write(this.tid, path, noncommutative.NewBytes(value.Bytes())); err != nil {
		panic(err)
	}
}

func (this *ImplStateDB) Exist(addr evmcommon.Address) bool {
	flag := accountExist(this.api.WriteCache().(*tempcache.WriteCache), addr, this.tid)
	// fmt.Println(addr, flag)
	return flag
}

func (this *ImplStateDB) Empty(addr evmcommon.Address) bool {
	return (!this.Exist(addr)) || (this.PeekBalance(addr).BitLen() == 0 && this.GetNonce(addr) == 0 && this.GetCodeSize(addr) == 0)
}

// SetTransientState sets transient storage for a given account. It
// adds the change to the journal so that it can be rolled back
// to its previous value if there is a revert.
func (s *ImplStateDB) SetTransientState(addr evmcommon.Address, key, value evmcommon.Hash) {
	prev := s.GetTransientState(addr, key)
	if prev == value {
		return
	}
	s.setTransientState(addr, key, value)
}

// setTransientState is a lower level setter for transient storage. It
// is called during a revert to prevent modifications to the journal.
func (s *ImplStateDB) setTransientState(addr evmcommon.Address, key, value evmcommon.Hash) {
	s.transientStorage.Set(addr, key, value)
}

// GetTransientState gets transient storage for a given account.
func (s *ImplStateDB) GetTransientState(addr evmcommon.Address, key evmcommon.Hash) evmcommon.Hash {
	return s.transientStorage.Get(addr, key)
}
func (this *ImplStateDB) Prepare(rules params.Rules, sender, coinbase common.Address, dest *common.Address, precompiles []common.Address, txAccesses types.AccessList) {
	//Do nothing
}

func (this *ImplStateDB) AddLog(log *evmtypes.Log) {
	this.logs[this.txHash] = append(this.logs[this.txHash], log)
}

func (this *ImplStateDB) ForEachStorage(addr evmcommon.Address, f func(evmcommon.Hash, evmcommon.Hash) bool) error {
	return nil
}

func (this *ImplStateDB) PrepareAccessList(sender evmcommon.Address, dest *evmcommon.Address, precompiles []evmcommon.Address, txAccesses evmtypes.AccessList) {
	// Do nothing.
}

func (this *ImplStateDB) AddressInAccessList(addr evmcommon.Address) bool { return true }

func (this *ImplStateDB) SlotInAccessList(addr evmcommon.Address, slot evmcommon.Hash) (addressOk bool, slotOk bool) {
	return true, true
}

func (this *ImplStateDB) PrepareFormer(txHash, bhash evmcommon.Hash, ti uint64) {
	this.refund = 0
	this.txHash = txHash
	this.tid = ti
	this.logs = make(map[evmcommon.Hash][]*evmtypes.Log)
}

func (this *ImplStateDB) GetLogs(hash evmcommon.Hash) []*evmtypes.Log {
	return this.logs[hash]
}
