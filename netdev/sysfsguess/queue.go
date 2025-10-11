// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
)

const (
	sysfsRxQueuePrefix = "rx-"
	sysfsTxQueuePrefix = "tx-"
)

// NetdevQueues discovers the queues (with ID and RX/TX type) of the passed
// netdev, populating the passed netdev's Queue field. The queues are discovered
// from the sysfs instance located at the specified sysfspath.
//
// Caveat emptor: we assume that the numbering embedded into the queue
// directories inside the sysfs map directly to queue IDs as used in the NAPI.
//
// Please note that NetdevQueues just discovers queues, but it doesn't discover
// Queue-NAPI and Queue-IRQ relationships.
func NetdevQueues(sysfspath string, ndev *rxtxlayout.Netdev) error {
	queuesDir, err := os.Open(path.Join(sysfspath, "sys/class/net", ndev.Name, "queues"))
	if err != nil {
		return fmt.Errorf("cannot determine queues of interface '%s', reason: %w",
			ndev.Name, err)
	}
	defer queuesDir.Close()

	queueEntries, err := queuesDir.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("cannot determine queues of interface '%s', reason: %w",
			ndev.Name, err)
	}
	queues := []*rxtxlayout.Queue{}
	for _, qEntry := range queueEntries {
		if !qEntry.IsDir() {
			continue
		}
		var qType rxtxlayout.QueueType
		name := qEntry.Name()
		switch {
		case strings.HasPrefix(name, sysfsRxQueuePrefix):
			qType = rxtxlayout.NETDEV_QUEUE_TYPE_RX
		case strings.HasPrefix(name, sysfsTxQueuePrefix):
			qType = rxtxlayout.NETDEV_QUEUE_TYPE_TX
		default:
			continue
		}
		qID, err := strconv.ParseUint(name[3:], 10, 32)
		if err != nil {
			continue
		}
		queues = append(queues, &rxtxlayout.Queue{
			ID:   uint(qID),
			Type: qType,
		})
	}
	ndev.Queues = queues

	return nil
}
