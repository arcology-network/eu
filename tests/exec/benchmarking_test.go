package tests

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func BenchmarkReverseString10k(b *testing.B) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrentlib/lib/")

	for i := 0; i < 10; i++ {
		DeployThenInvoke(targetPath, "multiprocess/mp_benchmarking.sol", "0.8.19", "MpBenchmarking", "benchmarkReverseString10k()", []byte{}, false)
	}
}
