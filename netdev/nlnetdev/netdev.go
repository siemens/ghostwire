// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"fmt"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/thediveo/lxkns/model"
	"golang.org/x/exp/maps"
)

// Discover the netdevs with their queues, NAPIs (including kernel threads), and
// IRQs.
//
// The following bits (or sometimes chunks) of information are missing:
//   - Netdevs in operstate DOWN are skipped by the Linux kernel's NETLINK netdev
//     API.
//   - Netdev.Name, as the NETLINK netdev API is ifindex-based, so we need to fill
//     in this afterwards.
//   - Netdev.OriginalName is user space anyway, so we need to fill this in
//     separately.
func (c *Conn) Netdevs(irqkthreads rxtxlayout.IRQKthreadMap, procs model.ProcessTable) ([]*rxtxlayout.Netdev, error) {
	allqueues, err := c.queues(0)
	if err != nil {
		return nil, fmt.Errorf("cannot list netdev queues, reason: %w", err)
	}
	allnapis, err := c.napis(0)
	if err != nil {
		return nil, fmt.Errorf("cannot list netdev NAPIs, reason: %w", err)
	}

	ndevsmap := map[int]*rxtxlayout.Netdev{}
	// Create the NAPI and IRQ objects first, as we come across them when
	// processing the NETLINK netdev API answer, because the queue objects will
	// hopefully reference them next.
	for _, napi := range allnapis {
		ndev, ok := ndevsmap[int(napi.IfIndex)]
		if !ok {
			ndev = &rxtxlayout.Netdev{
				Source: rxtxlayout.SourceNetlinkNetdev,
				Index:  int(napi.IfIndex),
				NAPIs:  map[uint]*rxtxlayout.NAPI{},
				IRQs:   map[uint]*rxtxlayout.IRQ{},
			}
			ndevsmap[int(napi.IfIndex)] = ndev
		}
		var irq *rxtxlayout.IRQ
		if napi.HasIRQ {
			var ok bool
			irq, ok = ndev.IRQs[uint(napi.IRQ)]
			if !ok {
				var irqkthreadpid model.PIDType
				irqkthread := irqkthreads[uint(napi.IRQ)]
				if irqkthread != nil {
					irqkthreadpid = irqkthread.PID
				}
				irq = &rxtxlayout.IRQ{
					Source:  rxtxlayout.SourceNetlinkNetdev,
					ID:      uint(napi.IRQ),
					PID:     irqkthreadpid,
					Kthread: irqkthread,
				}
			}
		}
		ndev.NAPIs[uint(napi.ID)] = &rxtxlayout.NAPI{
			ID:      uint(napi.ID),
			Index:   int(napi.IfIndex),
			PID:     model.PIDType(napi.PID),
			Kthread: procs[model.PIDType(napi.PID)],
			IRQ:     irq,
		}
	}
	// Now create queue objects, with proper NAPI and IRQ references – to the
	// extend the drivers fill in the necessary bits of information that the
	// NETLINK netdev API wants to send to us then.
	for _, queue := range allqueues {
		ndev, ok := ndevsmap[int(queue.IfIndex)]
		if !ok {
			ndev = &rxtxlayout.Netdev{
				Source: rxtxlayout.SourceNetlinkNetdev,
				Index:  int(queue.IfIndex),
				NAPIs:  map[uint]*rxtxlayout.NAPI{},
				IRQs:   map[uint]*rxtxlayout.IRQ{},
			}
			ndevsmap[int(queue.IfIndex)] = ndev
		}
		napi, ok := ndev.NAPIs[uint(queue.NapiID)]
		var irq *rxtxlayout.IRQ
		if ok {
			irq = napi.IRQ
		}
		ndev.Queues = append(ndev.Queues, &rxtxlayout.Queue{
			ID:   uint(queue.ID),
			Type: queue.Type,
			NAPI: napi,
			IRQ:  irq,
		})
	}
	return maps.Values(ndevsmap), nil
}
