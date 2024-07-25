// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nftables L4 proto getter", func() {

	DescribeTable("matches only TCP or UDP Cmp",
		func(expr expr.Any, expectedName string) {
			exprs, protoname := nufftables.OfTypeTransformed(
				nufftables.Expressions{
					expr,
				},
				TcpUdp,
			)
			if expectedName == "" {
				Expect(exprs).To(BeNil())
			} else {
				Expect(exprs).NotTo(BeNil())
			}
			Expect(protoname).To(Equal(expectedName))
		},
		Entry(nil, &expr.Meta{Key: expr.MetaKeyCGROUP}, ""),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, ""),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpEq}, ""),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_ETHERNET},
		}, ""),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_TCP},
		}, "tcp"),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_UDP},
		}, "udp"),
	)

	DescribeTable("matches only TCP or UDP Meta+Cmp combo",
		func(meta *expr.Meta, cmp *expr.Cmp, expectedName string) {
			exprs := nufftables.Expressions{}
			if meta != nil {
				exprs = append(exprs, meta)
			}
			if cmp != nil {
				exprs = append(exprs, cmp)
			}
			exprs, protoname := MetaL4ProtoTcpUdp(exprs)
			if expectedName == "" {
				Expect(exprs).To(BeNil())
			} else {
				Expect(exprs).NotTo(BeNil())
			}
			Expect(protoname).To(Equal(expectedName))
		},
		Entry(nil, nil, nil, ""),
		Entry(nil, &expr.Meta{}, nil, ""),
		Entry(nil, &expr.Meta{
			Key: expr.MetaKeyL4PROTO,
		}, nil, ""),
		Entry(nil, &expr.Meta{
			Key: expr.MetaKeyL4PROTO,
		}, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_ETHERNET},
		}, ""),
		Entry(nil, &expr.Meta{
			Key: expr.MetaKeyL4PROTO,
		}, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_TCP},
		}, "tcp"),
	)

})
