package execution

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
	"github.com/arcology-network/concurrenturl/interfaces"
	ccurlintef "github.com/arcology-network/concurrenturl/interfaces"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

type WriteCache struct {
	store    ccurlintef.DataStoreReader
	kvDict   map[string]ccurlintef.Univalue // Local KV lookup
	platform ccurlintef.Platform
	buffer   []ccurlintef.Univalue // Transition + access record buffer
	uniPool  *mempool.Mempool
}

func NewWriteCache(store ccurlintef.DataStoreReader, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]ccurlintef.Univalue)
	writeCache.platform = ccurlcommon.NewPlatform()
	writeCache.buffer = make([]ccurlintef.Univalue, 0, 64)

	writeCache.uniPool = mempool.NewMempool("writecache-univalue", func() interface{} { return new(univalue.Univalue) })
	return &writeCache
}

func (this *WriteCache) SetStore(store ccurlintef.DataStoreReader) { this.store = store }
func (this *WriteCache) Store() ccurlintef.DataStoreReader         { return this.store }
func (this *WriteCache) Cache() *map[string]ccurlintef.Univalue    { return &this.kvDict }

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string, T any) ccurlintef.Univalue {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, 0, common.FilterFirst(this.Store().Retrive(path, T)), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}) {
	univalue := this.GetOrInit(tx, path, T)
	return univalue.Get(tx, path, nil), univalue
}

// Get the value directly, skip the access counting at the univalue level
func (this *WriteCache) Peek(path string, T any) (interface{}, interface{}) {
	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ
	}

	v, _ := this.Store().Retrive(path, T)
	univ := univalue.NewUnivalue(ccurlcommon.SYSTEM, path, 0, 0, 0, v, nil)
	return univ.Value(), univ
}

func (this *WriteCache) Retrive(path string, T any) (interface{}, error) {
	typedv, _ := this.Peek(path, T)
	if typedv == nil || typedv.(ccurlintef.Type).IsDeltaApplied() {
		return typedv, nil
	}

	rawv, _, _ := typedv.(ccurlintef.Type).Get()
	return typedv.(ccurlintef.Type).New(rawv, nil, nil, typedv.(ccurlintef.Type).Min(), typedv.(ccurlintef.Type).Max()), nil // Return in a new univalue
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

func (this *WriteCache) AddTransitions(transitions []ccurlintef.Univalue) {
	if len(transitions) == 0 {
		return
	}

	newPathCreations := common.MoveIf(&transitions, func(v ccurlintef.Univalue) bool {
		return common.IsPath(*v.GetPath()) && !v.Preexist()
	})

	// Remove the changes from the existing paths, as they will be updated automatically when inserting sub elements.
	transitions = common.RemoveIf(&transitions, func(v ccurlintef.Univalue) bool {
		return common.IsPath(*v.GetPath())
	})

	// Not necessary at the moment, but good for the future if multiple level containers are available
	newPathCreations = indexer.Univalues(indexer.Sorter(newPathCreations))
	common.Foreach(newPathCreations, func(v *ccurlintef.Univalue, _ int) {
		(*v).Merge(this) // Write back to the parent writecache
	})

	common.Foreach(transitions, func(v *ccurlintef.Univalue, _ int) {
		(*v).Merge(this) // Write back to the parent writecache
	})
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]ccurlintef.Univalue)
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

func (*WriteCache) Sort(univals []interfaces.Univalue) []interfaces.Univalue {
	sort.SliceStable(univals, func(i, j int) bool {
		lhs := (*(univals[i].GetPath()))
		rhs := (*(univals[j].GetPath()))
		return bytes.Compare([]byte(lhs)[:], []byte(rhs)[:]) < 0
	})
	return univals
}

func (this *WriteCache) Export(preprocessors ...func([]ccurlintef.Univalue) []ccurlintef.Univalue) []ccurlintef.Univalue {
	this.buffer = common.MapValues(this.kvDict) //this.buffer[:0]

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []ccurlintef.Univalue {
			return processor(this.buffer)
		}, this.buffer)
	}

	common.RemoveIf(&this.buffer, func(v ccurlintef.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return this.buffer
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
