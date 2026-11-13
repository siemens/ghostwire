// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package xfs

import (
	"github.com/thediveo/nonstd/xiter"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("iterating over double-star globs", func() {

	It("yields correct matches in arbitrary deep subdirectories", func() {
		Expect(xiter.Keys(Glob("./_testfs/**/*.none"))).To(ConsistOf(
			"_testfs/foo/abc.none",
			"_testfs/foo/bar/baz/def.none",
			"_testfs/foo/bar/baz/xyz.none",
		))
	})

	It("yields nothing", func() {
		Expect(xiter.Keys(Glob("./_testfs/**/*.gnampf"))).To(BeEmpty())
	})

	It("aborts correctly", func() {
		count := 0
		for range Glob("./_testfs/**/*.none") {
			count++
			break
		}
		Expect(count).To(Equal(1))
	})

})
