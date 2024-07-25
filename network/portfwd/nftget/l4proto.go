// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"
	"golang.org/x/sys/unix"
)

// MetaL4ProtoTcpUdp returns the transport layer protocol name checked for from
// a Meta/Cmp twin-expression, together with the remaining expressions;
// otherwise, it returns nil.
func MetaL4ProtoTcpUdp(exprs nufftables.Expressions) (nufftables.Expressions, string) {
	return nufftables.PrefixedOfTypeTransformed(exprs, isMetaL4Proto, TcpUdp)
}

// PayloadL4ProtoTcpUdp returns the transport layer protocol name checked for
// from a Payload/Cmp twin-expression, together with the remaining expressions;
// otherwise, it returns nil.
func PayloadL4ProtoTcpUdp(exprs nufftables.Expressions) (nufftables.Expressions, string) {
	return nufftables.PrefixedOfTypeTransformed(exprs, isPayloadIPv4L4Proto, TcpUdp)
}

// isMetaL4Proto returns true if the given Meta expression accesses the L4PROTO
// key. See also "Matching Packet metainformation" from the nftables wiki:
// https://wiki.nftables.org/wiki-nftables/index.php/Matching_packet_metainformation
//
// This is the preferred way to do it on especially IPv6, as the “next header”
// isn't necessarily the transport protocol, but an additional header instead.
func isMetaL4Proto(meta *expr.Meta) bool {
	return meta.Key == expr.MetaKeyL4PROTO
}

// isPayloadIPv4L4Proto returns true if the given Payload expression accesses
// the “Protocol” IPv4 header field.
//
// See also [RFC 791, section 3.1] for the following IPv4 header structure; the
// word offsets are in decimal and not shown in the original RFC ASCII
// illustration.
//
//	     0                   1                   2                   3
//	     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+00 |Version|  IHL  |Type of Service|          Total Length         |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+04 |         Identification        |Flags|      Fragment Offset    |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+08 |  Time to Live |    Protocol   |         Header Checksum       |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+12 |                       Source Address                          |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+16 |                    Destination Address                        |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+20 |                    Options                    |    Padding    |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// [RFC 791, section 3.1]: https://datatracker.ietf.org/doc/html/rfc791#section-3.1
func isPayloadIPv4L4Proto(payl *expr.Payload) bool {
	return payl.OperationType == expr.PayloadLoad &&
		payl.Base == expr.PayloadBaseNetworkHeader &&
		payl.Offset == 9 && payl.Len == 1
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
