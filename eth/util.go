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

package eth

import (
	"reflect"

	stgcommcom "github.com/arcology-network/storage-committer/platform"
	"github.com/arcology-network/storage-committer/type/commutative"
	"github.com/arcology-network/storage-committer/type/noncommutative"
	"github.com/arcology-network/storage-committer/type/univalue"
)

// CreateNewAccount creates a new account in the write cache.
// It returns the transitions and an error, if any.
func CreateNewAccount(tx uint64, acct string, store interface {
	IfExists(string) bool
	Write(uint64, string, interface{}, ...any) (int64, error)
}) ([]*univalue.Univalue, error) {

	// accountRoot := common.StrCat(ccurlcommon.ETH10_ACCOUNT_PREFIX, string(acct), "/")
	// if !this.apiRouter.WriteCache().(*cache.WriteCache).IfExists(accountRoot)

	paths, typeids := stgcommcom.NewPlatform().GetBuiltins(acct)

	transitions := []*univalue.Univalue{}
	for i, path := range paths {
		var v any
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
		if !store.IfExists(path) {
			transitions = append(transitions, univalue.NewUnivalue(tx, path, 0, 1, 0, v, nil))

			if _, err := store.Write(tx, path, v); err != nil { // root path
				store.Write(tx, path, v)
				return nil, err
			}

			if !store.IfExists(path) {
				_, err := store.Write(tx, path, v)
				return transitions, err // root path
			}
		}
	}
	return transitions, nil
}

// This function is used for Multiprocessor execution ONLY !!!.
// This function converts a list of raw calls to a list of parallel job sequences. One job sequence is created for each caller.
// If there are N callers, there will be N job sequences. There sequences will be later added to a generation and executed in parallel.
// func NewGenerationFromMsgs(id uint32, numThreads uint8, evmMsgs []*evmcore.Message, api typeexec.EthApiRouter) *Generation {
// 	gen := NewGeneration(id, uint8(len(evmMsgs)), []*JobSequence{})
// 	slice.Foreach(evmMsgs, func(i int, msg **evmcore.Message) {
// 		gen.Add(new(JobSequence).NewFromCall(*msg, api.GetEU().(interface{ TxHash() [32]byte }).TxHash(), api))
// 	})
// 	gen.occurrences = gen.OccurrenceDict(gen.jobSeqs)
// 	api.SetSchedule(gen.occurrences)
// 	return gen
// }
