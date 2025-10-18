// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("sorting things", func() {

	DescribeTable("Netdevs",
		func(nameA, nameB string, expected int) {
			Expect(CompareNetdevsByName(
				&Netdev{Name: nameA},
				&Netdev{Name: nameB})).To(Equal(expected))
		},
		Entry(nil, "bar", "foo", -1),
		Entry(nil, "foo", "bar", 1),
		Entry(nil, "foo", "foo", 0),
		Entry(nil, "lo", "lo", 0),
		Entry(nil, "lo", "foo", -1),
		Entry(nil, "foo", "lo", 1),
	)

	DescribeTable("IRQs",
		func(ida, idb int, expected int) {
			Expect(CompareIRQsByID(
				&IRQ{ID: uint(ida)},
				&IRQ{ID: uint(idb)})).To(Equal(expected))
		},
		Entry(nil, 0, 42, -1),
		Entry(nil, 42, 0, 1),
		Entry(nil, 42, 42, 0),
	)

	DescribeTable("NAPIs",
		func(ida, idb int, expected int) {
			Expect(CompareNAPIsByID(
				&NAPI{ID: uint(ida)},
				&NAPI{ID: uint(idb)})).
				To(Equal(expected))
		},
		Entry(nil, 0, 42, -1),
		Entry(nil, 42, 0, 1),
		Entry(nil, 42, 42, 0),
	)

})
