package cache

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	mempool "github.com/arcology-network/common-lib/mempool"
	committercommon "github.com/arcology-network/concurrenturl/common"
	concurrenturlcommon "github.com/arcology-network/concurrenturl/common"
	"github.com/arcology-network/concurrenturl/commutative"
	indexer "github.com/arcology-network/concurrenturl/importer"
	"github.com/arcology-network/concurrenturl/interfaces"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

// WriteCache is a read-only data store used for caching.
type WriteCache struct {
	store    intf.ReadOnlyDataStore
	kvDict   map[string]intf.Univalue // Local KV lookup
	platform intf.Platform
	buffer   []intf.Univalue // Transition + access record buffer
	uniPool  *mempool.Mempool
}

// NewWriteCache creates a new instance of WriteCache; the store can be another instance of WriteCache,
// resulting in a cascading-like structure.
func NewWriteCache(store intf.ReadOnlyDataStore, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]intf.Univalue)
	writeCache.platform = concurrenturlcommon.NewPlatform()
	writeCache.buffer = make([]intf.Univalue, 0, 64)

	writeCache.uniPool = mempool.NewMempool("writecache-univalue", func() interface{} { return new(univalue.Univalue) })
	return &writeCache
}

// CreateNewAccount creates a new account in the write cache.
// It returns the transitions and an error, if any.
func (this *WriteCache) CreateNewAccount(tx uint32, acct string) ([]intf.Univalue, error) {
	paths, typeids := committercommon.NewPlatform().GetBuiltins(acct)

	transitions := []intf.Univalue{}
	for i, path := range paths {
		var v interface{}
		switch typeids[i] {
		case commutative.PATH: // Path
			v = commutative.NewPath()

		case uint8(reflect.Kind(noncommutative.STRING)): // delta big int
			v = noncommutative.NewString("")

		case uint8(reflect.Kind(commutative.UINT256)): // delta big int
			v = commutative.NewUnboundedU256()

		case uint8(reflect.Kind(commutative.UINT64)):
			v = commutative.NewUnboundedUint64()

		case uint8(reflect.Kind(noncommutative.INT64)):
			v = new(noncommutative.Int64)

		case uint8(reflect.Kind(noncommutative.BYTES)):
			v = noncommutative.NewBytes([]byte{})
		}

		if !this.IfExists(path) {
			transitions = append(transitions, univalue.NewUnivalue(tx, path, 0, 1, 0, v, nil))

			if _, err := this.Write(tx, path, v); err != nil { // root path
				return nil, err
			}

			if !this.IfExists(path) {
				_, err := this.Write(tx, path, v)
				return transitions, err // root path
			}
		}
	}
	return transitions, nil
}

// func (this *WriteCache) SetStore(store intf.ReadOnlyDataStore) { this.store = store }
func (this *WriteCache) ReadOnlyDataStore() intf.ReadOnlyDataStore { return this.store }
func (this *WriteCache) Cache() *map[string]intf.Univalue          { return &this.kvDict }

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	v := this.uniPool.Get().(*univalue.Univalue)
	return v
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string, T any) intf.Univalue {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.(*univalue.Univalue).Init(tx, path, 0, 0, 0, common.FilterFirst(this.ReadOnlyDataStore().Retrive(path, T)), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}, uint64) {
	univalue := this.GetOrInit(tx, path, T)
	return univalue.Get(tx, path, nil), univalue, 0
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}) (int64, error) {
	// fmt.Println("Write: ", path, "|", value)
	fee := int64(0) //Fee{}.Writer(path, value, this.writeCache)
	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, common.FilterSecond(this.write(tx, path, value))
	}
	return fee, errors.New("Error: Unknown data type !")
}

// func (this *WriteCache) ReadEx(tx uint32, path string, T any) (interface{}, uint64) {
// 	univ := this.GetOrInit(tx, path, T)
// 	// return univalue.Get(tx, path, nil), univalue
// 	return univ.Get(tx, path, nil), 0 //Fee{}.Reader(univ.(intf.Univalue))
// }

// Get data from the DB direcly, still under conflict protection
func (this *WriteCache) ReadCommitted(tx uint32, key string, T any) (interface{}, uint64) {
	if v, _, Fee := this.Read(tx, key, this); v != nil { // For conflict detection
		return v, Fee
	}

	v, _ := this.ReadOnlyDataStore().Retrive(key, T)
	if v == nil {
		return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v, nil))
	}
	return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v.(interfaces.Type), nil))
}

// Read th Nth element under a path
// func (this *WriteCache) getKeyByIdx(tx uint32, path string, idx uint64) (interface{}, uint64, error) {
// 	if !common.IsPath(path) {
// 		return nil, READ_NONEXIST, errors.New("Error: Not a path!!!")
// 	}

// 	meta, readFee := this.ReadEx(tx, path, new(commutative.Path)) // read the container meta
// 	return common.IfThen(meta == nil,
// 		meta,
// 		common.IfThenDo1st(idx < uint64(len(meta.(*orderedset.OrderedSet).Keys())), func() interface{} { return path + meta.(*orderedset.OrderedSet).Keys()[idx] }, nil),
// 	), readFee, nil
// }

// func (this *WriteCache) ReadAt(tx uint32, path string, idx uint64, T any) (interface{}, uint64, error) {
// 	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
// 		v, Fee := this.ReadEx(tx, key.(string), T)
// 		return v, Fee, nil
// 	} else {
// 		return key, Fee, err
// 	}
// }

// func (this *WriteCache) WriteAt(tx uint32, path string, idx uint64, T any) (int64, error) {
// 	if !common.IsPath(path) {
// 		return int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
// 	}

// 	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil {
// 		return this.Write(tx, key.(string), T)
// 	} else {
// 		return int64(Fee), err
// 	}
// }

// Get the raw value directly, skip the access counting at the univalue level
func (this *WriteCache) Find(path string, T any) (interface{}, interface{}) {
	if univ, ok := this.kvDict[path]; ok {
		return univ.Value(), univ
	}

	v, _ := this.ReadOnlyDataStore().Retrive(path, T)
	univ := univalue.NewUnivalue(committercommon.SYSTEM, path, 0, 0, 0, v, nil)
	return univ.Value(), univ
}

func (this *WriteCache) Retrive(path string, T any) (interface{}, error) {
	typedv, _ := this.Find(path, T)
	if typedv == nil || typedv.(intf.Type).IsDeltaApplied() {
		return typedv, nil
	}

	rawv, _, _ := typedv.(intf.Type).Get()
	return typedv.(intf.Type).New(rawv, nil, nil, typedv.(intf.Type).Min(), typedv.(intf.Type).Max()), nil // Return in a new univalue
}

// func (this *WriteCache) Do(tx uint32, path string, doer interface{}, T any) (interface{}, error) {
// 	univalue := this.GetOrInit(tx, path, T)
// return univalue.Do(tx, path, doer), nil
// }

func (this *WriteCache) write(tx uint32, path string, value interface{}) (int64, error) {
	parentPath := common.GetParentPath(path)
	if this.IfExists(parentPath) || tx == committercommon.SYSTEM { // The parent path exists or to inject the path directly
		univalue := this.GetOrInit(tx, path, value) // Get a univalue wrapper

		err := univalue.Set(tx, path, value, this)
		if err == nil {
			if strings.HasSuffix(parentPath, "/container/") || (!this.platform.IsSysPath(parentPath) && tx != committercommon.SYSTEM) { // Don't keep track of the system children
				parentMeta := this.GetOrInit(tx, parentPath, new(commutative.Path))
				err = parentMeta.Set(tx, path, univalue.Value(), this)
			}
		}
		return 0, err
	}
	return 0, errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) IfExists(path string) bool {
	if committercommon.ETH10_ACCOUNT_PREFIX_LENGTH == len(path) {
		return true
	}

	if v := this.kvDict[path]; v != nil {
		return v.Value() != nil // If value == nil means either it's been deleted or never existed.
	}
	return this.store.IfExists(path) //this.RetriveShallow(path, nil) != nil
}

func (this *WriteCache) AddTransitions(transitions []intf.Univalue) {
	if len(transitions) == 0 {
		return
	}

	newPathCreations := common.MoveIf(&transitions, func(v intf.Univalue) bool {
		return common.IsPath(*v.GetPath()) && !v.Preexist()
	})

	// Remove the changes from the existing paths, as they will be updated automatically when inserting sub elements.
	transitions = common.RemoveIf(&transitions, func(v intf.Univalue) bool {
		return common.IsPath(*v.GetPath())
	})

	// Not necessary at the moment, but good for the future if multiple level containers are available
	newPathCreations = indexer.Univalues(indexer.Sorter(newPathCreations))
	common.Foreach(newPathCreations, func(v *intf.Univalue, _ int) {
		(*v).CopyTo(this) // Write back to the parent writecache
	})

	common.Foreach(transitions, func(v *intf.Univalue, _ int) {
		(*v).CopyTo(this) // Write back to the parent writecache
	})
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]intf.Univalue)
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

func (this *WriteCache) Export(preprocessors ...func([]intf.Univalue) []intf.Univalue) []intf.Univalue {
	this.buffer = common.MapValues(this.kvDict) //this.buffer[:0]

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []intf.Univalue {
			return processor(this.buffer)
		}, this.buffer)
	}

	common.RemoveIf(&this.buffer, func(v intf.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return this.buffer
}

func (this *WriteCache) ExportAll(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) ([]interfaces.Univalue, []interfaces.Univalue) {
	all := this.Export(indexer.Sorter)
	// indexer.Univalues(all).Print()

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
