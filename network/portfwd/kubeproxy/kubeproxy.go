// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package docker

import (
	"bytes"
	"net"
	"strings"

	"github.com/google/nftables/expr"
	"github.com/google/nftables/xt"
	"github.com/siemens/ghostwire/v2/network/portfwd"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/portfinder"
	"golang.org/x/sys/unix"
)

const (
	kubeServicesChain         = "KUBE-SERVICES"
	kubeServiceChainPrefix    = "KUBE-SVC-"
	kubeSeparationChainPrefix = "KUBE-SEP-"
)

// Register this PortForwardings plugin.
func init() {
	plugger.Group[portfwd.PortForwardings]().Register(
		PortForwardings, plugger.WithPlugin("kubeproxy"))
}

// PortForwardings discovers kube-proxy's forwarded (virtual service address)
// ports from the “nat” table(s) (only for IPv4 and IPv6 respectively).
func PortForwardings(tables nufftables.TableMap, family nufftables.TableFamily) []*portfinder.ForwardedPortRange {
	switch family {
	case nufftables.TableFamilyIPv4, nufftables.TableFamilyIPv6:
	default:
		return nil
	}
	nattable := tables.Table("nat", family)
	if nattable == nil {
		return nil
	}
	forwardedPorts := []*portfinder.ForwardedPortRange{}
	kubeServices := nattable.ChainsByName[kubeServicesChain]
	if kubeServices == nil {
		return nil
	}

	for _, svcRules := range kubeServices.Rules {
		ip, proto, port, comment, svcChainName := service(svcRules.Exprs)
		if svcChainName == "" {
			continue
		}
		_ = comment // TODO: upstream nufftables
		for _, sepChain := range separations(nattable.ChainsByName[svcChainName]) {
			sc := nattable.ChainsByName[sepChain]
			if sc == nil {
				continue
			}
			for _, rule := range sc.Rules {
				_, target := nufftables.OfTypeFunc(rule.Exprs, isDNAT)
				if target == nil {
					continue
				}
				nr := target.Info.(*xt.NatRange2)
				forwardedPorts = append(forwardedPorts, &portfinder.ForwardedPortRange{
					Protocol:       proto,
					IP:             ip,
					PortMin:        port,
					PortMax:        port,
					ForwardIP:      net.IP(nr.MinIP),
					ForwardPortMin: nr.MinPort,
				})
			}
		}
	}

	return forwardedPorts
}

func isDNAT(target *expr.Target) bool {
	if target.Name != "DNAT" {
		return false
	}
	_, ok := target.Info.(*xt.NatRange2)
	return ok
}

func separations(chain *nufftables.Chain) (chains []string) {
	if chain == nil {
		return
	}
	for _, rule := range chain.Rules {
		_, verdict := nufftables.OfTypeFunc(rule.Exprs, isSepVerdict)
		if verdict == nil {
			continue
		}
		chains = append(chains, verdict.Chain)
	}
	return
}

func isSepVerdict(verdict *expr.Verdict) bool {
	return verdict.Kind == -3 && strings.HasPrefix(verdict.Chain, kubeSeparationChainPrefix)
}

func service(
	exprs nufftables.Expressions,
) (
	ip net.IP, protocol string, port uint16, comment string, chain string,
) {
	exprs, svcproto := nufftables.OfTypeFunc(exprs, isIPProtoTcpUdp)
	exprs, svcip := nufftables.OfTypeFunc(exprs, isIP)
	exprs, svccomment := nufftables.OfTypeFunc(exprs, isComment)
	exprs, svcport := nufftables.OfTypeFunc(exprs, isServicePort)
	exprs, svcchain := nufftables.OfTypeFunc(exprs, isServiceChain)
	if exprs == nil {
		return nil, "", 0, "", ""
	}

	ip = net.IP(svcip.Data)
	switch svcproto.Data[0] {
	case unix.IPPROTO_TCP:
		protocol = "tcp"
	case unix.IPPROTO_UDP:
		protocol = "udp"
	}
	port = uint16(svcport.Data[0])<<8 + uint16(svcport.Data[1])
	comment = string(bytes.TrimRight([]byte(*svccomment.Info.(*xt.Unknown)), "\x00"))
	chain = svcchain.Chain
	return
}

func isIPProtoTcpUdp(cmp *expr.Cmp) bool {
	if len(cmp.Data) != 1 {
		return false
	}
	switch cmp.Data[0] {
	case unix.IPPROTO_TCP, unix.IPPROTO_UDP:
		return true
	}
	return false
}

func isIP(cmp *expr.Cmp) bool {
	switch len(cmp.Data) {
	case 4, 16:
		return true
	}
	return false
}

func isComment(match *expr.Match) bool {
	return match.Name == "comment"
}

func isServicePort(cmp *expr.Cmp) bool {
	return len(cmp.Data) == 2
}

func isServiceChain(verdict *expr.Verdict) bool {
	return verdict.Kind == -3 && strings.HasPrefix(verdict.Chain, kubeServiceChainPrefix)
}
