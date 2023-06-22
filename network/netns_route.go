// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"fmt"
	"net"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Route is Gostwire's view on network stack routes.
type Route struct {
	Family               AddressFamily
	Type                 RouteType
	Destination          net.IPNet
	DestinationPrefixLen int
	NextHop              net.IP
	Index                int
	Nif                  Interface
	Table                int
	Priority             int
	Preference           uint8 // TODO: support from vishvananda/netlink missing
}

// RouteType represents the type of route and allows converting it to a string,
// such as "unicast", "local", et cetera; this mimics pyroute2's behavior.
type RouteType uint8

var routeTypesMap = map[RouteType]string{
	unix.RTN_UNSPEC:      "unspec",
	unix.RTN_UNICAST:     "unicast",
	unix.RTN_LOCAL:       "local",
	unix.RTN_BROADCAST:   "broadcast",
	unix.RTN_ANYCAST:     "anycast",
	unix.RTN_MULTICAST:   "multicast",
	unix.RTN_BLACKHOLE:   "blackhole",
	unix.RTN_UNREACHABLE: "unreachable",
	unix.RTN_PROHIBIT:    "prohibit",
	unix.RTN_THROW:       "throw",
	unix.RTN_NAT:         "nat",
	unix.RTN_XRESOLVE:    "xresolve",
}

// String returns the string representation of a route type value in form of a
// lower-case identifier derived from the RTN_* constant identifiers, with the
// RTN_ prefix removed. For instance, "unicast", "local", et cetera.
func (r RouteType) String() string {
	if s, ok := routeTypesMap[r]; ok {
		return s
	}
	return fmt.Sprintf("RouteType(%d)", r)
}

func (n *NetworkNamespace) discoverRoutes(nlh *netlink.Handle, family int) []Route {
	// Please note that RouteListFiltered filters its result to return only
	// RT_TABLE_MAIN as long as the filter mask is zero. However, internally,
	// RouteListFiltered always dumps all tables (cringe), so we opt into table
	// filtering in the filter mask, but then set the table to filter to to
	// "unspecified".
	nlroutes, err := nlh.RouteListFiltered(
		family,
		&netlink.Route{
			Table: unix.RT_TABLE_UNSPEC, // we want to get all routing tables
		},
		netlink.RT_FILTER_TABLE)
	if err != nil {
		return []Route{} // don't nil, so any marshaller will not try to do unwanted things.
	}
	routes := make([]Route, 0, len(nlroutes))
	for _, route := range nlroutes {
		// Only process main and local tables; the local table not least
		// contains the multicast routes.
		if route.Table != unix.RT_TABLE_MAIN && route.Table != unix.RT_TABLE_LOCAL {
			continue
		}

		var dst net.IPNet
		if route.Dst != nil {
			dst = *route.Dst
		} else if family == unix.AF_INET {
			dst.IP = net.IPv4zero
		} else {
			dst.IP = net.IPv6zero
		}
		prefixlen, _ := dst.Mask.Size()
		r := Route{
			Family:               AddressFamily(family),
			Type:                 RouteType(route.Type),
			Destination:          dst,
			DestinationPrefixLen: prefixlen,
			NextHop:              route.Gw,
			Index:                route.LinkIndex,         // zero for blackhole routes, etc.
			Nif:                  n.Nifs[route.LinkIndex], // also works for blackhole routes, etc., giving nil.
			Table:                route.Table,
			Priority:             route.Priority,
			// default to ICMPV6_ROUTER_PREF_MEDIUM for the moment; TODO: support from vishvananda/netlink
			Preference: 0,
		}
		routes = append(routes, r)
	}
	return routes
}
