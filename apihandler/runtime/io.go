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

package runtime

import (
	"fmt"

	"github.com/arcology-network/eu/common"
	intf "github.com/arcology-network/eu/interface"
)

// APIs under the concurrency namespace
type IoHandlers struct {
	api intf.EthApiRouter
}

func NewIoHandlers(api intf.EthApiRouter) *IoHandlers {
	return &IoHandlers{
		api: api,
	}
}
func (this *IoHandlers) Address() [20]byte {
	return common.IO_HANDLER
}

func (this *IoHandlers) Call(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64, _ bool) ([]byte, bool, int64) {
	// signature := codec.Bytes4{}.FromBytes(input[:])
	return this.print(caller, callee, input, origin, nonce)
}

func (this *IoHandlers) print(caller, callee [20]byte, input []byte, origin [20]byte, nonce uint64) ([]byte, bool, int64) {
	fmt.Println("caller:", caller)
	fmt.Println("input:", input)
	fmt.Println("origin:", origin)
	fmt.Println("nonce:", nonce)
	return []byte{}, true, 0
}
