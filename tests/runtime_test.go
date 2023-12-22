package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestResettable(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "storage/storage_test.sol", "0.8.19", "ResettableDeployer", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
