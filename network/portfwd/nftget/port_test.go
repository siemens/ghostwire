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

	DescribeTable("matches loading and comparing the L4 destination port",
		func(payl *expr.Payload, expectedPort uint16) {
			remexprs, port := PayloadPort(nufftables.Expressions{
				payl,
				&expr.Cmp{Op: expr.CmpOpEq, Data: []byte{0x12, 0x34}},
				&expr.Counter{},
			})
			Expect(port).To(Equal(expectedPort))
			if expectedPort != 0 {
				Expect(remexprs).To(ConsistOf(BeAssignableToTypeOf(&expr.Counter{})))
			} else {
				Expect(remexprs).To(BeNil())
			}
		},
		Entry("destination port", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseTransportHeader, Offset: 2, Len: 2}, uint16(0x1234)),
		Entry("source port", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseTransportHeader, Offset: 0, Len: 2}, uint16(0)),
		Entry("network header", &expr.Payload{OperationType: expr.PayloadLoad, Base: expr.PayloadBaseNetworkHeader, Offset: 2, Len: 2}, uint16(0)),
	)

})
