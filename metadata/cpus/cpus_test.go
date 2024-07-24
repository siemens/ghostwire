// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cpus

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("CPUs online", func() {

	DescribeTable("parsing CPU ranges",
		func(cpus string, expected model.CPUList) {
			Expect(parseCPUList(cpus)).To(Equal(expected))
		},
		Entry(nil, "", nil),
		Entry(nil, "abc", nil),
		Entry(nil, "42,", nil),
		Entry(nil, "42-", nil),
		Entry(nil, "42-abc", nil),
		Entry(nil, "def-42", nil),
		Entry(nil, "42", model.CPUList{{42, 42}}),
		Entry(nil, "42,43", model.CPUList{{42, 42}, {43, 43}}),
		Entry(nil, "1-42", model.CPUList{{1, 42}}),
		Entry(nil, "1-42,555,666-667", model.CPUList{{1, 42}, {555, 555}, {666, 667}}),
	)

})
