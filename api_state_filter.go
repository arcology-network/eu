package eu

import (
	commonlibcommon "github.com/arcology-network/common-lib/common"

	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/interfaces"
	intf "github.com/arcology-network/vm-adaptor/interface"
)

type StateFilter struct {
	api             intf.EthApiRouter
	ignoreAddresses map[string]bool
}

func NewExportFilter(api intf.EthApiRouter) *StateFilter {
	return &StateFilter{
		api,
		map[string]bool{},
	}
}

func (this *StateFilter) RemoveByAddress(addr string) {
	cache := this.api.Ccurl().WriteCache().Cache()
	commonlibcommon.MapRemoveIf(*cache,
		func(path string, _ interfaces.Univalue) bool {
			return path[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH:ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH] == addr
		},
	)
}

func (this *StateFilter) AddToAutoReversion(addr string) {
	if _, ok := (this.ignoreAddresses)[addr]; !ok {
		(this.ignoreAddresses)[addr] = true
	}
}

func (this *StateFilter) filterByAddress(transitions *[]interfaces.Univalue) []interfaces.Univalue {
	if len(this.ignoreAddresses) == 0 {
		return *transitions
	}

	out := commonlibcommon.RemoveIf(transitions, func(v interfaces.Univalue) bool {
		address := (*v.GetPath())[ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH : ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH+ccurlcommon.ETH10_ACCOUNT_LENGTH]
		_, ok := this.ignoreAddresses[address]
		return ok
	})

	return out
}

func (this *StateFilter) Raw() []interfaces.Univalue {
	transitions := this.api.Ccurl().Export()
	return this.filterByAddress(&transitions)
}

func (this *StateFilter) ByType() ([]interfaces.Univalue, []interfaces.Univalue) {
	accesses, transitions := this.api.Ccurl().ExportAll()
	return this.filterByAddress(&accesses),
		this.filterByAddress(&transitions)
}
