// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/innetns"
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/siemens/ghostwire/v2/passedthrough"
)

// Discover returns netdev configuration information about the RX/TX queue
// structure, et cetera, for the network namespaces passed in a namespaces
// discovery result.
func Discover(allnetns *discover.Result) (rxtxlayout.NetdevsByNetns, error) {
	irqkthreads := rxtxlayout.NewIRQKthreadsMap(allnetns.Processes)
	napikthreads := rxtxlayout.NewNAPIKthreadsMap(allnetns.Processes)

	m := rxtxlayout.NetdevsByNetns{}
	for _, netns := range allnetns.Namespaces[model.NetNS] {
		// Get the links/netdevs in this network namespace; we especially need
		// the alternate names in order to detect "romaing" netdevs managed by
		// passthrough drivers so we have to use RTNETLINK here.
		var links []netlink.Link
		if err := innetns.Run(netns, func() error {
			nlHandle, err := netlink.NewHandle(unix.NETLINK_ROUTE)
			if err != nil {
				return err
			}
			defer func() { _ = nlHandle.Close() }()
			links, err = nlHandle.LinkList()
			return err
		}); err != nil {
			continue
		}
		ndevs := DiscoverInNetnsWithLinks(netns, irqkthreads, napikthreads, links)
		m[netns] = ndevs
	}
	return m, nil
}

// DiscoverInNetnsWithLinks discovers the netdev configuration regarding RX/TX
// queue structure, et cetera, for a specific network namespace and with using
// the passed links information.
func DiscoverInNetnsWithLinks(
	netns model.Namespace,
	irqkthreads rxtxlayout.IRQKthreadMap,
	napikthreads rxtxlayout.NAPIKthreadMap,
	links []netlink.Link,
) []*rxtxlayout.Netdev {
	// first and foremost, we need a proper sysfs instance that shows us the
	// correct network namespace, because the /sys/class/net branch doesn't
	// dynamically change based on the viewer's network namespace, but instead
	// has been "frozen" when a particular sysfs instance was mounted. If we
	// cannot find one, we bail out and return an empty list.
	sysfsMounty := SysfsOfNetns(netns)
	if sysfsMounty == nil {
		return nil
	}
	defer sysfsMounty.Close()
	sysfspath, err := sysfsMounty.Resolve("/")
	if err != nil {
		return nil
	}
	return discoverLayouts(links, sysfspath, irqkthreads, napikthreads)
}

// discoverLayouts takes a set of netlink.Link descriptions and then tries to
// figure out the RX/TX queue, IRQ and NAPI layout, including related kernel
// threads.
//
// Separating this part allows for better unit testing with especially the
// otherwise difficult to test roaming netdevs where they originally had a
// different name and thus the current ifname would either not match any IRQ and
// NAPI kthreads at all, or worse, the wrong ones.
func discoverLayouts(
	links []netlink.Link,
	sysfspath string,
	irqkthreads rxtxlayout.IRQKthreadMap,
	napikthreads rxtxlayout.NAPIKthreadMap,
) []*rxtxlayout.Netdev {
	ndevs := make([]*rxtxlayout.Netdev, 0, len(links))
	for _, link := range links {
		ndevs = append(ndevs, &rxtxlayout.Netdev{
			Source:       rxtxlayout.SourceSysfs,
			Name:         link.Attrs().Name,
			Index:        link.Attrs().Index,
			OriginalName: passedthrough.OriginalIfname(link),
		})
	}
	for _, ndev := range ndevs {
		if err := NetdevQueues(sysfspath, ndev); err != nil {
			continue
		}
		NetdevIRQs(sysfspath, ndev) // fire-and-forget
	}
	rxtxlayout.FillInIRQKthreads(irqkthreads, ndevs)
	rxtxlayout.FillInNAPIKthreads(napikthreads, ndevs)
	return ndevs
}
