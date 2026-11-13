// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"
	"syscall"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/nufftables/portfinder"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("resolving forwarded port destinations", func() {

	// a fake topology, with "host" connected to container "c1" via a VETH pair,
	// and to container "c2" via a bridge.
	var netspaces map[string]*NetworkNamespace
	var nifs map[string]Interface

	ip := func(s string) net.IP {
		ip := net.ParseIP(s)
		if ip4 := ip.To4(); ip4 != nil {
			return ip4
		}
		return ip
	}

	// (re)builds the fake topology for each spec: the network namespaces and
	// their addressed interfaces, the VETH pairs vethH↔vethC1 and vethB↔vethC2,
	// bridge br0 with ports vethY and vethB, and the host's IPv4 routes; vethX
	// lacks a peer and 10.1.1.99 is a next hop not present at c1.
	BeforeEach(func() {
		netspaces = map[string]*NetworkNamespace{}
		nifs = map[string]Interface{}
		for _, name := range []string{"host", "c1", "c2"} {
			netspaces[name] = &NetworkNamespace{Nifs: map[int]Interface{}, NamedNifs: map[string]Interface{}}
		}
		add := func(netnsname string, index int, nif Interface, name, kind string, cidrs ...string) {
			netns := netspaces[netnsname]
			attrs := nif.Nif()
			attrs.Index, attrs.Name, attrs.Kind, attrs.Netns = index, name, kind, netns
			for _, cidr := range cidrs {
				addr, subnet, err := net.ParseCIDR(cidr)
				Expect(err).NotTo(HaveOccurred())
				prefixlen, _ := subnet.Mask.Size()
				if addr4 := addr.To4(); addr4 != nil {
					attrs.Addrsv4 = append(attrs.Addrsv4, Address{Address: addr4, PrefixLength: uint(prefixlen)})
				} else {
					attrs.Addrsv6 = append(attrs.Addrsv6, Address{Address: addr, PrefixLength: uint(prefixlen)})
				}
			}
			netns.Nifs[index] = nif
			netns.NamedNifs[name] = nif
			nifs[name] = nif
		}
		add("host", 1, &NifAttrs{}, "lo", "", "127.0.0.1/8")
		add("host", 2, &NifAttrs{}, "dummy0", "", "192.168.9.1/24", "2001:db8::1/64")
		add("host", 3, &VethAttrs{}, "vethH", "veth", "10.1.1.1/24")
		add("host", 4, &VethAttrs{}, "vethX", "veth")
		add("host", 5, &BridgeAttrs{}, "br0", "bridge", "172.17.0.1/16")
		add("host", 6, &VethAttrs{}, "vethB", "veth")
		add("host", 7, &VethAttrs{}, "vethY", "veth")
		add("c1", 1, &VethAttrs{}, "vethC1", "veth", "10.1.1.2/24")
		add("c1", 2, &NifAttrs{}, "app0", "", "10.2.2.2/32")
		add("c2", 1, &VethAttrs{}, "vethC2", "veth", "172.17.0.2/16")

		pair := func(a, b string) {
			nifs[a].(*VethAttrs).Peer = nifs[b]
			nifs[b].(*VethAttrs).Peer = nifs[a]
		}
		pair("vethH", "vethC1")
		pair("vethB", "vethC2")
		nifs["br0"].(*BridgeAttrs).Ports = Interfaces{nifs["vethY"], nifs["vethB"]}

		route := func(cidr, nexthop, nifname string) Route {
			_, dest, err := net.ParseCIDR(cidr)
			Expect(err).NotTo(HaveOccurred())
			prefixlen, _ := dest.Mask.Size()
			r := Route{Destination: *dest, DestinationPrefixLen: prefixlen, Nif: nifs[nifname]}
			if nexthop != "" {
				r.NextHop = ip(nexthop)
			}
			return r
		}
		netspaces["host"].Routesv4 = []Route{
			route("10.1.1.0/24", "", "vethH"),
			route("10.2.0.0/16", "10.1.1.2", "vethH"),
			route("10.77.0.0/16", "10.1.1.99", "vethH"),
			route("10.66.0.0/16", "", "vethX"),
			route("172.17.0.0/16", "", "br0"),
			route("192.168.9.0/24", "", "dummy0"),
		}
	})

	DescribeTable("locating IP addresses",
		func(dest string, netnsname string, nifname string) {
			netns, nif := netspaces["host"].WhereIs(ip(dest))
			if netnsname == "" {
				Expect(netns).To(BeNil())
				Expect(nif).To(BeNil())
				return
			}
			Expect(netns).To(BeIdenticalTo(netspaces[netnsname]))
			Expect(nif).To(BeIdenticalTo(nifs[nifname]))
		},
		Entry("own address", "192.168.9.1", "host", "dummy0"),
		Entry("loopback network", "127.0.0.11", "host", "lo"),
		Entry("directly connected VETH peer", "10.1.1.2", "c1", "vethC1"),
		Entry("via next hop", "10.2.2.2", "c1", "app0"),
		Entry("bridged", "172.17.0.2", "c2", "vethC2"),
		Entry("bridged, but not connected", "172.17.0.99", "", ""),
		Entry("VETH without peer", "10.66.0.1", "", ""),
		Entry("next hop not at peer", "10.77.0.1", "", ""),
		Entry("unsupported outgoing network interface", "192.168.9.5", "", ""),
		Entry("no route", "8.8.8.8", "", ""),
	)

	It("doesn't look for bridged network interfaces of non-bridges", func() {
		Expect(netspaces["host"].NifInBridgedNetwork(nifs["vethH"], ip("10.1.1.2"))).To(BeNil())
	})

	It("doesn't resolve unreachable destinations", func() {
		fp := &ForwardedPort{ForwardedPortRange: portfinder.ForwardedPortRange{ForwardIP: ip("8.8.8.8")}}
		ResolveForwardedPort(fp, netspaces["host"])
		Expect(fp.DestinationNetns).To(BeNil())
	})

	It("resolves IPv4 destinations and their listening processes", func() {
		proc := &model.Process{PID: 42}
		netspaces["c1"].Portsv4 = []ProcessSocket{
			{Protocol: syscall.IPPROTO_UDP, LocalPort: 80, LocalIP: ip("0.0.0.0"), PIDs: []model.PIDType{1}},
			{Protocol: syscall.IPPROTO_TCP, LocalPort: 80, LocalIP: ip("0.0.0.0"), SimplifiedState: Connected, PIDs: []model.PIDType{2}},
			{Protocol: syscall.IPPROTO_TCP, LocalPort: 80, LocalIP: ip("10.9.9.9"), SimplifiedState: Listening, PIDs: []model.PIDType{3}},
			{Protocol: syscall.IPPROTO_TCP, LocalPort: 80, LocalIP: ip("0.0.0.0"), SimplifiedState: Listening,
				PIDs: []model.PIDType{42}, Processes: []*model.Process{proc}},
		}
		fp := &ForwardedPort{
			ForwardedPortRange: portfinder.ForwardedPortRange{
				IP: ip("192.168.9.1"), ForwardIP: ip("10.1.1.2"), ForwardPortMin: 80,
			},
			Protocol: syscall.IPPROTO_TCP,
		}
		ResolveForwardedPort(fp, netspaces["host"])
		Expect(fp.DestinationNetns).To(BeIdenticalTo(netspaces["c1"]))
		Expect(fp.Nifs).To(ConsistOf(BeIdenticalTo(nifs["vethC1"])))
		Expect(fp.PIDs).To(ConsistOf(model.PIDType(42)))
		Expect(fp.Processes).To(ConsistOf(BeIdenticalTo(proc)))
	})

	It("resolves IPv6 destinations and their listening processes", func() {
		netspaces["host"].Portsv6 = []ProcessSocket{
			{Protocol: syscall.IPPROTO_UDP, LocalPort: 53, LocalIP: ip("::"), PIDs: []model.PIDType{666}},
		}
		fp := &ForwardedPort{
			ForwardedPortRange: portfinder.ForwardedPortRange{
				IP: ip("2001:db8::42"), ForwardIP: ip("2001:db8::1"), ForwardPortMin: 53,
			},
			Protocol: syscall.IPPROTO_UDP,
		}
		ResolveForwardedPort(fp, netspaces["host"])
		Expect(fp.DestinationNetns).To(BeIdenticalTo(netspaces["host"]))
		Expect(fp.PIDs).To(ConsistOf(model.PIDType(666)))
	})

})
