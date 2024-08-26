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

package committertest

import (
	"errors"
	"reflect"

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	commutative "github.com/arcology-network/common-lib/types/storage/commutative"
	noncommutative "github.com/arcology-network/common-lib/types/storage/noncommutative"
	"github.com/arcology-network/common-lib/types/storage/univalue"
	statestore "github.com/arcology-network/storage-committer"

	// "github.com/arcology-network/storage-committer/interfaces"
	interfaces "github.com/arcology-network/common-lib/types/storage/common"
	cache "github.com/arcology-network/common-lib/types/storage/writecache"
	"github.com/arcology-network/storage-committer/storage/proxy"
)

func Create_Ctrn_0(account string, store interfaces.ReadOnlyStore) ([]byte, []*univalue.Univalue, error) {
	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache

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

	rawTrans := writeCache.Export(univalue.Sorter)
	transitions := univalue.Univalues(slice.Clone(rawTrans)).To(univalue.ITTransition{})
	return univalue.Univalues(transitions).Encode(), transitions, nil
}

func ParallelInsert_Ctrn_0(account string, store interfaces.ReadOnlyStore) ([]byte, error) {

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
	path := commutative.NewPath() // create a path
	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", path); err != nil {
		return []byte{}, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-00", noncommutative.NewString("tx0-elem-00")); err != nil { /* The first Element */
		return []byte{}, err
	}

	if _, err := writeCache.Write(0, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/elem-01", noncommutative.NewString("tx0-elem-01")); err != nil { /* The second Element */
		return []byte{}, err
	}

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	return univalue.Univalues(transitions).Encode(), nil
}

func Create_Ctrn_1(account string, store interfaces.ReadOnlyStore) ([]byte, error) {

	sstore := statestore.NewStateStore(store.(*proxy.StorageProxy))
	writeCache := sstore.WriteCache
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

	transitions := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	return univalue.Univalues(transitions).Encode(), nil
}

func CheckPaths(account string, writeCache *cache.WriteCache) error {
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
	keys := v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	// Read the path again
	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-0/", new(commutative.Path))
	keys = v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}

	v, _, _ = writeCache.Read(1, "blcc://eth1.0/account/"+account+"/storage/ctrn-1/", new(commutative.Path))
	keys = v.(*deltaset.DeltaSet[string]).Elements()
	if !reflect.DeepEqual(keys, []string{"elem-00", "elem-01"}) {
		return errors.New("Error: Path don't match !")
	}
	return nil
}
