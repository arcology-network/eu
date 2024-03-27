/*
*   Copyright (c) 2023 Arcology Network
*   All rights reserved.

*   Licensed under the Apache License, Version 2.0 (the "License");
*   you may not use this file except in compliance with the License.
*   You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*   Unless required by applicable law or agreed to in writing, software
*   distributed under the License is distributed on an "AS IS" BASIS,
*   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*   See the License for the specific language governing permissions and
*   limitations under the License.
 */
package cache

import (
	commonlibcommon "github.com/arcology-network/common-lib/common"
	mapi "github.com/arcology-network/common-lib/exp/map"
	slice "github.com/arcology-network/common-lib/exp/slice"
	ccurlcommon "github.com/arcology-network/storage-committer/common"
	"github.com/arcology-network/storage-committer/univalue"
)

// WriteCacheFilter is a post processing filter for WriteCache.
// It is used to filter out the transitions based on the addresses.
// out the transitions based on the addresses.
type WriteCacheFilter struct {
	*WriteCache
	ignoreAddresses map[string]bool
}

func NewWriteCacheFilter(writeCache interface{}) *WriteCacheFilter {
	return &WriteCacheFilter{
		writeCache.(*WriteCache),
		map[string]bool{},
	}
}

func (this *WriteCacheFilter) ToBuffer() []*univalue.Univalue {
	return mapi.Values(*this.WriteCache.Cache())
}

func (this *WriteCacheFilter) RemoveByAddress(addr string) {
	commonlibcommon.MapRemoveIf(this.kvDict,
		func(path string, _ *univalue.Univalue) bool {
			return path[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH] == addr
		},
	)
}

func (this *WriteCacheFilter) AddToAutoReversion(addr string) {
	if _, ok := (this.ignoreAddresses)[addr]; !ok {
		(this.ignoreAddresses)[addr] = true
	}
}

func (this *WriteCacheFilter) filterByAddress(transitions *[]*univalue.Univalue) []*univalue.Univalue {
	if len(this.ignoreAddresses) == 0 {
		return *transitions
	}

	out := slice.RemoveIf(transitions, func(_ int, v *univalue.Univalue) bool {
		address := (*v.GetPath())[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH : ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		_, ok := this.ignoreAddresses[address]
		return ok
	})

	return out
}

func (this *WriteCacheFilter) ByType() ([]*univalue.Univalue, []*univalue.Univalue) {
	accesses, transitions := this.ExportAll()
	return this.filterByAddress(&accesses), this.filterByAddress(&transitions)
}
