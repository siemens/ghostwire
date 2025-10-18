// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rx/tx queues", func() {

	DescribeTable("QueueType textual representation",
		func(qt QueueType, expected string) {
			Expect(qt.String()).To(Equal(expected))
		},
		Entry("rx type", NETDEV_QUEUE_TYPE_RX, "RX"),
		Entry("tx type", NETDEV_QUEUE_TYPE_TX, "TX"),
		Entry("invalid type", QueueType(42), "QueueType(42)"),
	)

	DescribeTable("comparing QueueTypes for sorting",
		func(qta QueueType, ida int, qtb QueueType, idb int, expected int) {
			Expect(CompareQueuesByIndexAndType(
				&Queue{Type: qta, ID: uint(ida)},
				&Queue{Type: qtb, ID: uint(idb)})).
				To(Equal(expected))
		},
		Entry(nil, NETDEV_QUEUE_TYPE_RX, 42, NETDEV_QUEUE_TYPE_TX, 666, -1),
		Entry(nil, NETDEV_QUEUE_TYPE_RX, 666, NETDEV_QUEUE_TYPE_RX, 42, 1),
		Entry("RX before TX", NETDEV_QUEUE_TYPE_RX, 42, NETDEV_QUEUE_TYPE_TX, 42, -1),
		Entry("TX after RX", NETDEV_QUEUE_TYPE_TX, 42, NETDEV_QUEUE_TYPE_RX, 42, 1),
	)

})
