// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"bytes"
	"math"
	"path"

	"github.com/thediveo/faf"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
)

// Definitions for queue-related text elements in /sys/kernel/irq/$IRQ/actions;
// we use lowercase here and [NetdevIRQs] automatically lowercases the actions
// it parses. The rationale is that queue-related action casing is
// driver-specific, so there are drivers that use, for instance, “TxRx” (such as
// Intel i210), while others use use “rxtx” (such as VMware vmxnet3), et cetera.
var (
	sysfsActionRxQueue   = []byte("rx")
	sysfsActionTxQueue   = []byte("tx")
	sysfsActionRxTxQueue = []byte("rxtx")
	sysfsActionTxRxQueue = []byte("txrx")
)

const (
	ifnameActionField = iota
	queueTypeActionField
	queueIDActionField
)

// NetdevIRQs discovers the IRQs allocated to the queues and populates the IRQs
// map of the passed netdev. It additionally links the Queues with their IRQs.
//
// Please note that NetdevIRQs keeps shtumm in view of any errors, so the netdev
// will have an initialized, yet empty IRQs map after NetdevIRQs returns. The
// rationale is that callers will throw away any error details anyway, so we
// here instead cut corners.
//
// Please also note that we here discover the IRQ IDs (numbers) and their links
// to the netdev Queue objects, but not an (optionally) associated IRQ kernel
// thread. The latter is done separately in order to avoid repeatedly scanning
// the list of kernel threads over and over again.
func NetdevIRQs(sysfspath string, netdev *rxtxlayout.Netdev) {
	if netdev.IRQs == nil {
		netdev.IRQs = map[uint]*rxtxlayout.IRQ{}
	}

	irqIter := faf.ReadDir(path.Join(sysfspath, "sys/class/net", netdev.Name, "device/msi_irqs"))
	var actions []byte // reusable buffer for reading pseudo file contents
	for irqEntry := range irqIter {
		if irqEntry.IsDir() {
			continue
		}
		irqno, ok := faf.ParseUint(irqEntry.Name)
		if !ok || irqno > math.MaxUint {
			continue
		}
		irq := &rxtxlayout.IRQ{
			Source: rxtxlayout.SourceSysfs,
			ID:     uint(irqno),
		}
		netdev.IRQs[uint(irqno)] = irq

		actions, ok := faf.ReadFile(
			path.Join(sysfspath, "sys/kernel/irq", string(irqEntry.Name), "actions"),
			actions)
		if !ok {
			continue
		}
		for action := range bytes.SplitSeq(
			bytes.TrimSuffix(actions, []byte("\n")), []byte(",")) {
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
			fields := bytes.Split(action, []byte("-"))
			switch l := len(fields); {
			case l < 3:
				continue
			case l > 3:
				// $IFNAME contains dashes(?), so let's take the final three
				// fields only; and we ignore the butchered first final field
				// anyway.
				fields = fields[l-3:]
			}
			queueID, ok := faf.ParseUint(fields[queueIDActionField])
			if !ok || queueID > math.MaxUint {
				continue
			}
			// note that bytes.EqualFold has a dedicated ASCII fast path, so
			// this even spares us an intermediate allocation due to
			// bytes.ToLower. Nice.
			switch queueType := fields[queueTypeActionField]; {
			case bytes.EqualFold(queueType, sysfsActionRxTxQueue), bytes.EqualFold(queueType, sysfsActionTxRxQueue):
				if rxq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_RX); rxq != nil {
					rxq.IRQ = irq
				}
				if txq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_TX); txq != nil {
					txq.IRQ = irq
				}
			case bytes.EqualFold(queueType, sysfsActionRxQueue):
				if rxq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_RX); rxq != nil {
					rxq.IRQ = irq
				}
			case bytes.EqualFold(queueType, sysfsActionTxQueue):
				if txq := netdev.Queue(uint(queueID), rxtxlayout.NETDEV_QUEUE_TYPE_TX); txq != nil {
					txq.IRQ = irq
				}
			}
		}
	}
}
