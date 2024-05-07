// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"net"

	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nftables L3 IP address getter", func() {

	DescribeTable("matches only IPv4 or IPv6 Cmp",
		func(expr expr.Any, expectedIP net.IP, expectedOk bool) {
			exprs, ip := nufftables.OfTypeTransformed(
				nufftables.Expressions{
					expr,
				},
				IPv46,
			)
			if expectedOk {
				Expect(exprs).NotTo(BeNil())
			} else {
				Expect(exprs).To(BeNil())
			}
			Expect(ip).To(Equal(expectedIP))
		},
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, nil, false),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpEq}, nil, false),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{192, 168, 0, 1},
		}, net.ParseIP("192.168.0.1").To4(), true),
		Entry(nil, &expr.Cmp{
			Op: expr.CmpOpEq,
			Data: []byte{
				0xfe, 0x80, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xde, 0xad, 0xbe, 0xef,
			},
		}, net.ParseIP("fe80::dead:beef"), true),
	)

})
