// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"
	"syscall"

	"github.com/google/nftables"
	"github.com/siemens/ghostwire/v2/network/portfwd"
	_ "github.com/siemens/ghostwire/v2/network/portfwd/all" // activate all port forwarding detectors.
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/portfinder"
)

// ForwardedPort is Gostwire's view on forwarded ports (actually port ranges)
// inside a particular network namespace. The rewritten destinations the traffic
// should be forwarded to can be in other network namespaces, as long as packets
// can be properly forwarded to these new destinations.
//
// Please note that under less-than-ideal circumstances we might end up with
// traffic to a host port getting forwarded into a particular network namespace
// (which we identify based on the network interfaces' configured IP addresses),
// yet there isn't any open socket handling the forwarded traffic. In this
// situation, only DestinationNetns will be set and PIDs/Processes are empty.
type ForwardedPort struct {
	portfinder.ForwardedPortRange                   // "general" port forwarding information incl. protocol, et cetera.
	Protocol                      Protocol          // transport-layer protocol, such as syscall.IPPROTO_TCP, etc.
	DestinationNetns              *NetworkNamespace // network namespace where the IP forwarded to is.
	PIDs                          []model.PIDType   // processes having a listening/open socket for the port forwarded to.
	Processes                     []*model.Process  // processes having a listening/open socket for the port forwarded to.
	Nifs                          Interfaces        // network interfaces the traffic is forwarded to.
}

// discoverForwardedPorts discovers forwarded transport layer ports (tcp, udp)
// from the IPv4 and IPv6 netfilter NAT tables. Thanks to the compatibility
// layer, this also includes the corresponding iptables as managed by Docker,
// CNI plugins, and a myriad of other existing software that really doesn't
// bother itself with netfilter.
func (n *NetworkNamespace) discoverForwardedPorts() {
	// If netfilters doesn't come to us, we simply come to netfilters ;) The
	// reson is that nftables uses @mdlayher/netlink and that really is a major
	// P.I.T.A. with its network namespace "support". Luckily, we have
	// OpenInNetworkNamespace that properly supports the more complex use cases
	// of namespace references so we simply wrap creating a netfilter netlink
	// socket in it.
	var conn *nftables.Conn
	err := n.OpenInNetworkNamespace(func() error {
		var err error
		conn, err = nftables.New(nftables.AsLasting())
		return err
	})
	if err != nil {
		if conn != nil {
			conn.CloseLasting()
		}
		log.Errorf("cannot connect to netfilters, reason: %s", err.Error())
		return
	}
	defer conn.CloseLasting()
	n.ForwardedPortsv4 = n.discoverForwardedPortsOfFamily(conn, nufftables.TableFamilyIPv4)
	n.ForwardedPortsv6 = n.discoverForwardedPortsOfFamily(conn, nufftables.TableFamilyIPv6)
}

// discoverForwardedPortsOfFamily discovers and returns forwarded transport
// layer ports for a single specific table/address family; either IPv4 or IPv6.
// We don't care about Inet ... unless some container engine or CNI plugins
// start to actually use it.
func (n *NetworkNamespace) discoverForwardedPortsOfFamily(conn *nftables.Conn, family nufftables.TableFamily) []ForwardedPort {
	iptables, err := nufftables.GetFamilyTables(conn, family)
	if err != nil {
		log.Errorf("cannot retrieve %s netfilter tables, reason: %s",
			family, err.Error())
		return nil
	}
	forwardedPorts := []ForwardedPort{}
	for _, portForwardings := range plugger.Group[portfwd.PortForwardings]().Symbols() {
		fwdports := portForwardings(iptables, family)
		for _, fwdp := range fwdports {
			log.Debugf("discovered %s", fwdp)
			var proto Protocol
			switch fwdp.Protocol {
			case "tcp":
				proto = syscall.IPPROTO_TCP
			case "udp":
				proto = syscall.IPPROTO_UDP
			case "sctp":
				proto = syscall.IPPROTO_SCTP
			}
			forwardedPorts = append(forwardedPorts, ForwardedPort{
				ForwardedPortRange: *fwdp,
				Protocol:           proto,
			})
		}
	}
	return forwardedPorts
}

// Relate the destinations ports are forwarded to to their network
// namespaces and matching sockets, if any.
func completeForwardedPortInformation(netspaces NetworkNamespaces) {
	log.Debugf("resolving forwarded port destinations...")
	for _, netns := range netspaces {
		for idx := range netns.ForwardedPortsv4 {
			ResolveForwardedPort(&netns.ForwardedPortsv4[idx], netns)
		}
		for idx := range netns.ForwardedPortsv6 {
			ResolveForwardedPort(&netns.ForwardedPortsv6[idx], netns)
		}
	}
}

// ResolveForwardedPort attempts to fill in the missing pieces of information
// for forwarded ports:
//   - network namespace of (rewritten) destination IP,
//   - network interface related to the (rewritten) destination IP,
//   - the processes (and PIDs) with socket(s) willing to serve the forwarded
//     port.
func ResolveForwardedPort(forwardedPort *ForwardedPort, netns *NetworkNamespace) {
	destNetns, nif := netns.WhereIs(forwardedPort.ForwardIP)
	if destNetns == nil {
		return
	}
	netns = destNetns
	forwardedPort.DestinationNetns = netns
	forwardedPort.Nifs = Interfaces{nif}
	// Find the matching socket(s) that are willing to handle the forwarded
	// traffic.
	if len(forwardedPort.IP) == net.IPv6len {
		forwardedPort.PIDs, forwardedPort.Processes = findMatchingProcesses(
			forwardedPort.ForwardPortMin, forwardedPort.ForwardIP, forwardedPort.Protocol, netns.Portsv6)
	} else {
		forwardedPort.PIDs, forwardedPort.Processes = findMatchingProcesses(
			forwardedPort.ForwardPortMin, forwardedPort.ForwardIP, forwardedPort.Protocol, netns.Portsv4)
	}
}

// findMatchingProcesses tries to find matching sockets that are able to receive
// the traffic that gets forwarded to a port in the network namespace they're
// in. It then returns the list of processes/PIDs that are serving the matching
// socket(s).
func findMatchingProcesses(port uint16, addr net.IP, proto Protocol, sockets []ProcessSocket) (pids []model.PIDType, processes []*model.Process) {
	for _, socket := range sockets {
		if socket.Protocol != proto || socket.LocalPort != port {
			continue
		}
		// Don't take sockets into consideration that are connected TCP sockets,
		// as we're only interested in a listening TCP socket.
		if socket.Protocol == syscall.IPPROTO_TCP && socket.SimplifiedState != Listening {
			continue
		}
		// If the socket is bound to a specific local address: does it match the
		// address the traffic gets forwarded to?
		if !socket.LocalIP.IsUnspecified() && !socket.LocalIP.Equal(addr) {
			continue
		}
		pids = append(pids, socket.PIDs...)
		processes = append(processes, socket.Processes...)
	}
	return
}

// WhereIs determines the network namespace the specified IP address is located
// in and at the same time reachable from the current network namespace. It
// returns the matching network namespace and network interface, or nil if nothing suitable was found.
func (n *NetworkNamespace) WhereIs(destIP net.IP) (*NetworkNamespace, Interface) {
	// First, let's see if this is an IP address in our "home" network
	// namespace, because then we've found the destination network namespace.
	if n.NifWithAddress(destIP) != nil {
		return n, n.NifWithAddress(destIP)
	}
	// Special case: IPv4 loopback interface and loopback network ... this is
	// not included as direct subnet routes in the routes table, so we have to
	// check explicitly for this case ... not least as Docker uses port
	// forwarding on loopback for its embedded DNS recursive resolver.
	if lo := n.NamedNifs["lo"]; lo != nil {
		for _, addr := range lo.Nif().Addrsv4 {
			subnet := net.IPNet{
				IP:   addr.Address,
				Mask: net.CIDRMask(int(addr.PrefixLength), 32),
			}
			if subnet.Contains(destIP) {
				return n, lo
			}
		}
	}
	// Next, consult the routing tables to decide where a packet would leave our
	// current network namespace...
	var routes []Route
	if len(destIP) == net.IPv6len {
		routes = n.Routesv6
	} else {
		routes = n.Routesv4
	}
	bestRoute := Route{
		DestinationPrefixLen: -1,
	}
	for _, route := range routes {
		// TODO: route priority? (upstream issue)
		if route.Destination.Contains(destIP) && route.DestinationPrefixLen > bestRoute.DestinationPrefixLen {
			bestRoute = route
		}
	}
	// If we didn't find any route, call it a day; this also relies on proper
	// direct subnet routes being present for unified handling.
	if bestRoute.DestinationPrefixLen < 0 {
		return nil, nil
	}
	// ...otherwise since we didn't have a direct destination hit on one of our
	// network interfaces, now see through the network interface we are leaving
	// this network namespace and where this will lead us to?
	ip := destIP
	if bestRoute.NextHop != nil && !bestRoute.NextHop.IsUnspecified() {
		ip = bestRoute.NextHop
	}
	switch bestRoute.Nif.Nif().Kind {
	case "bridge":
		// The next hop (if any) or the ultimate destination should be one of
		// the peer VETH network interfaces with their other end connected to
		// the "outgoing" bridge. Please note that we're NOT covering multiple
		// chained bridges.
		nif := n.NifInBridgedNetwork(bestRoute.Nif.Nif(), ip)
		if nif == nil {
			return nil, nil
		}
		return nif.Nif().Netns.WhereIs(destIP)
	case "veth":
		// It's a directly connected VETH, for what that is worth. The other
		// VETH end must be either the next hop or the ultimate destination,
		// otherwise we know it's a complete and utter miss.
		veth, ok := bestRoute.Nif.(Veth)
		if !ok {
			return nil, nil
		}
		peer := veth.Veth().Peer
		if peer == nil {
			return nil, nil
		}
		if !peer.Nif().HasAddress(ip) {
			return nil, nil
		}
		return peer.Nif().Netns.WhereIs(destIP)
	}
	return nil, nil
}

// NifInBridgeNetwork returns the network Interface with the specified IP
// address connected somehow to the specified bridge; otherwise, nil.
func (n *NetworkNamespace) NifInBridgedNetwork(bridge Interface, addr net.IP) Interface {
	br, ok := bridge.Interface().(Bridge) // grmpf ... first get the proper Interface
	if !ok {
		return nil
	}
	for _, port := range br.Bridge().Ports {
		veth, ok := port.Interface().(Veth)
		if !ok {
			continue
		}
		peer := veth.Veth().Peer
		if peer == nil {
			continue
		}
		if peer.Nif().HasAddress(addr) {
			return peer
		}
	}
	return nil
}

// NifWithAddress returns the network Interface in this network namespace that
// has the specified IP address assigned to it, or nil if none could be found.
func (n *NetworkNamespace) NifWithAddress(addr net.IP) Interface {
	for _, nif := range n.Nifs {
		if nif.Nif().HasAddress(addr) {
			return nif
		}
	}
	return nil
}
