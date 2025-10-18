// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cpus

import (
	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CPUs online", func() {

	It("successfully fetches the list of online logical CPU(s)", func() {
		m := Metadata(gostwire.DiscoveryResult{})
		Expect(m).To(HaveKeyWithValue("cpus",
			Not(BeEmpty())))
	})

})
