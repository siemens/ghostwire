// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

type mockNamespace struct {
	model.Namespace
}

var _ = Describe("filling in netdev information", func() {

	It("fills in missing netdevs", func() {
		netnsXXXX := &mockNamespace{}
		netdevmapA := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{Source: SourceNetlinkNetdev, Name: "twoflower"},
			},
		}
		netdevmapB := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{Source: SourceSysfs, Name: "twoflower"},
				{Source: SourceSysfs, Name: "rincewind"},
			},
		}
		ndevs := Fillin(netdevmapA, netdevmapB)
		Expect(ndevs).To(HaveKey(netnsXXXX))
		Expect(ndevs[netnsXXXX]).To(ContainElement(And(
			HaveField("Name", "twoflower"),
			HaveField("Source", SourceNetlinkNetdev))))
		Expect(ndevs[netnsXXXX]).To(ContainElement(And(
			HaveField("Name", "rincewind"),
			HaveField("Source", SourceSysfs))))
	})

	It("fills in missing IRQs", func() {
		netnsXXXX := &mockNamespace{}
		netdevmapA := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{
					Source: SourceNetlinkNetdev,
					Name:   "twoflower",
					IRQs:   map[uint]*IRQ{},
				},
			},
		}
		netdevmapB := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{
					Source: SourceSysfs,
					Name:   "twoflower",
					IRQs: map[uint]*IRQ{
						42: {
							Source: SourceSysfs,
							ID:     42,
						},
					},
				},
			},
		}
		ndevs := Fillin(netdevmapA, netdevmapB)
		Expect(ndevs).To(HaveKey(netnsXXXX))
		Expect(ndevs[netnsXXXX]).To(ContainElement(And(
			HaveField("Name", "twoflower"),
			HaveField("IRQs", HaveKeyWithValue(
				uint(42),
				HaveField("Source", SourceSysfs))))))
	})

	It("fills in missing IRQs and links them to their Queue(s)", func() {
		netnsXXXX := &mockNamespace{}
		netdevmapA := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{
					Source: SourceNetlinkNetdev,
					Name:   "twoflower",
					Queues: Queues{
						{},
					},
					IRQs: map[uint]*IRQ{
						667: {
							ID: 667,
						},
					},
				},
			},
		}
		netdevmapA[netnsXXXX][0].Queues = Queues{
			{
				ID: 0,
			},
			{
				ID:   1,
				Type: NETDEV_QUEUE_TYPE_RX,
				IRQ:  netdevmapA[netnsXXXX][0].IRQs[42],
			},
			{
				ID:   1,
				Type: NETDEV_QUEUE_TYPE_TX,
				IRQ:  netdevmapA[netnsXXXX][0].IRQs[42],
			},
		}

		netdevmapB := NetdevsByNetns{
			netnsXXXX: []*Netdev{
				{
					Source: SourceSysfs,
					Name:   "twoflower",
					IRQs: map[uint]*IRQ{
						42: {
							Source: SourceSysfs,
							ID:     42,
						},
						667: {
							ID: 667,
						},
					},
				},
			},
		}
		netdevmapB[netnsXXXX][0].Queues = Queues{
			{
				ID:   0,
				Type: NETDEV_QUEUE_TYPE_RX,
				IRQ:  netdevmapB[netnsXXXX][0].IRQs[42],
			},
			{
				ID:   0,
				Type: NETDEV_QUEUE_TYPE_TX,
				IRQ:  &IRQ{ID: 123},
			},
			{
				ID:   1,
				Type: NETDEV_QUEUE_TYPE_TX,
				IRQ:  netdevmapB[netnsXXXX][0].IRQs[42],
			},
			{
				ID:   666,
				Type: NETDEV_QUEUE_TYPE_TX,
				IRQ:  netdevmapB[netnsXXXX][0].IRQs[42],
			},
			{
				ID:   667,
				Type: NETDEV_QUEUE_TYPE_TX,
				IRQ:  netdevmapB[netnsXXXX][0].IRQs[667],
			},
		}

		ndevs := Fillin(netdevmapA, netdevmapB)
		Expect(ndevs).To(HaveKey(netnsXXXX))
		Expect(ndevs[netnsXXXX]).To(ContainElement(And(
			HaveField("Name", "twoflower"),
			HaveField("Queues", ContainElements(
				And(
					HaveField("Type", NETDEV_QUEUE_TYPE_RX),
					HaveField("IRQ", HaveField("ID", uint(42)))),
				And(
					HaveField("Type", NETDEV_QUEUE_TYPE_TX),
					HaveField("IRQ", HaveField("ID", uint(42)))),
			),
			))))
	})

})
