// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import "fmt"

// QueueType indicates whether a given netdev queue is for RX (receiving) or TX
// (transmitting).
type QueueType uint32

// Netdev queues can be either RX or TX. Please see also the Linux kernel
// source: [netdev.h].
//
// [netdev.h]: https://elixir.bootlin.com/linux/v6.9.5/source/include/uapi/linux/netdev.h#L68
const (
	NETDEV_QUEUE_TYPE_RX QueueType = iota
	NETDEV_QUEUE_TYPE_TX
)

// String returns either “RX” or “TX” (in all upper case), depending on the
// QueueType.
func (q QueueType) String() string {
	switch q {
	case NETDEV_QUEUE_TYPE_RX:
		return "RX"
	case NETDEV_QUEUE_TYPE_TX:
		return "TX"
	}
	return fmt.Sprintf("QueueType(%d)", q)
}

// CompareQueuesByIndexAndType sorts Queue objects first by their numerical ID
// in increasing order, and then RX before TX.
func CompareQueuesByIndexAndType(a, b *Queue) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return int(a.Type) - int(b.Type)
	}
}
