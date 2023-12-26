package writecache

import (
	"errors"
	"math"
	"reflect"

	"github.com/arcology-network/common-lib/common"
	orderedset "github.com/arcology-network/common-lib/container/set"
	ccurlcommon "github.com/arcology-network/concurrenturl/common"

	commutative "github.com/arcology-network/concurrenturl/commutative"
	interfaces "github.com/arcology-network/concurrenturl/interfaces"
	noncommutative "github.com/arcology-network/concurrenturl/noncommutative"
	"github.com/arcology-network/concurrenturl/univalue"
)

const (
	READ_NONEXIST          = uint64(3)
	READ_COMMITTED_FROM_DB = uint64(1000) // Read for the state db
)

type ConcurrentUrl struct {
	writeCache *WriteCache
	// importer    *Importer
	// imuImporter *Importer // transitions that will take effect anyway regardless of execution failures or conflicts
	Platform *ccurlcommon.Platform
}

func NewConcurrentUrl(store interfaces.Datastore) *ConcurrentUrl {
	platform := ccurlcommon.NewPlatform()
	return &ConcurrentUrl{
		writeCache: NewWriteCache(store, platform),
		// importer:    NewImporter(store, platform),
		// imuImporter: NewImporter(store, platform),
		Platform: platform, //[]ccurlcommon.FilteredTransitionsInterface{&NonceFilter{}, &BalanceFilter{}},
	}
}

func (this *ConcurrentUrl) New(args ...interface{}) *ConcurrentUrl {
	return &ConcurrentUrl{
		writeCache: args[0].(*WriteCache),
		Platform:   ccurlcommon.NewPlatform(),
	}
}

func (this *ConcurrentUrl) WriteCache() *WriteCache { return this.writeCache }

// func (this *ConcurrentUrl) Importer() *Importer     { return this.importer }

// Get data from the DB direcly, still under conflict protection
func (this *ConcurrentUrl) ReadCommitted(tx uint32, key string, T any) (interface{}, uint64) {
	if v, Fee := this.Read(tx, key, this); v != nil { // For conflict detection
		return v, Fee
	}

	v, _ := this.WriteCache().Store().Retrive(key, T)
	if v == nil {
		return v, 0 // Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v, nil))
	}
	return v, 0 //Fee{}.Reader(univalue.NewUnivalue(tx, key, 1, 0, 0, v.(interfaces.Type), nil))
}

func (this *ConcurrentUrl) Init(store interfaces.Datastore) {
	// this.importer.Init(store)
	// this.imuImporter.Init(store)
}

func (this *ConcurrentUrl) Clear() {
	// this.importer.Store().Clear()

	this.writeCache.Clear()
	// this.importer.Clear()
	// this.imuImporter.Clear()
}

// load accounts
func (this *ConcurrentUrl) NewAccount(tx uint32, acct string) ([]interfaces.Univalue, error) {
	paths, typeids := this.Platform.GetBuiltins(acct)

	transitions := []interfaces.Univalue{}
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

		if !this.writeCache.IfExists(path) {
			transitions = append(transitions, univalue.NewUnivalue(tx, path, 0, 1, 0, v, nil))

			if err := this.writeCache.Write(tx, path, v); err != nil { // root path
				return nil, err
			}

			if !this.writeCache.IfExists(path) {
				return transitions, this.writeCache.Write(tx, path, v) // root path
			}
		}
	}
	return transitions, nil
}

func (this *ConcurrentUrl) IfExists(path string) bool {
	return this.writeCache.IfExists(path)
}

func (this *ConcurrentUrl) IndexOf(tx uint32, path string, key interface{}, T any) (uint64, uint64) {
	if !common.IsPath(path) {
		return math.MaxUint64, READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[string](key) {
			return pathInfo.(*commutative.Path).View().IdxOf(key.(string)), 0
		}
	}
	return math.MaxUint64, READ_NONEXIST
}

func (this *ConcurrentUrl) KeyAt(tx uint32, path string, index interface{}, T any) (string, uint64) {
	if !common.IsPath(path) {
		return "", READ_NONEXIST //, errors.New("Error: Not a path!!!")
	}

	getter := func(v interface{}) (uint32, uint32, uint32, interface{}) { return 1, 0, 0, v }
	if v, err := this.Do(tx, path, getter, T); err == nil {
		pathInfo := v.(interfaces.Univalue).Value()
		if common.IsType[*commutative.Path](pathInfo) && common.IsType[uint64](index) {
			return pathInfo.(*commutative.Path).View().KeyAt(index.(uint64)), 0
		}
	}
	return "", READ_NONEXIST
}

func (this *ConcurrentUrl) Peek(path string, T any) (interface{}, uint64) {
	typedv, _ := this.writeCache.Peek(path, T)
	var v interface{}
	if typedv != nil {
		v, _, _ = typedv.(interfaces.Type).Get()
	}
	return v, 0 //Fee{}.Reader(univ.(interfaces.Univalue))
}

func (this *ConcurrentUrl) PeekCommitted(path string, T any) (interface{}, uint64) {
	v, _ := this.writeCache.Store().Retrive(path, T)
	return v, READ_COMMITTED_FROM_DB
}

func (this *ConcurrentUrl) Read(tx uint32, path string, T any) (interface{}, uint64) {
	typedv, _ := this.writeCache.Read(tx, path, T)
	// fmt.Println("Read: ", path, "|", typedv)
	return typedv, 0 //Fee{}.Reader(univ.(interfaces.Univalue))
}

func (this *ConcurrentUrl) Write(tx uint32, path string, value interface{}) (int64, error) {
	// fmt.Println("Write: ", path, "|", value)
	fee := int64(0) //Fee{}.Writer(path, value, this.writeCache)
	if value == nil || (value != nil && value.(interfaces.Type).TypeID() != uint8(reflect.Invalid)) {
		return fee, this.writeCache.Write(tx, path, value)
	}

	return fee, errors.New("Error: Unknown data type !")
}

func (this *ConcurrentUrl) Do(tx uint32, path string, doer interface{}, T any) (interface{}, error) {
	return this.writeCache.Do(tx, path, doer, T), nil
}

// Read th Nth element under a path
func (this *ConcurrentUrl) getKeyByIdx(tx uint32, path string, idx uint64) (interface{}, uint64, error) {
	if !common.IsPath(path) {
		return nil, READ_NONEXIST, errors.New("Error: Not a path!!!")
	}

	meta, readFee := this.Read(tx, path, new(commutative.Path)) // read the container meta
	return common.IfThen(meta == nil,
		meta,
		common.IfThenDo1st(idx < uint64(len(meta.(*orderedset.OrderedSet).Keys())), func() interface{} { return path + meta.(*orderedset.OrderedSet).Keys()[idx] }, nil),
	), readFee, nil
}

// Read th Nth element under a path
func (this *ConcurrentUrl) ReadAt(tx uint32, path string, idx uint64, T any) (interface{}, uint64, error) {
	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
		v, Fee := this.Read(tx, key.(string), T)
		return v, Fee, nil
	} else {
		return key, Fee, err
	}
}

// Read th Nth element under a path
func (this *ConcurrentUrl) DoAt(tx uint32, path string, idx uint64, do interface{}, T any) (interface{}, uint64, error) {
	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil && key != nil {
		v, err := this.Do(tx, key.(string), do, T)
		return v, Fee, err
	} else {
		return key, Fee, err
	}
}

// Read th Nth element under a path
func (this *ConcurrentUrl) PopBack(tx uint32, path string, T any) (interface{}, int64, error) {
	if !common.IsPath(path) {
		return nil, int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
	}
	pathDecoder := T

	meta, Fee := this.Read(tx, path, pathDecoder) // read the container meta

	subkeys := meta.(*orderedset.OrderedSet).Keys()
	if subkeys == nil || len(subkeys) == 0 {
		return nil, int64(Fee), errors.New("Error: The path is either empty or doesn't exist")
	}

	key := path + subkeys[len(subkeys)-1]

	value, Fee := this.Read(tx, key, pathDecoder)
	if value == nil {
		return nil, int64(Fee), errors.New("Error: Empty container!")
	}

	writeFee, err := this.Write(tx, key, nil)
	return value, writeFee, err
}

// Read th Nth element under a path
func (this *ConcurrentUrl) WriteAt(tx uint32, path string, idx uint64, T any) (int64, error) {
	if !common.IsPath(path) {
		return int64(READ_NONEXIST), errors.New("Error: Not a path!!!")
	}

	if key, Fee, err := this.getKeyByIdx(tx, path, idx); err == nil {
		return this.Write(tx, key.(string), T)
	} else {
		return int64(Fee), err
	}
}

func (this *ConcurrentUrl) Import(transitions []interfaces.Univalue, args ...interface{}) *ConcurrentUrl {
	invTransitions := make([]interfaces.Univalue, 0, len(transitions))

	for i := 0; i < len(transitions); i++ {
		if transitions[i].Persistent() { // Peristent transitions are immune to conflict detection
			invTransitions = append(invTransitions, transitions[i]) //
			transitions[i] = nil                                    // mark the peristent transitions
		}
	}
	common.Remove(&transitions, nil) // Remove the Peristent transitions from the transition lists

	// common.ParallelExecute(
	// 	func() { this.imuImporter.Import(invTransitions, args...) },
	// 	func() { this.importer.Import(transitions, args...) })
	return this
}

// func (this *ConcurrentUrl) Snapshot(preTransitions []interfaces.Univalue) interfaces.Datastore {
// 	// transitions := []interfaces.Univalue(Univalues(common.Clone(this.Export())).To(ITCTransition{}))
// 	// transitions = append(transitions, preTransitions...)

// 	transientDB := ccurlstorage.NewTransientDB(this.WriteCache().Store()) // Should be the same as Importer().Store()
// 	preTransitions = common.Remove(&preTransitions, nil)
// 	Url := NewConcurrentUrl(transientDB).Import(preTransitions).Sort()

// 	ids := Univalues(preTransitions).UniqueTXs()
// 	return snapshotUrl.Commit(ids).Importer().Store() // Commit these changes to the a transient DB
// }

// Call this as s
func (this *ConcurrentUrl) Sort() *ConcurrentUrl {
	// common.ParallelExecute(
	// 	func() { this.imuImporter.SortDeltaSequences() },
	// 	func() { this.importer.SortDeltaSequences() })

	return this
}

func (this *ConcurrentUrl) Finalize(txs []uint32) *ConcurrentUrl {
	// if txs != nil && len(txs) == 0 { // Commit all the transactions when txs == nil
	// 	return this
	// }

	// // this.imuImporter.MergeStateDelta()
	// // this.importer.WhilteList(txs)
	// // this.importer.MergeStateDelta()

	// common.ParallelExecute(
	// 	func() { this.imuImporter.MergeStateDelta() },
	// 	func() {
	// 		this.importer.WhilteList(txs)   // Remove all the transitions generated by the conflicting transactions
	// 		this.importer.MergeStateDelta() // Finalize states
	// 	},
	// )
	return this
}

func (this *ConcurrentUrl) WriteToDbBuffer() [32]byte {
	// keys, values := this.importer.KVs()
	// invKeys, invVals := this.imuImporter.KVs()

	// keys, values = append(keys, invKeys...), append(values, invVals...)
	// return this.importer.Store().Precommit(keys, values) // save the transitions to the DB buffer
	return [32]byte{}
}

func (this *ConcurrentUrl) SaveToDB() {
	// store := this.importer.Store()
	// store.Commit(0) // Commit to the state store
	// this.Clear()
}

func (this *ConcurrentUrl) Commit(txs []uint32) *ConcurrentUrl {
	if txs != nil && len(txs) == 0 {
		this.Clear()
		return this
	}
	this.Finalize(txs)
	this.WriteToDbBuffer() // Export transitions and save them to the DB buffer.
	this.SaveToDB()
	return this
}

func (this *ConcurrentUrl) Export(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) []interfaces.Univalue {
	return this.writeCache.Export(preprocessors...)
}

func (this *ConcurrentUrl) ExportAll(preprocessors ...func([]interfaces.Univalue) []interfaces.Univalue) ([]interfaces.Univalue, []interfaces.Univalue) {
	// all := this.Export(Sorter)
	// // Univalues(all).Print()

	// accesses := Univalues(common.Clone(all)).To(ITCAccess{})
	// transitions := Univalues(common.Clone(all)).To(ITCTransition{})

	return nil, nil
}

func (this *ConcurrentUrl) Print() {
	this.writeCache.Print()
}
