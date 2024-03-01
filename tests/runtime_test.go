package tests

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestResettable(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "storage/storage_test.sol", "0.8.19", "ResettableDeployer", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestInstances(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	result, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "NumConcurrentInstanceTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
	fmt.Println(result.ReturnData)
}

func TestDeferred(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	_, err, _, _ := DeployThenInvoke(targetPath, "runtime/Runtime_test.sol", "0.8.19", "DeferredTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
