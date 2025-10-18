// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/netdb"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/network"
)

// newRootCmd creates the root command with usage and version information, as
// well as the available CLI flags (including descriptions).
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "lsallnifs",
		Short:   "dumpns outputs discovered network namespaces with interfaces, containers, ...",
		Version: "foobar",
		Args:    cobra.NoArgs,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return clippy.BeforeCommand(cmd)
		},
		RunE: lsallnifs,
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

	clippy.AddFlags(rootCmd)

	return
}

var lc = map[network.SocketSimplifiedState]string{
	network.Unconnected: "?",
	network.Listening:   "👂",
	network.Connected:   "↔",
}

func lsallnifs(cmd *cobra.Command, _ []string) error {
	out := cmd.OutOrStdout()
	fmt.Fprint(out, "lsallnifs\n")

	showAll, _ := cmd.PersistentFlags().GetBool("all")
	showTenants, _ := cmd.PersistentFlags().GetBool("tenants")
	showPorts, _ := cmd.PersistentFlags().GetBool("ports")
	showAddrs, _ := cmd.PersistentFlags().GetBool("addresses")
	//showRoutes, _ := cmd.PersistentFlags().GetBool("routes")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()

	fmt.Fprint(out, "discovering network namespaces and containers...\n")
	allnetns := gostwire.Discover(context.Background(), cizer, nil)

	for _, netns := range allnetns.Netns.Sorted() {
		fmt.Fprintf(out, "net:[%d] with %s:\n", netns.ID().Ino, netns.DisplayName())

		// Section "Tenants"
		if showAll || showTenants {
			fmt.Fprint(out, "  tenants:\n")
			tenants := netns.Tenants[:]
			tenants.Sort()
			for _, tenant := range tenants {
				// skip kthreadd which always shows up as an independent leader process
				if tenant.Process.PID == 2 {
					continue
				}
				fmt.Fprintf(out, "    %s\n", tenant.Name())
				fmt.Fprintf(out, "      /etc/hostname: '%s', UTS hostname: '%s', /etc/domainname: '%s'\n",
					tenant.DNS.EtcHostname, tenant.DNS.Hostname, tenant.DNS.EtcDomainname)
				fmt.Fprintf(out, "      search list: %s\n", strings.Join(tenant.DNS.Searchlist, ", "))
				addrs := []string{}
				for _, addr := range tenant.DNS.Nameservers {
					addrs = append(addrs, addr.String())
				}
				fmt.Fprintf(out, "      name servers: %s\n", strings.Join(addrs, ", "))
				fmt.Fprintf(out, "      hosts:\n")
				for name, ip := range tenant.DNS.Hosts {
					fmt.Fprintf(out, "        %s %s\n", name, ip.String())
				}
			}
		}

		// Section "Transports"
		if showAll || showPorts {
			fmt.Fprint(out, "  transports:\n")
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
					fmt.Fprintf(out, "    %s %s%s %s:%d%s %s:%d%s ↷ %s\n",
						lc[port.SimplifiedState], port.Protocol.String(), viasock6,
						network.IP(port.LocalIP).String(), port.LocalPort, serviceList(localservice),
						network.IP(port.RemoteIP).String(), port.RemotePort, serviceList(remoteservice),
						strings.Join(nifnames, ", "))
				}
			}
			listPorts(append(netns.Portsv4[:], netns.Portsv6...))
		}

		// Section "Network Interfaces"
		fmt.Fprint(out, "  network interfaces:\n")
		allnifs := netns.NifList()
		allnifs.Sort()
		for _, netif := range allnifs {
			nif := netif.Nif()
			alias := ""
			if nif.Alias != "" {
				alias = fmt.Sprintf(" ~'%s'", nif.Alias)
			}
			fmt.Fprintf(out, "    %s %s(%d)%s: kind %s, address %s\n",
				nif.State.TerminalIcon(), nif.Name, nif.Index, alias, nif.Kind, nif.L2Addr.String())

			// Addresses, addresses, addresses...
			if showAll || showAddrs {
				nifaddrs := append(nif.Addrsv4, nif.Addrsv6...)
				nifaddrs.Sort()
				for _, addr := range nifaddrs {
					fmt.Fprintf(out, "        %s/%d\n", addr.Address.String(), addr.PrefixLength)
				}
			}

			// Is this a bridge port? Then show its bridge...
			if nif.Bridge != nil {
				bridge := nif.Bridge.(network.Bridge).Bridge()
				fmt.Fprintf(out, "        ⌒ %s(%d)\n",
					bridge.Name, bridge.Index)
			}
			// Is this a MACVLAN master? Then list its MACVLANs...
			if macvlans := nif.Slaves.OfKind("macvlan"); len(macvlans) != 0 {
				for _, macvlan := range macvlans {
					macvlan := macvlan.Nif()
					fmt.Fprintf(out, "       ↳ MACVLAN: %s(%d) in %s\n",
						macvlan.Name, macvlan.Index, macvlan.Netns.DisplayName())
				}
			}
			// Has it VXLAN overlays? Then list its VXLANs...
			if vxlans := nif.Slaves.OfKind("vxlan"); len(vxlans) != 0 {
				for _, vxlan := range vxlans {
					vxlan := vxlan.(network.Vxlan).Vxlan()
					fmt.Fprintf(out, "       ↳ VXLAN overlay ID %d: %s(%d) in %s\n",
						vxlan.VID, vxlan.Name, vxlan.Index, vxlan.Netns.DisplayName())
				}
			}
			// Is this a bridge? Then list its ports...
			if bridge, ok := netif.(network.Bridge); ok {
				bridge := bridge.Bridge()
				for _, port := range bridge.Ports {
					port := port.Nif()
					fmt.Fprintf(out, "        ◌ port: %s(%d)\n",
						port.Name, port.Index)
				}
			}
			// Is this a MACVLAN? Then show its master...
			if macvlan, ok := netif.(network.Macvlan); ok {
				macvlan := macvlan.Macvlan()
				master := macvlan.Master.Nif()
				fmt.Fprintf(out, "      %s mode\n", macvlan.Mode.String())
				fmt.Fprintf(out, "       ☝  master %s(%d) in %s\n",
					master.Name, master.Index, master.Netns.DisplayName())
			}
			// Is this a VETH? Then show its peer...
			if veth, ok := netif.(network.Veth); ok {
				veth := veth.Veth()
				peer := veth.Peer.(network.Veth).Veth()
				fmt.Fprintf(out, "        ↔ %s(%d) in %s\n",
					peer.Name, peer.Index, peer.Netns.DisplayName())
			}
			// Is this a VXLAN? Then show its underlay master...
			if vxlan, ok := netif.(network.Vxlan); ok {
				vxlan := vxlan.Vxlan()
				fmt.Fprintf(out, "      VID %d, dest port %d\n", vxlan.VID, vxlan.DestinationPort)
				master := vxlan.Master.Nif()
				fmt.Fprintf(out, "       👇  underlay %s(%d) in %s\n",
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
