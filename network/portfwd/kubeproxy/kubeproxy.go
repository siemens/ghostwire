// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package kubeproxy

import (
	"net"
	"strings"

	"github.com/google/nftables/expr"
	"github.com/google/nftables/xt"
	"github.com/siemens/ghostwire/v2/network/portfwd"
	"github.com/siemens/ghostwire/v2/network/portfwd/nftget"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/dsl"
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
		ip, proto, port, comment, svcChainName := virtualServiceDetails(svcRules.Exprs)
		if svcChainName == "" {
			continue
		}
		_ = comment // TODO: upstream nufftables
		for _, sepChain := range serviceProviderChains(nattable.ChainsByName[svcChainName]) {
			sc := nattable.ChainsByName[sepChain]
			if sc == nil {
				continue
			}
			for _, rule := range sc.Rules {
				_, dnat := dsl.TargetDNAT(rule.Exprs)
				if dnat == nil {
					continue
				}
				forwardedPorts = append(forwardedPorts, &portfinder.ForwardedPortRange{
					Protocol:       proto,
					IP:             ip,
					PortMin:        port,
					PortMax:        port,
					ForwardIP:      net.IP(dnat.MinIP),
					ForwardPortMin: dnat.MinPort,
				})
			}
		}
	}

	return forwardedPorts
}

// serviceProviderChains returns a list of service separation chain names, given
// a specific service chain.
func serviceProviderChains(chain *nufftables.Chain) (chains []string) {
	if chain == nil {
		return
	}
	for _, rule := range chain.Rules {
		_, chainName := nufftables.OfTypeTransformed(rule.Exprs, getJumpVerdictSeparationChain)
		if chainName == "" {
			continue
		}
		chains = append(chains, chainName)
	}
	return
}

// getJumpVerdictSeparationChain returns the chain name for a service separation
// as given in a jump verdict.
func getJumpVerdictSeparationChain(verdict *expr.Verdict) (string, bool) {
	if verdict.Kind != expr.VerdictJump || !strings.HasPrefix(verdict.Chain, kubeSeparationChainPrefix) {
		return "", false
	}
	return verdict.Chain, true
}

// virtualServiceDetails extracts service details (such as service IP address
// and port, et cetera) from the specified nft expressions. In case of any
// errors, it returns zero values.
func virtualServiceDetails(
	exprs nufftables.Expressions,
) (
	ip net.IP, protocol string, port uint16, comment string, chain string,
) {
	// Try to glance the needed information from the expressions we were given;
	// if there is any problem, then we will end up with nil remaining
	// expressions as our warning signal.
	exprs, protocol = nufftables.OfTypeTransformed(exprs, getTcpUdp)
	exprs, ip = nufftables.OfTypeTransformed(exprs, nftget.IPv46)
	exprs, comment = nufftables.OfTypeTransformed(exprs, getComment)
	exprs, port = nufftables.OfTypeTransformed(exprs, nftget.Port)
	exprs, chain = nufftables.OfTypeTransformed(exprs, getJumpVerdictServiceChain)
	if exprs == nil {
		return nil, "", 0, "", ""
	}
	return
}

// getTcpUdp returns the transport protocol name enclosed in a Cmp expression
// for TCP and UDP, otherwise false.
func getTcpUdp(cmp *expr.Cmp) (string, bool) {
	if len(cmp.Data) != 1 {
		return "", false
	}
	switch cmp.Data[0] {
	case unix.IPPROTO_TCP:
		return "tcp", true
	case unix.IPPROTO_UDP:
		return "udp", true
	}
	return "", false
}

// getComment returns the comment enclosed in a “comment” Match expression, or
// otherwise false if the passed Match expression isn't a comment. Ah, turtles
// all the way down.
//
// Use with [nufftables.OfTypeTransformed].
func getComment(match *expr.Match) (string, bool) {
	if match.Name != "comment" {
		return "", false
	}
	info, ok := match.Info.(*xt.Comment)
	if !ok {
		return "", false
	}
	return string(*info), true
}

// getPort returns the port number from a Cmp expression; otherwise, returns
// false.
func getPort(cmp *expr.Cmp) (uint16, bool) {
	if len(cmp.Data) != 2 {
		return 0, false
	}
	// network order
	return uint16(cmp.Data[0])<<8 + uint16(cmp.Data[1]), true
}

// getJumpVerdictServiceChain returns the chain name for a service as given in a
// jump verdict.
func getJumpVerdictServiceChain(verdict *expr.Verdict) (string, bool) {
	if verdict.Kind != expr.VerdictJump || !strings.HasPrefix(verdict.Chain, kubeServiceChainPrefix) {
		return "", false
	}
	return verdict.Chain, true
}
