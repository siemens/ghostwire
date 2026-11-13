// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("queue discovery", func() {

	It("reports an error for invalid sysfs path", func() {
		Expect(NetdevQueues("./_test/süsfuss", &rxtxlayout.Netdev{
			Name: "!foobarz",
		})).To(MatchError(
			MatchRegexp(`cannot determine queues of interface .*, reason: .* no such file or directory`)))
	})

	It("reports an error for invalid ifname", func() {
		Expect(NetdevQueues("./_test/twoflower", &rxtxlayout.Netdev{
			Name: "!foobarz",
		})).To(MatchError(
			MatchRegexp(`cannot determine queues of interface .*, reason: .* no such file or directory`)))
	})

	It("discovers queues and ignores nonsense in sys/class/net/.../queues", func() {
		ndev := &rxtxlayout.Netdev{
			Name: "twoflower",
		}
		Expect(NetdevQueues("./_test/twoflower", ndev)).To(Succeed())
		Expect(ndev.Queues).To(ConsistOf(
			And(HaveField("ID", uint(0)), HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX)),
			And(HaveField("ID", uint(0)), HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX)),
			And(HaveField("ID", uint(1)), HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX)),
			And(HaveField("ID", uint(1)), HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX)),
		))
	})

})
