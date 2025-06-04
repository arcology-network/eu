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

package exectest

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestResettable(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "storage/storage_test.sol", "0.8.19", "ResettableDeployer", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestInstances(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "NumConcurrentInstanceTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
	// fmt.Println(result.ReturnData)

	// _, err, _, _ = DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "NumConcurrentInstanceTest", "call2()", []byte{}, false)
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(result.ReturnData)
}

func TestDeferred(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "DeferredTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestTopup(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "TopupGasTest", "init()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestPrint(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "PrintTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

// func TestPropertiesToCalleeStruct(t *testing.T) {
// 	currentPath, _ := os.Getwd()
// 	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

// 	_, err, eu, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "SequentializerTest", "", []byte{}, false)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	trans := eu.Api().WriteCache().(*tempcache.WriteCache).Export()
// 	univalue.Univalues(trans).Print()

// 	// Extract callees from the transition set and save them to a dictionary.
// 	dict := new(scheduler.Callee).ToCallee(trans)
// 	if len(dict) != 1 {
// 		t.Error("Expecting 1 callees")
// 	}

// 	// Export the callees from the dictionary
// 	callees := mapi.Values(dict)
// 	if len(callees[0].Except) != 3 {
// 		t.Error("Expecting 3 excepts", len(callees[0].Except))
// 	}

// 	if !callees[0].Sequential {
// 		t.Error("Expecting Parallel exection")
// 	}

// 	buffer := scheduler.Callees(callees).Encode()

// 	out := scheduler.Callees{}.Decode(buffer).(scheduler.Callees)
// 	if len(out) != 1 || !out[0].Equal(callees[0]) {
// 		t.Error("Expecting the same callees")
// 	}
// }
