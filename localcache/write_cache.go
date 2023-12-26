package writecache

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mempool "github.com/arcology-network/common-lib/mempool"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/indexer"
	ccurlintf "github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type WriteCache struct {
	store    ccurlintf.ReadOnlyDataStore
	kvDict   map[string]ccurlintf.Univalue // Local KV lookup
	platform ccurlintf.Platform
	buffer   []ccurlintf.Univalue // Transition + access record buffer
	uniPool  *mempool.Mempool
}

func NewWriteCache(store ccurlintf.ReadOnlyDataStore, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]ccurlintf.Univalue)
	writeCache.platform = ccurlcommon.NewPlatform()
	writeCache.buffer = make([]ccurlintf.Univalue, 0, 64)

	writeCache.uniPool = mempool.NewMempool("writecache-univalue", func() interface{} { return new(univalue.Univalue) })
	return &writeCache
}

func (this *WriteCache) SetStore(store ccurlintf.ReadOnlyDataStore) { this.store = store }
func (this *WriteCache) Store() ccurlintf.ReadOnlyDataStore         { return this.store }
func (this *WriteCache) Cache() *map[string]ccurlintf.Univalue      { return &this.kvDict }

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string, T any) ccurlintf.Univalue {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, 0, common.FilterFirst(this.store.Retrive(path, T)), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}) {
	univalue := this.GetOrInit(tx, path, T)
	return univalue.Get(tx, path, nil), univalue
}

func (this *WriteCache) ReadCommitted(tx uint32, key string, T any) (interface{}, uint64) {
	if v, _ := this.Read(tx, key, this); v != nil { // For conflict detection
		return v, 0
	}

	v, _ := this.store.Retrive(key, T)
	if v == nil {
		return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v, nil))
	}
	return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v.(interfaces.Type), nil))
}

// Get the value directly, skip the access counting at the univalue level
func (this *WriteCache) Peek(path string, T any) (interface{}, interface{}) {
	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ
	}

	v, _ := this.store.Retrive(path, T)
	univ := univalue.NewUnivalue(ccurlcommon.SYSTEM, path, 0, 0, 0, v, nil)
	return univ.Value(), univ
}

func (this *WriteCache) Retrive(path string, T any) (interface{}, error) {
	typedv, _ := this.Peek(path, T)
	if typedv == nil || typedv.(ccurlintf.Type).IsDeltaApplied() {
		return typedv, nil
	}

	rawv, _, _ := typedv.(ccurlintf.Type).Get()
	return typedv.(ccurlintf.Type).New(rawv, nil, nil, typedv.(ccurlintf.Type).Min(), typedv.(ccurlintf.Type).Max()), nil // Return in a new univalue
}

func (this *WriteCache) Do(tx uint32, path string, doer interface{}, T any) interface{} {
	univalue := this.GetOrInit(tx, path, T)
	return univalue.Do(tx, path, doer)
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}) error {
	parentPath := common.GetParentPath(path)
	if this.IfExists(parentPath) || tx == ccurlcommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.GetOrInit(tx, path, value) // Get a univalue wrapper

		err := univalue.Set(tx, path, value, this)
		if err == nil {
			if strings.HasSuffix(parentPath, "/container/") || (!this.platform.IsSysPath(parentPath) && tx != ccurlcommon.SYSTEM) { // Don't keep track of the system children
				parentMeta := this.GetOrInit(tx, parentPath, new(commutative.Path))
				err = parentMeta.Set(tx, path, univalue.Value(), this)
			}
		}
		return err
	}
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) IfExists(path string) bool {
	if ccurlcommon.ETH10_ACCOUNT_PREFIX_LENGTH == len(path) {
		return true
	}

	if v := this.kvDict[path]; v != nil {
		return v.Value() != nil // If value == nil means either it's been deleted or never existed.
	}
	return this.store.IfExists(path) //this.RetriveShallow(path, nil) != nil
}

func (this *WriteCache) AddTransitions(transitions []ccurlintf.Univalue) {
	if len(transitions) == 0 {
		return
	}

	newPathCreations := common.MoveIf(&transitions, func(v ccurlintf.Univalue) bool {
		return common.IsPath(*v.GetPath()) && !v.Preexist()
	})

	// Remove the changes from the existing paths, as they will be updated automatically when inserting sub elements.
	transitions = common.RemoveIf(&transitions, func(v ccurlintf.Univalue) bool {
		return common.IsPath(*v.GetPath())
	})

	// Not necessary at the moment, but good for the future if multiple level containers are available
	newPathCreations = indexer.Univalues(this.Sort(newPathCreations))
	common.Foreach(newPathCreations, func(v *ccurlintf.Univalue, _ int) {
		(*v).Merge(this) // Write back to the parent writecache
	})

	common.Foreach(transitions, func(v *ccurlintf.Univalue, _ int) {
		(*v).Merge(this) // Write back to the parent writecache
	})
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]ccurlintf.Univalue)
}

func (this *WriteCache) Equal(other *WriteCache) bool {
	thisBuffer := common.MapValues(this.kvDict)
	sort.SliceStable(thisBuffer, func(i, j int) bool {
		return *thisBuffer[i].GetPath() < *thisBuffer[j].GetPath()
	})

	otherBuffer := common.MapValues(other.kvDict)
	sort.SliceStable(otherBuffer, func(i, j int) bool {
		return *otherBuffer[i].GetPath() < *otherBuffer[j].GetPath()
	})

	cacheFlag := reflect.DeepEqual(thisBuffer, otherBuffer)
	return cacheFlag
}

func (*WriteCache) Sort(univals []ccurlintf.Univalue) []ccurlintf.Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}

func (this *WriteCache) Export(preprocessors ...func([]ccurlintf.Univalue) []ccurlintf.Univalue) []ccurlintf.Univalue {
	this.buffer = common.MapValues(this.kvDict) //this.buffer[:0]

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []ccurlintf.Univalue {
			return processor(this.buffer)
		}, this.buffer)
	}

	common.RemoveIf(&this.buffer, func(v ccurlintf.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return this.buffer
}

func (this *WriteCache) ExportAll(preprocessors ...func([]ccurlintf.Univalue) []ccurlintf.Univalue) ([]ccurlintf.Univalue, []ccurlintf.Univalue) {
	all := this.Export(this.Sort)

	accesses := indexer.Univalues(common.Clone(all)).To(indexer.ITCAccess{})
	transitions := indexer.Univalues(common.Clone(all)).To(indexer.ITCTransition{})
	return accesses, transitions
}

func (this *WriteCache) Print() {
	values := common.MapValues(this.kvDict)
	sort.SliceStable(values, func(i, j int) bool {
		return *values[i].GetPath() < *values[j].GetPath()
	})

	for i, elem := range values {
		fmt.Println("Level : ", i)
		elem.Print()
	}
}
