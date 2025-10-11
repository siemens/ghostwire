// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"github.com/siemens/ghostwire/v2/innetns"
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/siemens/ghostwire/v2/passedthrough"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Discover returns netdev configuration information about the RX/TX queue
// structure, et cetera, for the network namespaces passed in a namespaces
// discovery result.
func Discover(allnetns *discover.Result) (rxtxlayout.NetdevsByNetns, error) {
	irqkthreads := rxtxlayout.NewIRQKthreadsMap(allnetns.Processes)

	m := rxtxlayout.NetdevsByNetns{}
	for _, netns := range allnetns.Namespaces[model.NetNS] {
		ndevs := discoverInNetns(netns, irqkthreads, allnetns.Processes)
		m[netns] = ndevs
	}
	return m, nil
}

func discoverInNetns(
	netns model.Namespace,
	irqkthreads rxtxlayout.IRQKthreadMap,
	procs model.ProcessTable,
) []*rxtxlayout.Netdev {
	// Ensure to clean up any NETLINK netdev API connection we might have
	// opened...
	var conn *Conn
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	// We need the naming information about the links, as this is unfortunately
	// not covered by the NETLINK netdev API.
	var links []netlink.Link
	if err := innetns.Run(netns, func() error {
		nlHandle, err := netlink.NewHandle(unix.NETLINK_ROUTE)
		if err != nil {
			return nil
		}
		defer nlHandle.Close()
		links, err = nlHandle.LinkList()
		if err != nil {
			return nil
		}

		conn, err = Dial(nil)
		return err
	}); err != nil {
		return nil
	}
	// Now query the netdev API for queue, IRQ, and NAPI information...
	netdevs, err := conn.Netdevs(irqkthreads, procs)
	if err != nil {
		return nil
	}
	// If we've come this far, we now build ifindex->link map and then fill in
	// the netdev names and original names into the netdev information we've got
	// in the previous step...
	linksByIndex := map[int]netlink.Link{}
	for _, link := range links {
		linksByIndex[link.Attrs().Index] = link
	}
	for _, netdev := range netdevs {
		link, ok := linksByIndex[netdev.Index]
		if !ok {
			continue
		}
		netdev.Name = link.Attrs().Name
		netdev.OriginalName = passedthrough.OriginalIfname(link)
	}

	return netdevs
}
