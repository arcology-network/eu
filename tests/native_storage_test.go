package tests

import (
	"encoding/hex"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/holiman/uint256"
	sha3 "golang.org/x/crypto/sha3"
)

func TestSlotHash(t *testing.T) {
	_ctrn := uint256.NewInt(2).Bytes32()
	_elem := uint256.NewInt(2).Bytes32()

	hash := sha3.NewLegacyKeccak256()
	hash.Write(append(_elem[:], _ctrn[:]...))
	v := hash.Sum(nil)
	fmt.Println(v)
	fmt.Println("0x" + hex.EncodeToString(v))
}

func TestNativeStorage(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/native/")
	err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "NativeStorage", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}

func TestGasDebitInFailedTx(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/native/")
	err, _, _ := DeployThenInvoke(targetPath, "NativeStorage.sol", "0.8.19", "TestFailed", "call()", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
