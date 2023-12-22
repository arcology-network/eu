package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestContainerPair(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "array/bytes_bool_test.sol", "0.8.19", "PairTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256BasicMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "map/u256_test.sol", "0.8.19", "U256MapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestU256ConcurrentMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/")

	err, _, _ := DeployThenInvoke(targetPath, "lib/map/u256_test.sol", "0.8.19", "ConcurrenctU256MapTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestAddressBooleanMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "map/addressBoolean_test.sol", "0.8.19", "AddressBooleanMapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestConcurrentAddressBooleanMap(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "map/addressBoolean_test.sol", "0.8.19", "AddressBooleanMapConcurrentTest", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestStringUint256Map(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/lib/")

	err, _, _ := DeployThenInvoke(targetPath, "map/stringUint256_test.sol", "0.8.19", "StringUint256MapTest", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
