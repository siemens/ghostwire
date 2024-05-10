// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"
	"golang.org/x/sys/unix"
)

// L4ProtoTcpUdp returns the transport layer protocol name checked for from a
// Meta/Cmp combo together with the remaining expressions; otherwise, it returns
// nil.
func L4ProtoTcpUdp(exprs nufftables.Expressions) (nufftables.Expressions, string) {
	exprs, _ = nufftables.OfTypeFunc(exprs, isL4Proto)
	if exprs == nil {
		return nil, ""
	}
	exprs, proto := nufftables.OfTypeTransformed(exprs, TcpUdp)
	return exprs, proto
}

// isL4Proto returns true if the given Meta expression accesses the L4PROTO key.
// See also "Matching Packet metainformation" from the nftables wiki:
// https://wiki.nftables.org/wiki-nftables/index.php/Matching_packet_metainformation
func isL4Proto(meta *expr.Meta) bool {
	return meta.Key == expr.MetaKeyL4PROTO
}

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
