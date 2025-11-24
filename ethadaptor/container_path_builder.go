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
	commutative "github.com/arcology-network/common-lib/crdt/commutative"
	"github.com/arcology-network/common-lib/types"
	stgcommon "github.com/arcology-network/state-engine/common"
	cache "github.com/arcology-network/state-engine/state/cache"
	evmcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"

	// eucommon ""github.com/arcology-network/state-engine/state/cache""
	intf "github.com/arcology-network/eu/interface"
)

// This class helps to build paths to access container related data in Arcology storage structure.
type ContainerPathBuilder struct {
	apiRouter intf.EthApiRouter
	subDir    string
}

func NewPathBuilder(subDir string, api intf.EthApiRouter) *ContainerPathBuilder {
	return &ContainerPathBuilder{
		subDir:    subDir,
		apiRouter: api,
	}
}

// Make Arcology paths under the current account
func (this *ContainerPathBuilder) CreateNewAccount(txIndex uint64, account types.Address) (bool, string) {
	accountRoot := common.StrCat(stgcommon.ETH_ACCOUNT_PREFIX, string(account), "/")
	if !this.apiRouter.StateCache().(*cache.StateCache).IfExists(accountRoot) {
		return true, this.key(account) // ALready exists
	}

	_, err := CreateDefaultPaths(txIndex, string(account), this.apiRouter.StateCache().(*cache.StateCache))
	return err == nil, this.key(account)
}

func (this *ContainerPathBuilder) Key(caller [20]byte) string { // container ID
	return this.key(types.Address(hexutil.Encode(caller[:])))
}

func (this *ContainerPathBuilder) key(account types.Address) string { // container ID
	return common.StrCat(stgcommon.ETH_ACCOUNT_PREFIX, string(account), this.subDir, "/")
}

func (this *ContainerPathBuilder) Root() string { // container ID
	return common.StrCat(stgcommon.ETH_ACCOUNT_PREFIX, this.subDir, "/")
}

func (this *ContainerPathBuilder) GetPathType(caller evmcommon.Address) uint8 {
	pathStr := this.Key(caller) // Container path
	if len(pathStr) == 0 {
		return 0
	}
	return this.PathElemTypeIDs(pathStr) // Get the path type
}

func (this *ContainerPathBuilder) PathElemTypeIDs(pathStr string) uint8 {
	_, path, _ := this.apiRouter.StateCache().(*cache.StateCache).Peek(pathStr, commutative.Path{})
	return path.(*commutative.Path).ElemType
}
