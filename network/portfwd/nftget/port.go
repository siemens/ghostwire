// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
)

// Port returns the (transport) port number from a Cmp expression; otherwise,
// returns false.
func Port(cmp *expr.Cmp) (uint16, bool) {
	if cmp.Op != expr.CmpOpEq || len(cmp.Data) != 2 {
		return 0, false
	}
	// port number is always in network order, that is, big endian.
	return uint16(cmp.Data[0])<<8 + uint16(cmp.Data[1]), true
}
