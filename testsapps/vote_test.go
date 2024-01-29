package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"

	tests "github.com/arcology-network/eu/tests"
)

func TestParaVote(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/")
	err, _, _ := tests.DeployThenInvoke(targetPath, "examples/vote/vote_mp_test.sol", "0.8.19", "BallotTest", "", []byte{}, false)
	if err != nil {
		t.Error(err.Error())
	}
}
