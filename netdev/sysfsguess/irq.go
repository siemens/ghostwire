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

// Definitions for queue-related text elements in /sys/kernel/irq/$IRQ/actions;
// we use lowercase here and [NetdevIRQs] automatically lowercases the actions
// it parses. The rationale is that queue-related action casing is
// driver-specific, so there are drivers that use, for instance, “TxRx” (such as
// Intel i210), while others use use “rxtx” (such as VMware vmxnet3), et cetera.
const (
	sysfsActionRxQueue   = "rx"
	sysfsActionTxQueue   = "tx"
	sysfsActionRxTxQueue = "rxtx"
	sysfsActionTxRxQueue = "txrx"
)

const (
	ifnameActionField = iota
	queueTypeActionField
	queueIDField
)

// NetdevIRQs discovers the IRQs allocated to the queues and populates the IRQs
// map of the passed netdev. It additionally links the Queues with their IRQs.
//
// Please note that we here discover the IRQ IDs (numbers) and their links to
// the netdev Queue objects, but not an (optionally) associated IRQ kernel
// thread. The latter is done separately in order to avoid repeatedly scanning
// the list of kernel threads over and over again.
func NetdevIRQs(sysfspath string, netdev *rxtxlayout.Netdev) error {
	irqsDir, err := os.Open(path.Join(sysfspath, "sys/class/net", netdev.Name, "device/msi_irqs"))
	if err != nil {
		return fmt.Errorf("cannot determine IRQs of interface '%s', reason: %w",
			netdev.Name, err)
	}
	defer irqsDir.Close()

	irqEntries, err := irqsDir.ReadDir(-1)
	if err != nil {
		return fmt.Errorf("cannot determine IRQs of interface '%s', reason: %w",
			netdev.Name, err)
	}

	if netdev.IRQs == nil {
		netdev.IRQs = map[uint]*rxtxlayout.IRQ{}
	}

	for _, irqEntry := range irqEntries {
		if irqEntry.IsDir() {
			continue
		}
		irqno, err := strconv.ParseUint(irqEntry.Name(), 10, 32)
		if err != nil {
			continue
		}
		irq := &rxtxlayout.IRQ{
			Source: rxtxlayout.SourceSysfs,
			ID:     uint(irqno),
		}
		netdev.IRQs[uint(irqno)] = irq

		actions, err := os.ReadFile(path.Join(sysfspath, "sys/kernel/irq", irqEntry.Name(), "actions"))
		if err != nil {
			continue
		}
		for _, action := range strings.Split(strings.TrimSuffix(string(actions), "\n"), ",") {
			// action format:
			//   $IFNAME-$QUEUETYPE-$QUEUEID
			// see:
			// https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-kernel-irq
			//
			// Note: when a link gets renamed, then the action won't reflect the
			// new name, so we can only ignore the action's $IFNAME, not
			// checking for consistency ... because there isn't any to rely on.
			//
			// Note #2: let's hope that this doesn't blow up in our faces when
			// there are udev-originating netdev names containing dashes... :(
			fields := strings.Split(action, "-")
			switch l := len(fields); {
			case l < 3:
				continue
			case l > 3:
				// $IFNAME contains dashes(?), so let's take the final three
				// fields only; and we ignore the butchered first final field
				// anyway.
				fields = fields[l-3:]
			}
			queueID, err := strconv.ParseUint(fields[queueIDField], 10, 32)
			if err != nil {
				continue
			}
			switch strings.ToLower(fields[queueTypeActionField]) {
			case sysfsActionRxTxQueue, sysfsActionTxRxQueue:
				if rxq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_RX); rxq != nil {
					rxq.IRQ = irq
				}
				if txq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_TX); txq != nil {
					txq.IRQ = irq
				}
			case "rx":
				if rxq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_RX); rxq != nil {
					rxq.IRQ = irq
				}
			case "tx":
				if txq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_TX); txq != nil {
					txq.IRQ = irq
				}
			}

		}
	}

	return nil
}
