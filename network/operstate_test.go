// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("operational state", func() {

	DescribeTable("names and icons",
		func(state OperState, name string, icon string) {
			Expect(state.Name()).To(Equal(name))
			Expect(state.TerminalIcon()).To(Equal(icon))
		},
		Entry(nil, Unknown, "UNKNOWN", "◭"),
		Entry(nil, NotPresent, "NOTPRESENT", "✖"),
		Entry(nil, Down, "DOWN", "▽"),
		Entry(nil, LowerLayerDown, "LOWERLAYERDOWN", "⏬"),
		Entry(nil, Testing, "TESTING", "*"),
		Entry(nil, Dormant, "DORMANT", "💤"),
		Entry(nil, Up, "UP", "▲"),
		Entry(nil, OperState(42), "OperState(42)", ""),
	)

})
