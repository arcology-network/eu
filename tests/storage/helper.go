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

	"github.com/arcology-network/common-lib/exp/deltaset"
	"github.com/arcology-network/common-lib/exp/slice"
	eth "github.com/arcology-network/eu/eth"
	"github.com/arcology-network/eu/gas"
	statestore "github.com/arcology-network/storage-committer"
	stgcommcommon "github.com/arcology-network/storage-committer/common"
	commutative "github.com/arcology-network/storage-committer/type/commutative"
	noncommutative "github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/arcology-network/storage-committer/type/univalue"

	// "github.com/arcology-network/storage-committer/interfaces"
	interfaces "github.com/arcology-network/storage-committer/common"
	cache "github.com/arcology-network/storage-committer/storage/cache"
	stgcommitter "github.com/arcology-network/storage-committer/storage/committer"
	"github.com/arcology-network/storage-committer/storage/proxy"
	stgproxy "github.com/arcology-network/storage-committer/storage/proxy"
)

func GenerateDB(addr [20]uint8) (string, *cache.WriteCache, stgcommcommon.ReadOnlyStore, error) {
	store := chooseDataStore()
	sstore := statestore.NewStateStore(store.(*stgproxy.StorageProxy))
	writeCache := sstore.WriteCache

	acct := CreateAccount(addr)
	if _, err := eth.CreateDefaultPaths(stgcommcommon.SYSTEM, acct, writeCache); err != nil { // NewAccount account structure {
		return acct, nil, nil, errors.New("Failed to create new account: " + err.Error())
	}

	acctTrans := univalue.Univalues(slice.Clone(writeCache.Export(univalue.Sorter))).To(univalue.ITTransition{})
	committer := stgcommitter.NewStateCommitter(store, sstore.GetWriters())
	committer.Import(univalue.Univalues{}.Decode(univalue.Univalues(acctTrans).Encode()).(univalue.Univalues))
	committer.Precommit([]uint64{stgcommcommon.SYSTEM})
	committer.Commit(10)

	return acct, writeCache, store, nil
}

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

func GenPrepayerInfo() (*gas.PrepayerInfo, *gas.PrepayerInfo) {
	Alice := &gas.PrepayerInfo{}
	Alice.Hash = [32]byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06,
		0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
		0x10, 0x11, 0x12, 0x13, 0x14,
		0x15, 0x16, 0x17, 0x18, 0x19, 0x1a,
		0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20}

	Alice.TX = 12345
	Alice.From = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}

	Alice.To = [20]byte{0x11, 0x22, 0x33, 0x44, 0x55,
		0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc,
		0xdd, 0xee, 0xff, 0x12, 0x34, 0x56, 0x78, 0x90}
	Alice.Signature = [4]byte{0x12, 0x34, 0x56, 0x78}
	Alice.GasUsed = 100000
	Alice.InitialGas = 200000
	Alice.PrepayedAmount = 50000
	Alice.GasRemaining = 150000
	Alice.Successful = true

	Alice2 := &gas.PrepayerInfo{}
	Alice2.Hash = [32]byte{0x21, 0x22, 0x23, 0x24, 0x25, 0x26,
		0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
		0x30,
		0x31, 0x32, 0x33, 0x34, 0x35,
		0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b,
		0x3c, 0x3d, 0x3e,
		0x3f, 0x40}

	Alice2.TX = 54321

	Alice2.From = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

	Alice2.To = [20]byte{0x21, 0x22, 0x23, 0x24, 0x25,
		0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c,
		0x2d, 0x2e, 0x2f, 0x30, 0x31, 0x32, 0x33, 0x34}

	Alice2.Signature = [4]byte{0x56, 0x78, 0x9a, 0xbc}
	Alice2.GasUsed = 200000
	Alice2.InitialGas = 300000
	Alice2.PrepayedAmount = 100000
	Alice2.GasRemaining = 200000
	Alice2.Successful = false

	return Alice, Alice2
}
