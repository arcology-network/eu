package eu

import (
	commonlibcommon "github.com/arcology-network/common-lib/common"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/univalue"
	"github.com/arcology-network/eu/cache"
)

type StateFilter struct {
	*cache.WriteCache
	ignoreAddresses map[string]bool
}

func NewStateFilter(cache *cache.WriteCache) *StateFilter {
	return &StateFilter{
		cache,
		map[string]bool{},
	}
}

func (this *StateFilter) RemoveByAddress(addr string) {
	commonlibcommon.MapRemoveIf(this.Cache(),
		func(path string, _ *univalue.Univalue) bool {
			return path[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH] == addr
		},
	)
}

func (this *StateFilter) AddToAutoReversion(addr string) {
	if _, ok := (this.ignoreAddresses)[addr]; !ok {
		(this.ignoreAddresses)[addr] = true
	}
}

func (this *StateFilter) filterByAddress(transitions *[]*univalue.Univalue) []*univalue.Univalue {
	if len(this.ignoreAddresses) == 0 {
		return *transitions
	}

	out := commonlibcommon.RemoveIf(transitions, func(v *univalue.Univalue) bool {
		address := (*v.GetPath())[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH : ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		_, ok := this.ignoreAddresses[address]
		return ok
	})

	return out
}

func (this *StateFilter) Raw() []*univalue.Univalue {
	transitions := this.Export()
	return this.filterByAddress(&transitions)
}

func (this *StateFilter) ByType() ([]*univalue.Univalue, []*univalue.Univalue) {
	accesses, transitions := this.ExportAll()
	return this.filterByAddress(&accesses),
		this.filterByAddress(&transitions)
}
