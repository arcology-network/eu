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
	commonlib "github.com/arcology-network/common-lib/common"
	intf "github.com/arcology-network/eu/interface"

	stgcommon "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

type EthPathBuilder struct {
	// ccurl cache.WriteCache
}

func (this *EthPathBuilder) AccountExist(writeCache *cache.WriteCache, account evmcommon.Address, tid uint64) bool {
	return writeCache.IfExists(this.AccountRootPath(account))
}

func (this *EthPathBuilder) AccountRootPath(account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/")
}

func (this *EthPathBuilder) StorageRootPath(account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/storage/native/")
}

func (this *EthPathBuilder) BalancePath(account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/balance")
}

func (this *EthPathBuilder) NoncePath(account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/nonce")
}

func (this *EthPathBuilder) CodePath(account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/code")
}

func getAccountRootPath(writeCache *cache.WriteCache, account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/")
}

func getStorageRootPath(writeCache *cache.WriteCache, account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/storage/native/")
}

func getLocalStorageKeyPath(api intf.EthApiRouter, account evmcommon.Address, key evmcommon.Hash) string {
	return stgcommon.ETH10_ACCOUNT_PREFIX + hexutil.Encode(account[:]) + "/storage/native/local/" + "0"
}

func getStorageKeyPath(api intf.EthApiRouter, account evmcommon.Address, key evmcommon.Hash) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/storage/native/", key.Hex())
}

func getBalancePath(writeCache *cache.WriteCache, account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/balance")
}

func getNoncePath(writeCache *cache.WriteCache, account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/nonce")
}

func getCodePath(writeCache *cache.WriteCache, account evmcommon.Address) string {
	return commonlib.StrCat(stgcommon.ETH10_ACCOUNT_PREFIX, hexutil.Encode(account[:]), "/code")
}

func accountExist(writeCache *cache.WriteCache, account evmcommon.Address, tid uint64) bool {
	return writeCache.IfExists(getAccountRootPath(writeCache, account))
}

// Create a new account in the write cache
func createAccount(writeCache *cache.WriteCache, account evmcommon.Address, tid uint64) {
	if _, err := CreateNewAccount(tid, hexutil.Encode(account[:]), writeCache); err != nil {
		panic(err)
	}

	if _, err := writeCache.Write(tid, getBalancePath(writeCache, account), commutative.NewUnboundedU256()); err != nil { // Initialize balance
		panic(err)
	}

	if _, err := writeCache.Write(tid, getNoncePath(writeCache, account), commutative.NewUnboundedUint64()); err != nil {
		panic(err)
	}
	// if err := writeCache.Write(tid, getCodePath(writeCache, account), noncommutative.NewBytes(nil)); err != nil {
	// 	panic(err)
	// }
}
