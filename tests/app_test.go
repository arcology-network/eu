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
	err, _, _ := DeployThenInvoke(targetPath, "examples/vote/vote_mp_test.sol", "0.8.19", "BallotTest", "", []byte{}, false)
	if err != nil {
		t.Error(err.Error())
	}
}

func TestParallelSimpleAuction(t *testing.T) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(filepath.Dir(currentPath)), "concurrentlib/")
	err, _, _ := DeployThenInvoke(targetPath, "examples/simple-auction/simple_auction_mp_test.sol", "0.8.19", "ParalllelSimpleAuctionTest", "", []byte{}, false)
	if err != nil {
		t.Error(err.Error())
	}
}