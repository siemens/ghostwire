// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nftget

import (
	"github.com/google/nftables/expr"
	"github.com/thediveo/nufftables"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("nftables L4 port getter", func() {

	DescribeTable("matches only L4 Port Cmp",
		func(expr expr.Any, expectedPort uint16, expectedOk bool) {
			exprs, port := nufftables.OfTypeTransformed(
				nufftables.Expressions{
					expr,
				},
				Port,
			)
			if expectedOk {
				Expect(exprs).NotTo(BeNil())
			} else {
				Expect(exprs).To(BeNil())
			}
			Expect(port).To(Equal(expectedPort))
		},
		Entry(nil, &expr.Cmp{Op: expr.CmpOpGt}, uint16(0), false),
		Entry(nil, &expr.Cmp{Op: expr.CmpOpEq}, uint16(0), false),
		Entry(nil, &expr.Cmp{
			Op:   expr.CmpOpEq,
			Data: []byte{0x12, 0x34},
		}, uint16(0x1234), true),
	)

})
