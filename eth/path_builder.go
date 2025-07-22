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
	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/types"
	ccurlcommon "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	// eucommon ""github.com/arcology-network/storage-committer/storage/cache""
	intf "github.com/arcology-network/eu/interface"
)

// Ccurl connectors for Arcology APIs
type PathBuilder struct {
	apiRouter intf.EthApiRouter
	// apiRouter.WriteCache()     *concurrenturl.ConcurrentUrl
	subDir string
}

func NewPathBuilder(subDir string, api intf.EthApiRouter) *PathBuilder {
	return &PathBuilder{
		subDir:    subDir,
		apiRouter: api,
		// apiRouter.WriteCache():     apiRouter.WriteCache(),
	}
}

// Make Arcology paths under the current account
func (this *PathBuilder) CreateNewAccount(txIndex uint64, account types.Address, typeid uint8, isTransient bool) (bool, string) {
	accountRoot := common.StrCat(ccurlcommon.ETH10_ACCOUNT_PREFIX, string(account), "/")
	if !this.apiRouter.WriteCache().(*cache.WriteCache).IfExists(accountRoot) {
		_, err := CreateNewAccount(txIndex, string(account), this.apiRouter.WriteCache().(*cache.WriteCache))
		return err == nil, this.key(account)
	}
	return true, this.key(account) // ALready exists
}

func (this *PathBuilder) Key(caller [20]byte) string { // container ID
	return this.key(types.Address(hexutil.Encode(caller[:])))
}

func (this *PathBuilder) key(account types.Address) string { // container ID
	return common.StrCat(ccurlcommon.ETH10_ACCOUNT_PREFIX, string(account), this.subDir, "/")
}

func (this *PathBuilder) Root() string { // container ID
	return common.StrCat(ccurlcommon.ETH10_ACCOUNT_PREFIX, this.subDir, "/")
}

func (this *PathBuilder) GetPathType(caller evmcommon.Address) uint8 {
	pathStr := this.Key(caller) // Container path
	if len(pathStr) == 0 {
		return 0
	}
	return this.PathElemTypeIDs(pathStr) // Get the path type
}

func (this *PathBuilder) PathElemTypeIDs(pathStr string) uint8 {
	_, path, _ := this.apiRouter.WriteCache().(*cache.WriteCache).Peek(pathStr, commutative.Path{})
	return path.(*commutative.Path).ElemType
}
