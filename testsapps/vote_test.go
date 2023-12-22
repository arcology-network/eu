package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func TestParaVote(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/")
	err, _, _ := DeployThenInvoke(targetPath, "examples/vote/vote_parallelized_mp_test.sol", "0.8.19", "ParaBallotCaller", "", []byte{}, false)
	if err != nil {
		t.Error(err)
	}
}
