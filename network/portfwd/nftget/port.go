// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"
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

// PayloadPort returns the (transport) destination port number from a
// payload-cmp twin expression, together with the remaining expressions;
// otherwise, it returns nil expressions and a zero port.
func PayloadPort(exprs nufftables.Expressions) (nufftables.Expressions, uint16) {
	return nufftables.PrefixedOfTypeTransformed(exprs, isL4DestPort, Port)
}

// isL4DestPort returns true if the Payload expression addresses the
// “destination port” field in a TCP or UDP transport header.
func isL4DestPort(payl *expr.Payload) bool {
	return payl.OperationType == expr.PayloadLoad &&
		payl.Base == expr.PayloadBaseTransportHeader &&
		payl.Offset == 2 && payl.Len == 2
}
