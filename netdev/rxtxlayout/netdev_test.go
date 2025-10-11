// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rxtxstruct model", func() {

	DescribeTable("stringifies Source(s)",
		func(s Source, expected string) {
			Expect(s.String()).To(Equal(expected))
		},
		Entry(nil, Source(0), ""),
		Entry(nil, SourceNetlinkNetdev, "netlink"),
		Entry(nil, SourceSysfs, "sysfs"),
		Entry(nil, SourceSysfs|SourceNetlinkNetdev, "netlink,sysfs"),
	)

	It("stringifies QueueType", func() {
		Expect(QueueType(42).String()).To(Equal("QueueType(42)"))
		Expect(QueueType(NETDEV_QUEUE_TYPE_RX).String()).To(Equal("RX"))
		Expect(QueueType(NETDEV_QUEUE_TYPE_TX).String()).To(Equal("TX"))
	})

	When("looking up netdev queues", func() {

		It("returns nil when Queue cannot be found", func() {
			ndev := &Netdev{
				Queues: []*Queue{
					{ID: 0, Type: NETDEV_QUEUE_TYPE_TX},
					{ID: 1, Type: NETDEV_QUEUE_TYPE_RX},
				},
			}
			Expect(ndev.Queue(42, NETDEV_QUEUE_TYPE_RX)).To(BeNil())
		})

		It("returns the correct Queue", func() {
			ndev := &Netdev{
				Queues: []*Queue{
					{ID: 0, Type: NETDEV_QUEUE_TYPE_TX},
					{ID: 1, Type: NETDEV_QUEUE_TYPE_RX},
				},
			}
			Expect(ndev.Queue(0, NETDEV_QUEUE_TYPE_RX)).To(BeNil())
			Expect(*ndev.Queue(0, NETDEV_QUEUE_TYPE_TX)).To(Equal(*ndev.Queues[0]))
			Expect(ndev.Queue(1, NETDEV_QUEUE_TYPE_TX)).To(BeNil())
			Expect(*ndev.Queue(1, NETDEV_QUEUE_TYPE_RX)).To(Equal(*ndev.Queues[1]))
		})

	})

	It("returns the max queue IDs for RX, TX", func() {
		qs := Queues{
			{ID: 2, Type: NETDEV_QUEUE_TYPE_TX},
			{ID: 42, Type: NETDEV_QUEUE_TYPE_TX},
			{ID: 666, Type: NETDEV_QUEUE_TYPE_RX},
			{ID: 0, Type: NETDEV_QUEUE_TYPE_RX},
		}
		rxid, txid := qs.MaxQueueIDs()
		Expect(rxid).To(Equal(uint(666)))
		Expect(txid).To(Equal(uint(42)))
	})

	When("sorting", func() {

		DescribeTable("Netdevs",
			func(a, b *Netdev, expected int) {
				Expect(SortNetdevsByName(a, b)).To(Equal(expected))
			},
			Entry(nil, &Netdev{Name: "bar"}, &Netdev{Name: "foo"}, -1),
			Entry(nil, &Netdev{Name: "foo"}, &Netdev{Name: "bar"}, 1),
			Entry(nil, &Netdev{Name: "foo"}, &Netdev{Name: "foo"}, 0),
			Entry(nil, &Netdev{Name: "lo"}, &Netdev{Name: "lo"}, 0),
			Entry(nil, &Netdev{Name: "lo"}, &Netdev{Name: "foo"}, -1),
			Entry(nil, &Netdev{Name: "foo"}, &Netdev{Name: "lo"}, 1),
		)

		DescribeTable("Queues",
			func(a, b *Queue, expected int) {
				Expect(SortQueuesByIndexAndType(a, b)).To(Equal(expected))
			},
			Entry(nil, &Queue{ID: 0}, &Queue{ID: 42}, -1),
			Entry(nil, &Queue{ID: 42}, &Queue{ID: 0}, 1),
			Entry(nil, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_RX}, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_TX}, -1),
			Entry(nil, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_TX}, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_RX}, 1),
			Entry(nil, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_TX}, &Queue{ID: 42, Type: NETDEV_QUEUE_TYPE_TX}, 0),
		)

		DescribeTable("IRQs",
			func(a, b *IRQ, expected int) {
				Expect(SortIRQsByID(a, b)).To(Equal(expected))
			},
			Entry(nil, &IRQ{ID: 0}, &IRQ{ID: 42}, -1),
			Entry(nil, &IRQ{ID: 42}, &IRQ{ID: 0}, 1),
			Entry(nil, &IRQ{ID: 42}, &IRQ{ID: 42}, 0),
		)

		DescribeTable("NAPIs",
			func(a, b *NAPI, expected int) {
				Expect(SortNAPIsByID(a, b)).To(Equal(expected))
			},
			Entry(nil, &NAPI{ID: 0}, &NAPI{ID: 42}, -1),
			Entry(nil, &NAPI{ID: 42}, &NAPI{ID: 0}, 1),
			Entry(nil, &NAPI{ID: 42}, &NAPI{ID: 42}, 0),
		)

	})

})
