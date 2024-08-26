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
	ccurlcommon "github.com/arcology-network/common-lib/types/storage/common"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	cache "github.com/arcology-network/common-lib/types/storage/writecache"
	"github.com/ethereum/go-ethereum/common/hexutil"

	// adaptorcommon "github.com/arcology-network/storage-committer/storage/writecache"
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
func (this *PathBuilder) New(txIndex uint32, deploymentAddr types.Address) bool {
	if !this.newStorageRoot(deploymentAddr, txIndex) { // Create the root path if has been created yet.
		return false
	}
	return this.newContainerRoot(deploymentAddr, txIndex) //
}

func (this *PathBuilder) newStorageRoot(account types.Address, txIndex uint32) bool {
	accountRoot := common.StrCat(ccurlcommon.ETH10_ACCOUNT_PREFIX, string(account), "/")
	if !this.apiRouter.WriteCache().(*cache.WriteCache).IfExists(accountRoot) {
		_, err := CreateNewAccount(txIndex, string(account), this.apiRouter.WriteCache().(*cache.WriteCache))
		return err == nil
		// return common.FilterFirst(this.apiRouter.WriteCache().(*cache.WriteCache).CreateNewAccount() != nil // Create a new account
	}
	return true // ALready exists
}

func (this *PathBuilder) newContainerRoot(account types.Address, txIndex uint32) bool {
	containerRoot := this.key(account)

	if !this.apiRouter.WriteCache().(*cache.WriteCache).IfExists(containerRoot) {
		_, err := this.apiRouter.WriteCache().(*cache.WriteCache).Write(txIndex, containerRoot, commutative.NewPath()) // Create a new container
		return err == nil
	}
	return true // Already exists
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
