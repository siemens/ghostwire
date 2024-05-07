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
		func(expr expr.Any, expectedName string, expectedOk bool) {
			exprs, protoname := nufftables.OfTypeTransformed(
				nufftables.Expressions{
					expr,
				},
				TcpUdp,
			)
			if expectedOk {
				Expect(exprs).NotTo(BeNil())
			} else {
				Expect(exprs).To(BeNil())
			}
			Expect(protoname).To(Equal(expectedName))
		},
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, "", false),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpEq}, "", false),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_ETHERNET},
		}, "", false),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_TCP},
		}, "tcp", true),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{unix.IPPROTO_UDP},
		}, "udp", true),
	)

})
