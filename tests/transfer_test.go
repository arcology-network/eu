package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

// tests "github.com/arcology-network/vm-adaptor/tests"

func TestTransfer(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/apps/")
	err, _, _ := DeployThenInvoke(targetPath, "transfer/transfer_test.sol", "0.8.19", "Transfer", "transferToContract()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
