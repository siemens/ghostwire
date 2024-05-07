// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"golang.org/x/sys/unix"
)

// TcpUdp returns the transport protocol name enclosed in a Cmp expression
// testing the L4 protocol for TCP and UDP, otherwise false.
func TcpUdp(cmp *expr.Cmp) (string, bool) {
	if cmp.Op != expr.CmpOpEq || len(cmp.Data) != 1 {
		return "", false
	}
	switch cmp.Data[0] {
	case unix.IPPROTO_TCP:
		return "tcp", true
	case unix.IPPROTO_UDP:
		return "udp", true
	}
	return "", false
}
