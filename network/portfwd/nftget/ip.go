// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"
)

var iPv4Unspecified = net.IP{0, 0, 0, 0}

// OptionalIPv46 returns the IPv4 or IPv6 address enclosed in a Cmp expression,
// otherwise the unspecified IP address of the particular IP family.
func OptionalIPv46(exprs nufftables.Expressions, family nftables.TableFamily) (nufftables.Expressions, net.IP) {
	remexpr, ip := nufftables.OfTypeTransformed(exprs, IPv46)
	if remexpr == nil {
		switch family {
		case nftables.TableFamilyIPv4:
			return exprs, iPv4Unspecified
		case nftables.TableFamilyIPv6:
			return exprs, net.IPv6unspecified
		}
		return exprs, nil
	}
	return remexpr, ip
}

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
