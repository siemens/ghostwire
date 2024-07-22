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

// The unspecified IPv4 address in genuine IPv4 format, and not in IPv4-mapped
// IPv6 address format. See also
// https://www.rfc-editor.org/rfc/rfc5156.html#section-2.2 about IPv4-mapped
// (IPv6) addresses.
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

// OptionalDestIPv46 returns the IPv4 or IPv6 address in a twin Payload network
// header load and Cmp expression; otherwise, it returns the unspecified IP
// address of the particular IP family.
func OptionalDestIPv46(exprs nufftables.Expressions, family nftables.TableFamily) (nufftables.Expressions, net.IP) {
	expr, ip := nufftables.OptionalPrefixedOfTypeTransformed(exprs,
		IsPayloadDestIP, IPv46)
	if ip == nil {
		switch family {
		case nftables.TableFamilyIPv4:
			return exprs, iPv4Unspecified
		case nftables.TableFamilyIPv6:
			return exprs, net.IPv6unspecified
		}
		return exprs, nil
	}
	return expr, ip
}

// IsPayloadDestIP returns true if the passed Payload expression loads an IP
// address of the correct size (either 4 or 16 bytes) from the correct IPv4 or
// IPv6 network layer header.
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
// See also [RFC 8200, section 3] for the following IPv6 header structure; the
// word offsets are in decimal and not shown in the original RFC ASCII
// illustration. Similarly, the bit number heading has been added for improved
// clarification.
//
//	     0                   1                   2                   3
//	     0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+00 |Version| Traffic Class |           Flow Label                  |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+04 |         Payload Length        |  Next Header  |   Hop Limit   |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+08 |                                                               |
//	    +                                                               +
//	+12 |                                                               |
//	    +                         Source Address                        +
//	+16 |                                                               |
//	    +                                                               +
//	+20 |                                                               |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//	+24 |                                                               |
//	    +                                                               +
//	+28 |                                                               |
//	    +                      Destination Address                      +
//	+32 |                                                               |
//	    +                                                               +
//	+36 |                                                               |
//	    +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
//
// [RFC 791, section 3.1]: https://datatracker.ietf.org/doc/html/rfc791#section-3.1
// [RFC 8200, section 3]: https://datatracker.ietf.org/doc/html/rfc8200#section-3
func IsPayloadDestIP(payl *expr.Payload) bool {
	if payl.OperationType != expr.PayloadLoad ||
		payl.Base != expr.PayloadBaseNetworkHeader {
		return false
	}
	switch payl.Len {
	case 4:
		return payl.Offset == 16
	case 16:
		return payl.Offset == 24
	default:
		return false
	}
}
