package exectest

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	mapi "github.com/arcology-network/common-lib/exp/map"
	scheduler "github.com/arcology-network/scheduler"
	tempcache "github.com/arcology-network/storage-committer/storage/tempcache"
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

	result, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "NumConcurrentInstanceTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result.ReturnData)
}

func TestDeferred(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "DeferredTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestPropertiesToCalleeStruct(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	_, err, eu, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "SequentializerTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
	trans := eu.Api().WriteCache().(*tempcache.WriteCache).Export()

	// Extract callees from the transition set and save them to a dictionary.
	dict := new(scheduler.Callee).ToCallee(trans)
	if len(dict) != 1 {
		t.Error("Expecting 1 callees")
	}

	// Export the callees from the dictionary
	callees := mapi.Values(dict)
	if len(callees[0].Except) != 3 {
		t.Error("Expecting 3 excepts", len(callees[0].Except))
	}

	if !callees[0].Sequential {
		t.Error("Expecting Parallel exection")
	}

	buffer := scheduler.Callees(callees).Encode()

	out := scheduler.Callees{}.Decode(buffer).(scheduler.Callees)
	if len(out) != 1 || !out[0].Equal(callees[0]) {
		t.Error("Expecting the same callees")
	}
}
