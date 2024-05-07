// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package docker

import (
	"github.com/google/nftables/expr"
	"github.com/siemens/ghostwire/v2/network/portfwd"
	"github.com/siemens/ghostwire/v2/network/portfwd/nftget"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/portfinder"
)

// Register this PortForwardings plugin.
func init() {
	plugger.Group[portfwd.PortForwardings]().Register(
		PortForwardings, plugger.WithPlugin("docker"))
}

// PortForwardings discovers Docker's forwarded ports from the “nat” table(s)
// (only for IPv4 and IPv6 respectively).
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
	forwardedPorts := forwardedPortsMk1(nattable)
	forwardedPorts = append(forwardedPorts, forwardedPortsMk2(nattable)...)
	return forwardedPorts
}

func forwardedPortsMk1(nattable *nufftables.Table) []*portfinder.ForwardedPortRange {
	forwardedPorts := []*portfinder.ForwardedPortRange{}
	for _, chain := range nattable.ChainsByName {
		for _, rule := range chain.Rules {
			fp := portfinder.ForwardedPort(rule)
			if fp == nil {
				continue
			}
			log.Debugf("discovered %s", fp)
			forwardedPorts = append(forwardedPorts, fp)
		}
	}
	return forwardedPorts
}

// forwardedPortsMk2 discovers container-local port forwarding rules (especially
// for Docker's embedded DNS resolver) as created by newer Docker versions
// (25+?).
func forwardedPortsMk2(nattable *nufftables.Table) []*portfinder.ForwardedPortRange {
	chain := nattable.ChainsByName["DOCKER_OUTPUT"]
	if chain == nil {
		return nil
	}
	forwardedPorts := []*portfinder.ForwardedPortRange{}
	for _, rule := range chain.Rules {
		exprs, _ := nufftables.OfTypeFunc(rule.Exprs, isL4Proto)
		exprs, proto := nufftables.OfTypeTransformed(exprs, nftget.TcpUdp)
		if exprs == nil {
			continue
		}
	}
	return forwardedPorts
}

func isL4Proto(meta *expr.Meta) bool {
	return meta.Key == expr.MetaKeyL4PROTO
}
