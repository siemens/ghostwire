// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package docker

import (
	"github.com/google/nftables/xt"
	"github.com/siemens/ghostwire/v2/network/portfwd"
	"github.com/siemens/ghostwire/v2/network/portfwd/nftget"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/dsl"
	"github.com/thediveo/nufftables/portfinder"
)

// Register this PortForwardings plugin.
func init() {
	plugger.Group[portfwd.PortForwardings]().Register(
		PortForwardings, plugger.WithPlugin("docker"))
}

// PortForwardings discovers Docker's forwarded ports from the “nat” table(s)
// (only for IPv4 and IPv6 respectively). As nftables expression generation is
// considered to be the private black art by the nftables project and we simply
// don't want to call arbitrary binaries, including nftables binaries, we need
// to live with expression generation changing from time to time.
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
	return grabPortForwardings(nattable)
}

// grabPortForwardings is a convenience helper to wire up the individual port
// forwarding detectors in a single place, making maintenance easier for both
// PROD and TEST.
func grabPortForwardings(nattable *nufftables.Table) []*portfinder.ForwardedPortRange {
	forwardedPorts := forwardedPortsMk1(nattable)
	forwardedPorts = append(forwardedPorts, forwardedPortsMk2(nattable)...)
	forwardedPorts = append(forwardedPorts, forwardedPortsMk3(nattable)...)
	return forwardedPorts
}

// forwardedPortsMk1 discovers host and container-local port forwarding rules
// (especially for Docker's embedded DNS resolver) as created by older Docker
// versions and/or an older iptables compatibility layer.
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

// forwardedPortsInChainMk2 discovers container-local port forwarding rules from
// the given chain. When given a nil chain, it simply returns a nil slice.
func forwardedPortsInChainMk2(chain *nufftables.Chain) []*portfinder.ForwardedPortRange {
	if chain == nil {
		return nil
	}
	family := chain.Table.Family
	forwardedPorts := []*portfinder.ForwardedPortRange{}
	for _, rule := range chain.Rules {
		exprs, proto := nftget.MetaL4ProtoTcpUdp(rule.Exprs)
		exprs, origIP := nftget.OptionalIPv46(exprs, family)
		exprs, port := nufftables.OfTypeTransformed(exprs, nftget.Port)
		exprs, dnat := dsl.TargetDNAT(exprs)
		if exprs == nil || dnat.Flags&dnatWithIPsAndPorts != dnatWithIPsAndPorts || port == 0 {
			continue
		}
		fp := &portfinder.ForwardedPortRange{
			Protocol:       proto,
			IP:             origIP,
			PortMin:        port,
			PortMax:        port,
			ForwardIP:      dnat.MinIP,
			ForwardPortMin: dnat.MinPort,
		}
		log.Debugf("discovered %s", fp)
		forwardedPorts = append(forwardedPorts, fp)
	}
	return forwardedPorts
}

// dnatWithIPsAndPorts are the flags that need to be set in order for the
// xt.NatRange(2) data structures to contain an IP range as well as a transport
// layer port range.
const dnatWithIPsAndPorts = uint(xt.NatRangeMapIPs | xt.NatRangeProtoSpecified)

// forwardedPortsMk2 discovers container-local port forwarding rules (such as
// for Docker's embedded DNS resolver) as created by newer Docker versions
// (25+?) and/or "newer" iptables compatibility layer.
func forwardedPortsMk2(nattable *nufftables.Table) []*portfinder.ForwardedPortRange {
	forwardedPorts := forwardedPortsInChainMk2(nattable.ChainsByName["DOCKER"])
	forwardedPorts = append(forwardedPorts, forwardedPortsInChainMk2(nattable.ChainsByName["DOCKER_OUTPUT"])...)
	return forwardedPorts
}

func forwardedPortsInChainMk3(chain *nufftables.Chain) []*portfinder.ForwardedPortRange {
	if chain == nil {
		return nil
	}
	family := chain.Table.Family
	forwardedPorts := []*portfinder.ForwardedPortRange{}
	for _, rule := range chain.Rules {
		exprs, origIP := nftget.OptionalDestIPv46(rule.Exprs, family)
		exprs, proto := nftget.PayloadL4ProtoTcpUdp(exprs)
		exprs, port := nftget.PayloadPort(exprs)
		exprs, dnat := dsl.TargetDNAT(exprs)
		if exprs == nil || dnat.Flags&dnatWithIPsAndPorts != dnatWithIPsAndPorts || port == 0 {
			continue
		}
		fp := &portfinder.ForwardedPortRange{
			Protocol:       proto,
			IP:             origIP,
			PortMin:        port,
			PortMax:        port,
			ForwardIP:      dnat.MinIP,
			ForwardPortMin: dnat.MinPort,
		}
		log.Debugf("discovered %s", fp)
		forwardedPorts = append(forwardedPorts, fp)
	}
	return forwardedPorts
}

func forwardedPortsMk3(nattable *nufftables.Table) []*portfinder.ForwardedPortRange {
	forwardedPorts := forwardedPortsInChainMk3(nattable.ChainsByName["DOCKER"])
	forwardedPorts = append(forwardedPorts, forwardedPortsInChainMk3(nattable.ChainsByName["DOCKER_OUTPUT"])...)
	return forwardedPorts
}
