// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package docker

import (
	"github.com/siemens/ghostwire/v2/network/portfwd"
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
