package cache

import (
	common "github.com/arcology-network/common-lib/common"
	commonlibcommon "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/univalue"
)

// WriteCacheFilter is a wrapper of WriteCache and it filters
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
	return common.MapValues(this.WriteCache.Cache())
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

	out := array.RemoveIf(transitions, func(_ int, v *univalue.Univalue) bool {
		address := (*v.GetPath())[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH : ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		_, ok := this.ignoreAddresses[address]
		return ok
	})

	return out
}

func (this *WriteCacheFilter) ByType() ([]*univalue.Univalue, []*univalue.Univalue) {
	accesses, transitions := this.ExportAll()
	return this.filterByAddress(&accesses),
		this.filterByAddress(&transitions)
}
