// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xdpnetdev

import (
	"github.com/siemens/ghostwire/v2/netdev/nlnetdev"
)

// MaxRxQueueID returns the maximum allowed queue ID for the specified network
// interface. Silently assumes highest queue ID 0 if the network interface queue
// information cannot be retrieved.
func MaxRxQueueID(ifindex int) uint32 {
	conn, err := nlnetdev.Dial(nil)
	if err != nil {
		return 0
	}
	defer conn.Close()
	queues, err := conn.Queues(uint32(ifindex))
	if err != nil {
		return 0
	}
	maxRxQueueID, _ := queues.MaxQueueIDs()
	return uint32(maxRxQueueID)
}
