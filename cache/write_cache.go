package cache

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	common "github.com/arcology-network/common-lib/common"
	"github.com/arcology-network/common-lib/exp/array"
	mempool "github.com/arcology-network/common-lib/exp/mempool"
	ccurl "github.com/arcology-network/concurrenturl"
	committercommon "github.com/arcology-network/concurrenturl/common"
	platform "github.com/arcology-network/concurrenturl/platform"

	"github.com/arcology-network/concurrenturl/commutative"
	importer "github.com/arcology-network/concurrenturl/importer"
	"github.com/arcology-network/concurrenturl/interfaces"
	intf "github.com/arcology-network/concurrenturl/interfaces"
	"github.com/arcology-network/concurrenturl/noncommutative"
	univalue "github.com/arcology-network/concurrenturl/univalue"
)

// WriteCache is a read-only data store used for caching.
type WriteCache struct {
	store    intf.ReadOnlyDataStore
	kvDict   map[string]*univalue.Univalue // Local KV lookup
	platform intf.Platform
	buffer   []*univalue.Univalue // Transition + access record buffer
	uniPool  *mempool.Mempool[*univalue.Univalue]
}

// NewWriteCache creates a new instance of WriteCache; the store can be another instance of WriteCache,
// resulting in a cascading-like structure.
func NewWriteCache(store intf.ReadOnlyDataStore, args ...interface{}) *WriteCache {
	var writeCache WriteCache
	writeCache.store = store
	writeCache.kvDict = make(map[string]*univalue.Univalue)
	writeCache.platform = platform.NewPlatform()
	writeCache.buffer = make([]*univalue.Univalue, 0, 64)

	writeCache.uniPool = mempool.NewMempool[*univalue.Univalue](4096, 64, func() *univalue.Univalue {
		return new(univalue.Univalue)
	})
	return &writeCache
}

// CreateNewAccount creates a new account in the write cache.
// It returns the transitions and an error, if any.
func (this *WriteCache) CreateNewAccount(tx uint32, acct string) ([]*univalue.Univalue, error) {
	paths, typeids := platform.NewPlatform().GetBuiltins(acct)

	transitions := []*univalue.Univalue{}
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

		// fmt.Println(path)
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

func (this *WriteCache) ReadOnlyDataStore() intf.ReadOnlyDataStore { return this.store }
func (this *WriteCache) Cache() map[string]*univalue.Univalue      { return this.kvDict }

func (this *WriteCache) NewUnivalue() *univalue.Univalue {
	return this.uniPool.New()
}

// If the access has been recorded
func (this *WriteCache) GetOrInit(tx uint32, path string, T any) *univalue.Univalue {
	unival := this.kvDict[path]
	if unival == nil { // Not in the kvDict, check the datastore
		unival = this.NewUnivalue()
		unival.Init(tx, path, 0, 0, 0, common.FilterFirst(this.ReadOnlyDataStore().Retrive(path, T)), this)
		this.kvDict[path] = unival // Adding to kvDict
	}
	return unival
}

func (this *WriteCache) Read(tx uint32, path string, T any) (interface{}, interface{}, uint64) {
	univalue := this.GetOrInit(tx, path, T)
	return univalue.Get(tx, path, nil), univalue, 0
}

func (this *WriteCache) write(tx uint32, path string, value interface{}) error {
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
		return err
	}
	return errors.New("Error: The parent path doesn't exist: " + parentPath)
}

func (this *WriteCache) Write(tx uint32, path string, value interface{}) (int64, error) {
	fee := int64(0) //Fee{}.Writer(path, value, this.writeCache)
	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, this.write(tx, path, value)
	}
	return fee, errors.New("Error: Unknown data type !")
}

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

// Get the raw value directly, skip the access counting at the univalue level
func (this *WriteCache) InCache(path string) (interface{}, bool) {
	univ, ok := this.kvDict[path]
	return univ, ok
}

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

func (this *WriteCache) IfExists(path string) bool {
	if committercommon.ETH10_ACCOUNT_PREFIX_LENGTH == len(path) {
		return true
	}

	if v := this.kvDict[path]; v != nil {
		return v.Value() != nil // If value == nil means either it's been deleted or never existed.
	}
	return this.store.IfExists(path) //this.RetriveShallow(path, nil) != nil
}

func (this *WriteCache) AddTransitions(transitions []*univalue.Univalue) {
	if len(transitions) == 0 {
		return
	}

	newPathCreations := array.MoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		return common.IsPath(*v.GetPath()) && !v.Preexist()
	})

	// Remove the changes from the existing paths, as they will be updated automatically when inserting sub elements.
	transitions = array.RemoveIf(&transitions, func(_ int, v *univalue.Univalue) bool {
		return common.IsPath(*v.GetPath())
	})

	// Not necessary at the moment, but good for the future if multiple level containers are available
	newPathCreations = univalue.Univalues(importer.Sorter(newPathCreations))
	array.Foreach(newPathCreations, func(_ int, v **univalue.Univalue) {
		(*v).CopyTo(this) // Write back to the parent writecache
	})

	array.Foreach(transitions, func(_ int, v **univalue.Univalue) {
		(*v).CopyTo(this) // Write back to the parent writecache
	})
}

func (this *WriteCache) Clear() {
	this.kvDict = make(map[string]*univalue.Univalue)
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

func (this *WriteCache) Export(preprocessors ...func([]*univalue.Univalue) []*univalue.Univalue) []*univalue.Univalue {
	this.buffer = common.MapValues(this.kvDict) //this.buffer[:0]

	for _, processor := range preprocessors {
		this.buffer = common.IfThenDo1st(processor != nil, func() []*univalue.Univalue {
			return processor(this.buffer)
		}, this.buffer)
	}

	array.RemoveIf(&this.buffer, func(_ int, v *univalue.Univalue) bool { return v.Reads() == 0 && v.IsReadOnly() }) // Remove peeks
	return this.buffer
}

func (this *WriteCache) ExportAll(preprocessors ...func([]*univalue.Univalue) []*univalue.Univalue) ([]*univalue.Univalue, []*univalue.Univalue) {
	all := this.Export(importer.Sorter)
	// univalue.Univalues(all).Print()

	accesses := univalue.Univalues(array.Clone(all)).To(importer.ITAccess{})
	transitions := univalue.Univalues(array.Clone(all)).To(importer.ITTransition{})
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

func (this *WriteCache) KVs() ([]string, []intf.Type) {
	transitions := univalue.Univalues(array.Clone(this.Export(importer.Sorter))).To(importer.ITTransition{})

	values := make([]intf.Type, len(transitions))
	keys := array.ParallelAppend(transitions, 4, func(i int, v *univalue.Univalue) string {
		values[i] = v.Value().(intf.Type)
		return *v.GetPath()
	})
	return keys, values
}

// This function is used to write the cache to the data source directly to bypass all the intermediate steps,
// including the conflict detection.
//
// It's mainly used for TESTING purpose.
func (this *WriteCache) FlushToDataSource(store interfaces.Datastore) interfaces.Datastore {
	committer := ccurl.NewStorageCommitter(store)
	acctTrans := univalue.Univalues(array.Clone(this.Export(importer.Sorter))).To(importer.IPTransition{})

	txs := array.Append(acctTrans, func(_ int, v *univalue.Univalue) uint32 {
		return v.GetTx()
	})

	committer.Import(acctTrans)
	committer.Sort()
	committer.Precommit(txs)
	committer.Commit()
	this.Clear()

	return store
}
