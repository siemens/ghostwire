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
		ip, proto, port, comment, svcChainName := virtualServiceDetails(svcRules.Exprs)
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

// isDNAT returns true, if the passed nft Target expression is a DNAT target
// expression, otherwise false.
func isDNAT(target *expr.Target) bool {
	if target.Name != "DNAT" {
		return false
	}
	_, ok := target.Info.(*xt.NatRange2)
	return ok
}

// separations returns a list of service separation chain names, given a
// specific service chain.
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
	exprs, ip = nufftables.OfTypeTransformed(exprs, getIPv46)
	exprs, comment = nufftables.OfTypeTransformed(exprs, getComment)
	exprs, port = nufftables.OfTypeTransformed(exprs, getPort)
	exprs, chain = nufftables.OfTypeTransformed(exprs, getJumpVerdictChain)
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

// getIPv46 returns the IPv4 or IPv6 address enclosed in a Cmp expression,
// otherwise false.
func getIPv46(cmp *expr.Cmp) (net.IP, bool) {
	switch len(cmp.Data) {
	case 4, 16:
		return net.IP(cmp.Data), true
	}
	return nil, false
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

// getJumpVerdictChain returns the chain name for a service as given in a jump
// verdict.
func getJumpVerdictChain(verdict *expr.Verdict) (string, bool) {
	if verdict.Kind != expr.VerdictJump || !strings.HasPrefix(verdict.Chain, kubeServiceChainPrefix) {
		return "", false
	}
	return verdict.Chain, true
}
