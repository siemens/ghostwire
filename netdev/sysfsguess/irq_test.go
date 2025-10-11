// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IRQ discovery", func() {

	Context("queue IRQs", func() {

		It("reports an error for invalid sysfs path", func() {
			Expect(NetdevIRQs("./_test/süsfuss", &rxtxlayout.Netdev{
				Name: "!foobarz",
			})).To(MatchError(
				MatchRegexp(`cannot determine IRQs of interface .*, reason: .* no such file or directory`)))
		})

		It("reports an error for invalid ifname", func() {
			Expect(NetdevIRQs("./_test/twoflower", &rxtxlayout.Netdev{
				Name: "!foobarz",
			})).To(MatchError(
				MatchRegexp(`cannot determine IRQs of interface .*, reason: .* no such file or directory`)))
		})

		It("discovers IRQs for queues and ignores nonsense in sys/kernel/irq/...", func() {
			ndev := &rxtxlayout.Netdev{
				Name: "twoflower",
			}
			Expect(NetdevQueues("./_test/twoflower", ndev)).To(Succeed())
			Expect(NetdevIRQs("./_test/twoflower", ndev)).To(Succeed())
			Expect(ndev.Queues).To(ConsistOf(
				And(
					HaveField("ID", uint(0)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX),
					HaveField("IRQ", HaveField("ID", uint(42)))),
				And(HaveField("ID", uint(0)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX),
					HaveField("IRQ", HaveField("ID", uint(42)))),
				And(HaveField("ID", uint(1)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX),
					HaveField("IRQ", HaveField("ID", uint(666)))),
				And(HaveField("ID", uint(1)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX),
					HaveField("IRQ", HaveField("ID", uint(666)))),
			))
		})

	})

	Context("IRQ kthreads", func() {

		It("relates netdev IRQs to IRQ kthreads", func() {
			ndev := &rxtxlayout.Netdev{
				Name: "twoflower",
			}
			Expect(NetdevQueues("./_test/twoflower", ndev)).To(Succeed())
			Expect(NetdevIRQs("./_test/twoflower", ndev)).To(Succeed())

			procs := model.ProcessTable{
				2: &model.Process{},
				13: &model.Process{
					PID: 13,
					ProTaskCommon: model.ProTaskCommon{
						Name: "irq/42-twoflower",
					},
				},
				14: &model.Process{
					PID: 14,
					ProTaskCommon: model.ProTaskCommon{
						Name: "irq/666-twoflower",
					},
				},
			}
			procs[2].Children = []*model.Process{
				procs[13], procs[14],
			}
			irqkthreads := rxtxlayout.NewIRQKthreadsMap(procs)
			rxtxlayout.FillInIRQKthreads(irqkthreads, []*rxtxlayout.Netdev{
				ndev,
			})
			Expect(ndev.IRQs).To(ConsistOf(
				And(HaveField("ID", uint(42)),
					HaveField("PID", model.PIDType(13)),
					HaveField("Kthread", HaveField("PID", model.PIDType(13)))),
				And(HaveField("ID", uint(666)),
					HaveField("PID", model.PIDType(14)),
					HaveField("Kthread", HaveField("PID", model.PIDType(14))))))
		})

	})

})
