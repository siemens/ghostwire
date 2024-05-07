// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"net"

	"github.com/google/nftables/expr"
)

// IPv46 returns the IPv4 or IPv6 address enclosed in a Cmp expression,
// otherwise false.
func IPv46(cmp *expr.Cmp) (net.IP, bool) {
	if cmp.Op != expr.CmpOpEq {
		return nil, false
	}
	switch len(cmp.Data) {
	case 4, 16:
		return net.IP(cmp.Data), true
	}
	return nil, false
}
