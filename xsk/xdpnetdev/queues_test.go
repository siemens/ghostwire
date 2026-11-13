// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xdpnetdev

import (
	"os"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/spacetest/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("netdev queues", func() {

	It("queries the highest RX queue ID", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		defer netns.EnterTransient()()

		dmy := dummy.NewTransient()
		Expect(MaxRxQueueID(dmy.Attrs().Index)).To(BeZero())
	})

	It("returns 0 RX max queue ID for an invalid ifindex", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		defer netns.EnterTransient()()

		Expect(MaxRxQueueID(0)).To(BeZero())
	})

})
