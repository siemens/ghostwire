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

	It("returns no IP for other families when there is no IP to match", func() {
		exprs := nufftables.Expressions{&expr.Cmp{Op: expr.CmpOpGt}}
		remexprs, ip := OptionalIPv46(exprs, nftables.TableFamilyINet)
		Expect(remexprs).To(Equal(exprs))
		Expect(ip).To(BeNil())
	})

	DescribeTable("matches loading the destination IP from the network header",
		func(payl *expr.Payload, expected bool) {
			Expect(IsPayloadDestIP(payl)).To(Equal(expected))
		},
		Entry("IPv4", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4}, true),
		Entry("IPv6", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16}, true),
		Entry("IPv4 source", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 12, Len: 4}, false),
		Entry("IPv6 source", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 8, Len: 16}, false),
		Entry("wrong length", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 2}, false),
		Entry("transport header", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseTransportHeader, Offset: 16, Len: 4}, false),
		Entry("write", &expr.Payload{OperationType: expr.PayloadWrite, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4}, false),
	)

	DescribeTable("optionally matches destination IPv4 or IPv6, defaulting to unspecified IP",
		func(exprs nufftables.Expressions, family nftables.TableFamily, expectedRemaining int, expectedIP net.IP) {
			remexprs, ip := OptionalDestIPv46(exprs, family)
			Expect(remexprs).To(HaveLen(expectedRemaining))
			Expect(ip).To(Equal(expectedIP))
		},
		Entry("IPv4 destination",
			nufftables.Expressions{
				&expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 16, Len: 4},
				&expr.Cmp{Op: expr.CmpOpEq, Data: []byte{10, 0, 0, 1}},
			},
			nftables.TableFamilyIPv4, 0, net.IP{10, 0, 0, 1}),
		Entry("IPv6 destination",
			nufftables.Expressions{
				&expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 24, Len: 16},
				&expr.Cmp{Op: expr.CmpOpEq, Data: net.ParseIP("fe80::dead:beef")},
			},
			nftables.TableFamilyIPv6, 0, net.ParseIP("fe80::dead:beef")),
		Entry("no IPv4 destination",
			nufftables.Expressions{&expr.Counter{}},
			nftables.TableFamilyIPv4, 1, net.IP{0, 0, 0, 0}),
		Entry("no IPv6 destination",
			nufftables.Expressions{&expr.Counter{}},
			nftables.TableFamilyIPv6, 1, net.IPv6unspecified),
		Entry("other family",
			nufftables.Expressions{&expr.Counter{}},
			nftables.TableFamilyINet, 1, nil),
	)

})
