// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nftables L3 IP address getter", func() {

	DescribeTable("matches only IPv4 or IPv6 Cmp",
		func(expr expr.Any, expectedIP net.IP) {
			exprs, ip := nufftables.OfTypeTransformed(
				nufftables.Expressions{
					expr,
				},
				IPv46,
			)
			if expectedIP == nil {
				Expect(exprs).To(BeNil())
			} else {
				Expect(exprs).NotTo(BeNil())
			}
			Expect(ip).To(Equal(expectedIP))
		},
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, nil),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpEq}, nil),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{192, 168, 0, 1},
		}, net.ParseIP("192.168.0.1").To4()),
		Entry(nil, &expr.Cmp{
			Op: expr.CmpOpEq,
			Data: []byte{
				0xfe, 0x80, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xde, 0xad, 0xbe, 0xef,
			},
		}, net.ParseIP("fe80::dead:beef")),
	)

	DescribeTable("optional matches IPv4 or IPv6, defaulting to unspecified IP",
		func(expr expr.Any, family nftables.TableFamily, expectedIP net.IP) {
			exprs, ip := OptionalIPv46(
				nufftables.Expressions{
					expr,
				},
				family,
			)
			if expectedIP == nil {
				Expect(exprs).To(BeNil())
			} else {
				Expect(exprs).NotTo(BeNil())
			}
			Expect(ip).To(Equal(expectedIP))
		},
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, nftables.TableFamilyIPv4, net.IP{0, 0, 0, 0}),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, nftables.TableFamilyIPv6, net.IPv6unspecified),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{192, 168, 0, 1},
		}, nftables.TableFamilyIPv4, net.ParseIP("192.168.0.1").To4()),
		Entry(nil, &expr.Cmp{
			Op: expr.CmpOpEq,
			Data: []byte{
				0xfe, 0x80, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0x00, 0x00, 0x00, 0x00,
				0xde, 0xad, 0xbe, 0xef,
			},
		}, nftables.TableFamilyIPv4, net.ParseIP("fe80::dead:beef")),
	)
})
