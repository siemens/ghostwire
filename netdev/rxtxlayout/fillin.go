// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"maps"
	"slices"
)

func Fillin(netdevmapA, netdevmapB NetdevsByNetns) NetdevsByNetns {
	ndevs := maps.Clone(netdevmapA)
	// Fill in netdevs that are in B, but not in A: such as netdevs that are in
	// operstate DOWN and thus only discovered via sysfs, but not NETLINK netdev.
	for netns, ndevsB := range netdevmapB {
		for _, ndevB := range ndevsB {
			if slices.ContainsFunc(netdevmapA[netns],
				func(ndevA *Netdev) bool { return ndevA.Name == ndevB.Name }) {
				continue
			}
			ndevs[netns] = append(ndevs[netns], ndevB)
		}
	}
	// Fill in netdev IRQ information that is in B, but not in A
	for netns, ndevsB := range netdevmapB {
		for _, ndevB := range ndevsB {
			ndIdx := slices.IndexFunc(ndevs[netns],
				func(ndev *Netdev) bool { return ndev.Name == ndevB.Name })
			if ndIdx < 0 {
				continue
			}
			for _, irqB := range ndevB.IRQs {
				ndev := ndevs[netns][ndIdx]
				if ndev.IRQs[irqB.ID] != nil {
					continue
				}
				ndev.IRQs[irqB.ID] = irqB
				// also link in to queue, if applicable; we have the IRQ, so
				// which queues (of netdevB) link to it?
				for _, queueB := range ndevB.Queues {
					if queueB.IRQ != irqB {
						continue
					}
					qIdx := slices.IndexFunc(ndev.Queues,
						func(q *Queue) bool {
							return q.ID == queueB.ID && q.Type == queueB.Type
						})
					if qIdx < 0 {
						continue
					}
					ndev.Queues[qIdx].IRQ = irqB
				}
			}
		}
	}
	return ndevs
}
