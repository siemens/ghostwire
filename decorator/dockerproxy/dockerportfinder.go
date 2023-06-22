// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockerproxy

import (
	"context"
	"net"
	"strconv"
	"strings"
	"syscall"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/engineclient/moby"
	"golang.org/x/exp/slices"
)

const dockerProxy = "docker-proxy"

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("dockerportfinder"))
}

// forwardedPortKey is the unique index for a forwarded port.
type forwardedPortKey struct {
	Protocol network.Protocol
	IP       string // in binary
	Port     uint16
}

// Decorate scans the found Docker engine processes for their child proxy
// processes, adding the found forwarded ports to the already known forwarded
// ports in the network namespaces of their corresponding Docker engines.
func Decorate(ctx context.Context, allnetns network.NetworkNamespaces, allprocs model.ProcessTable, engines []*model.ContainerEngine) {
	log.Debugf("discovering user-space Docker forwarded ports")

	for _, engine := range engines {
		// If it ain't a Docker engine, we can skip it.
		if engine.Type != moby.Type {
			continue
		}
		engineproc, ok := allprocs[engine.PID]
		if !ok {
			continue
		}
		netns, ok := allnetns[engineproc.Namespaces[model.NetNS].ID()]
		if !ok {
			continue
		}
		// Now see if any and which docker proxy processes are to be found...
		forwardedPorts := map[forwardedPortKey]network.ForwardedPort{}
		for _, child := range engineproc.Children {
			if !strings.HasSuffix(child.Name, dockerProxy) {
				continue
			}
			fp := forwardedPort(child)
			if fp.IP == nil {
				continue
			}
			network.ResolveForwardedPort(&fp, netns)
			forwardedPorts[forwardedPortKey{
				Protocol: fp.Protocol,
				IP:       string(fp.IP),
				Port:     fp.PortMin,
			}] = fp
		}
		if len(forwardedPorts) == 0 {
			continue
		}
		// finally update the list of discovered forwarded ports for the network
		// namespace of the particular Docker engine process.
		netns.ForwardedPortsv4 = mergeForwardedPorts(net.IPv4len, netns.ForwardedPortsv4, forwardedPorts)
		netns.ForwardedPortsv6 = mergeForwardedPorts(net.IPv6len, netns.ForwardedPortsv6, forwardedPorts)
	}
}

// mergeForwardedPorts returns the merge of the already known forwarded ports
// and the newly discovered user-land forwarded ports, restricted to the
// specified IP family.
func mergeForwardedPorts(IPvLen int, ports []network.ForwardedPort, moreports map[forwardedPortKey]network.ForwardedPort) []network.ForwardedPort {
	for _, port := range slices.Clone(ports) {
		delete(moreports, forwardedPortKey{
			Protocol: port.Protocol,
			IP:       string(port.IP),
			Port:     port.PortMin,
		})
	}
	for _, port := range moreports {
		if len(port.IP) != IPvLen {
			continue
		}
		log.Debugf("docker proxy: %s", port)
		ports = append(ports, port)
	}
	return ports
}

// forwardedPort decodes the port forwarding information passed to a Docker
// proxy process on its command line and returns it.
func forwardedPort(proxyproc *model.Process) (fp network.ForwardedPort) {
	cliarg := proxyproc.Cmdline
	for idx := 1; idx < len(cliarg)-1; idx += 2 {
		switch cliarg[idx] {
		case "-proto":
			proto := cliarg[idx+1]
			if proto == "" {
				return network.ForwardedPort{}
			}
			fp.ForwardedPortRange.Protocol = proto
			switch proto {
			case "tcp":
				fp.Protocol = syscall.IPPROTO_TCP
			case "udp":
				fp.Protocol = syscall.IPPROTO_UDP
			case "sctp":
				fp.Protocol = syscall.IPPROTO_SCTP
			}
		case "-host-ip":
			ip := parseIP(cliarg[idx+1])
			if ip == nil {
				return network.ForwardedPort{}
			}
			fp.IP = ip
		case "-host-port":
			port := parsePort(cliarg[idx+1])
			if port == 0 {
				return network.ForwardedPort{}
			}
			fp.PortMin = port
			fp.PortMax = port
		case "-container-ip":
			ip := parseIP(cliarg[idx+1])
			if ip == nil {
				return network.ForwardedPort{}
			}
			fp.ForwardIP = ip
		case "-container-port":
			port := parsePort(cliarg[idx+1])
			if port == 0 {
				return network.ForwardedPort{}
			}
			fp.ForwardPortMin = port
		}
	}
	return
}

// parseIP returns the (canonical) IPv4 or IPv6 binary address representation
// for the given textual IP address representation, or nil if the textual
// representation is invalid.
func parseIP(s string) net.IP {
	ip := net.ParseIP(s)
	if ip == nil {
		return nil
	}
	if ipv4 := ip.To4(); ipv4 != nil {
		return ipv4
	}
	return ip
}

// parsePort returns the port number for the specified textual port number (in
// decimal). It returns 0 for invalid textual representations.
func parsePort(s string) uint16 {
	port, _ := strconv.ParseUint(s, 10, 16)
	return uint16(port)
}
