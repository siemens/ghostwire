// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"strings"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/turtlefinder"

	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/netdb"
)

// newRootCmd creates the root command with usage and version information, as
// well as the available CLI flags (including descriptions).
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lsallnifs",
		Short:   "dumpns outputs discovered network namespaces with interfaces, containers, ...",
		Version: "foobar",
		Args:    cobra.NoArgs,
		RunE:    lsallnifs,
	}
	// Sets up the flags.
	rootCmd.PersistentFlags().BoolP(
		"debug", "d", false,
		"show debug output")

	rootCmd.PersistentFlags().BoolP(
		"all", "x", false,
		"everything, but the kitchen sink")
	rootCmd.PersistentFlags().BoolP(
		"tenants", "t", false,
		"show tenants")
	rootCmd.PersistentFlags().BoolP(
		"ports", "p", false,
		"show open ports")
	rootCmd.PersistentFlags().BoolP(
		"addresses", "a", false,
		"show network interface addresses")
	rootCmd.PersistentFlags().BoolP(
		"routes", "r", false,
		"show routes")

	return
}

var lc = map[network.SocketSimplifiedState]string{
	network.Unconnected: "?",
	network.Listening:   "👂",
	network.Connected:   "↔",
}

func lsallnifs(cmd *cobra.Command, _ []string) error {
	log.Infof("Gostwire (The Sequel)")

	if debug, _ := cmd.PersistentFlags().GetBool("debug"); debug {
		log.SetLevel(log.DebugLevel)
		log.Debugf("debug logging enabled")
	}
	showAll, _ := cmd.PersistentFlags().GetBool("all")
	showTenants, _ := cmd.PersistentFlags().GetBool("tenants")
	showPorts, _ := cmd.PersistentFlags().GetBool("ports")
	showAddrs, _ := cmd.PersistentFlags().GetBool("addresses")
	//showRoutes, _ := cmd.PersistentFlags().GetBool("routes")

	log.Debugf("using TurtleFinder")
	enginectx, enginecancel := context.WithCancel(context.Background())
	cizer := turtlefinder.New(func() context.Context { return enginectx })
	defer enginecancel()
	defer cizer.Close()

	log.Debugf("discovering network namespaces and containers...")
	allnetns := gostwire.Discover(context.Background(), cizer, nil)

	for _, netns := range allnetns.Netns.Sorted() {
		log.Infof("net:[%d] with %s:\n", netns.ID().Ino, netns.DisplayName())

		// Section "Tenants"
		if showAll || showTenants {
			log.Infof("  tenants:")
			tenants := netns.Tenants[:]
			tenants.Sort()
			for _, tenant := range tenants {
				// skip kthreadd which always shows up as an independent leader process
				if tenant.Process.PID == 2 {
					continue
				}
				log.Infof("    %s", tenant.Name())
				log.Infof("      /etc/hostname: '%s', UTS hostname: '%s', /etc/domainname: '%s'",
					tenant.DNS.EtcHostname, tenant.DNS.Hostname, tenant.DNS.EtcDomainname)
				log.Infof("      search list: %s", strings.Join(tenant.DNS.Searchlist, ", "))
				addrs := []string{}
				for _, addr := range tenant.DNS.Nameservers {
					addrs = append(addrs, addr.String())
				}
				log.Infof("      name servers: %s", strings.Join(addrs, ", "))
				log.Infof("      hosts:")
				for name, ip := range tenant.DNS.Hosts {
					log.Infof("        %s %s", name, ip.String())
				}
			}
		}

		// Section "Transports"
		if showAll || showPorts {
			log.Infof("  transports:")
			listPorts := func(ports network.ProcessSockets) {
				ports.Sort()
				for _, port := range ports {
					nifnames := []string{}
					for _, nif := range port.Nifs {
						if len(nifnames) > 3 {
							nifnames = append(nifnames, "…")
							break
						}
						nifnames = append(nifnames, nif.Nif().Name)
					}
					viasock6 := ""
					if port.IPv4Mapped {
						viasock6 = " ⑥"
					}
					localservice := netdb.ServiceByPort(int(port.LocalPort), strings.ToLower(port.Protocol.String()))
					remoteservice := netdb.ServiceByPort(int(port.RemotePort), strings.ToLower(port.RemoteIP.String()))
					log.Infof("    %s %s%s %s:%d%s %s:%d%s ↷ %s",
						lc[port.SimplifiedState], port.Protocol.String(), viasock6,
						network.IP(port.LocalIP).String(), port.LocalPort, serviceList(localservice),
						network.IP(port.RemoteIP).String(), port.RemotePort, serviceList(remoteservice),
						strings.Join(nifnames, ", "))
				}
			}
			listPorts(append(netns.Portsv4[:], netns.Portsv6...))
		}

		// Section "Network Interfaces"
		log.Infof("  network interfaces:")
		allnifs := netns.NifList()
		allnifs.Sort()
		for _, netif := range allnifs {
			nif := netif.Nif()
			alias := ""
			if nif.Alias != "" {
				alias = fmt.Sprintf(" ~'%s'", nif.Alias)
			}
			log.Infof("    %s %s(%d)%s: kind %s, address %s\n",
				nif.State.TerminalIcon(), nif.Name, nif.Index, alias, nif.Kind, nif.L2Addr.String())

			// Addresses, addresses, addresses...
			if showAll || showAddrs {
				nifaddrs := append(nif.Addrsv4, nif.Addrsv6...)
				nifaddrs.Sort()
				for _, addr := range nifaddrs {
					log.Infof("        %s/%d", addr.Address.String(), addr.PrefixLength)
				}
			}

			// Is this a bridge port? Then show its bridge...
			if nif.Bridge != nil {
				bridge := nif.Bridge.(network.Bridge).Bridge()
				log.Infof("        ⌒ %s(%d)",
					bridge.Name, bridge.Index)
			}
			// Is this a MACVLAN master? Then list its MACVLANs...
			if macvlans := nif.Slaves.OfKind("macvlan"); len(macvlans) != 0 {
				for _, macvlan := range macvlans {
					macvlan := macvlan.Nif()
					log.Infof("       ↳ MACVLAN: %s(%d) in %s",
						macvlan.Name, macvlan.Index, macvlan.Netns.DisplayName())
				}
			}
			// Has it VXLAN overlays? Then list its VXLANs...
			if vxlans := nif.Slaves.OfKind("vxlan"); len(vxlans) != 0 {
				for _, vxlan := range vxlans {
					vxlan := vxlan.(network.Vxlan).Vxlan()
					log.Infof("       ↳ VXLAN overlay ID %d: %s(%d) in %s",
						vxlan.VID, vxlan.Name, vxlan.Index, vxlan.Netns.DisplayName())
				}
			}
			// Is this a bridge? Then list its ports...
			if bridge, ok := netif.(network.Bridge); ok {
				bridge := bridge.Bridge()
				for _, port := range bridge.Ports {
					port := port.Nif()
					log.Infof("        ◌ port: %s(%d)",
						port.Name, port.Index)
				}
			}
			// Is this a MACVLAN? Then show its master...
			if macvlan, ok := netif.(network.Macvlan); ok {
				macvlan := macvlan.Macvlan()
				master := macvlan.Master.Nif()
				log.Infof("      %s mode", macvlan.Mode.String())
				log.Infof("       ☝  master %s(%d) in %s",
					master.Name, master.Index, master.Netns.DisplayName())
			}
			// Is this a VETH? Then show its peer...
			if veth, ok := netif.(network.Veth); ok {
				veth := veth.Veth()
				peer := veth.Peer.(network.Veth).Veth()
				log.Infof("        ↔ %s(%d) in %s",
					peer.Name, peer.Index, peer.Netns.DisplayName())
			}
			// Is this a VXLAN? Then show its underlay master...
			if vxlan, ok := netif.(network.Vxlan); ok {
				vxlan := vxlan.Vxlan()
				log.Infof("      VID %d, dest port %d", vxlan.VID, vxlan.DestinationPort)
				master := vxlan.Master.Nif()
				log.Infof("       👇  underlay %s(%d) in %s",
					master.Name, master.Index, master.Netns.DisplayName())
			}
		}
	}
	return nil
}

func serviceList(s *netdb.Service) string {
	if s == nil {
		return ""
	}
	return ` "` + strings.Join(append([]string{s.Name}, s.Aliases...), ", ") + `"`
}
