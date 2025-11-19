/*
 *   Copyright (c) 2024 Arcology Network

 *   This program is free software: you can redistribute it and/or modify
 *   it under the terms of the GNU General Public License as published by
 *   the Free Software Foundation, either version 3 of the License, or
 *   (at your option) any later version.

 *   This program is distributed in the hope that it will be useful,
 *   but WITHOUT ANY WARRANTY; without even the implied warranty of
 *   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *   GNU General Public License for more details.

 *   You should have received a copy of the GNU General Public License
 *   along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package exectest

import (
	"os"
	"path"
	"path/filepath"
	"testing"
)

func BenchmarkReverseString10k(b *testing.B) {
	currentPath, _ := os.Getwd()
	targetPath := path.Join(path.Dir(path.Dir(filepath.Dir(currentPath))), "concurrent/lib/")

	for i := 0; i < 10; i++ {
		DeployThenInvoke(targetPath, "multiprocess/mp_benchmarking.sol", "0.8.19", "MpBenchmarking", "benchmarkReverseString10k()", []byte{}, false)
	}
}
