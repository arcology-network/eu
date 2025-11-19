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

package stgtest

import (
	"errors"
	"reflect"

	"github.com/arcology-network/common-lib/exp/slice"
	"github.com/arcology-network/common-lib/exp/softdeltaset"
	eth "github.com/arcology-network/eu/eth"

	// "github.com/arcology-network/eu/gas"
	statestore "github.com/arcology-network/state-engine"
	stgcommon "github.com/arcology-network/state-engine/common"
	commutative "github.com/arcology-network/state-engine/type/commutative"
	noncommutative "github.com/arcology-network/state-engine/type/noncommutative"
	statecell "github.com/arcology-network/state-engine/type/statecell"

	// "github.com/arcology-network/state-engine/interfaces"
	interfaces "github.com/arcology-network/state-engine/common"
	cache "github.com/arcology-network/state-engine/state/cache"
	statecommitter "github.com/arcology-network/state-engine/state/committer"
	"github.com/arcology-network/state-engine/storage/proxy"
)

func GenerateDB(addr [20]uint8) (string, *cache.StateCache, stgcommon.ReadOnlyStore, error) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	acct := CreateAccount(addr)
	if _, err := eth.CreateDefaultPaths(stgcommon.SYSTEM, acct, writeCache); err != nil { // NewAccount account structure {
		return acct, nil, nil, errors.New("Failed to create new account: " + err.Error())
	}

	acctTrans := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	committer := statecommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(statecell.StateCells{}.Decode(statecell.StateCells(acctTrans).Encode()).(statecell.StateCells))
	committer.Precommit([]uint64{stgcommon.SYSTEM})
	committer.Commit(10)

	return acct, writeCache, store, nil
}

func Create_Ctrn_0(account string, store interfaces.ReadOnlyStore) ([]byte, []*statecell.StateCell, error) {
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache

	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
		return []byte{}, nil, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
		return []byte{}, nil, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
		return []byte{}, nil, err
	}

	rawTrans := writeCache.Export(statecell.Sorter)
	transitions := statecell.StateCells(slice.Clone(rawTrans)).To(statecell.ITTransition{})
	return statecell.StateCells(transitions).Encode(), transitions, nil
}

// func ParallelInsert_Ctrn_0(account string, store interfaces.ReadOnlyStore) ([]byte, error) {
// 	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
// 	writeCache := sstore.StateCache
// 	path := commutative.NewPath() // create a path
// 	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
// 		return []byte{}, err
// 	}

// 	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
// 		return []byte{}, err
// 	}

// 	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
// 		return []byte{}, err
// 	}

// 	transitions := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
// 	return statecell.StateCells(transitions).Encode(), nil
// }

func Create_Ctrn_1(account string, store interfaces.ReadOnlyStore) ([]byte, error) {
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.StateCache
	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", path); err != nil {
		return []byte{}, err
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00", noncommutative.NewString("tx1-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if _, err := writeCache.Write(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01", noncommutative.NewString("tx1-elem-00")); err != nil { /* The second Element */
		return []byte{}, err
	}

	transitions := statecell.StateCells(slice.Clone(writeCache.Export(statecell.Sorter))).To(statecell.ITTransition{})
	return statecell.StateCells(transitions).Encode(), nil
}

func CheckPaths(account string, writeCache *cache.StateCache) error {
	v, _, _ := writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", new(noncommutative.String))
	if v.(string) != "tx0-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", new(noncommutative.String))
	if v.(string) != "tx0-elem-01" {
		return errors.New("Error: Not match")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-00", new(noncommutative.String))
	if v.(string) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/elem-01", new(noncommutative.String))
	if v.(string) != "tx1-elem-00" {
		return errors.New("Error: Not match")
	}

	//Read the path
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", new(commutative.Path))
	keys := v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	// Read the path again
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", new(commutative.Path))
	keys = v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", new(commutative.Path))
	keys = v.(*softdeltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}
	return nil
}
