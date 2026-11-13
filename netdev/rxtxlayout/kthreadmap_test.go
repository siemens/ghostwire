// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("kthread maps", func() {

	It("returns a nil map", func() {
		Expect(NewKthreadMap[uint, string](nil, nil)).To(BeNil())
		Expect(NewKthreadsMap[uint, string](nil, nil)).To(BeNil())
	})

})
