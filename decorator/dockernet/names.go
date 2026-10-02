// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"github.com/moby/moby/api/types/network"
)

// BridgeNameOptionName optionally specifies the name of the Linux-kernel bridge
// for a Docker “bridge” network. If missing, then the default naming scheme
// applies, taking the first 12 hex digits of the network's ID and prepending
// them with “br-”.
const BridgeNameOptionName = "com.docker.network.bridge.name"

// XvlanParentOptionName optionally specifies the name of a Linux-kernel master
// network interface for Docker “macvlan” and “ipvaln” networks. If missing, the
// driver will instead supply its own dummy-type network interface.
const XvlanParentOptionName = "parent"

// mapNifNamesToDockerNetworks returns a mapping from Linux kernel network
// interface names to their corresponding Docker networks, if any.
func mapNifNamesToDockerNetworks(nets []network.Network) map[string]*network.Network {
	m := map[string]*network.Network{}
	for idx := range nets {
		net := &nets[idx]
		fn := nifnamerByNetworkType[net.Driver]
		if fn == nil {
			continue
		}
		nifname := fn(net)
		if nifname == "" {
			continue
		}
		m[nifname] = net
	}
	return m
}

// nifnamer is a function that given a Docker (custom) network then returns the
// name of the corresponding Linux network interface.
type nifnamer func(*network.Network) string

// nifnamerByNetworkType maps docker network driver names to their corresponding
// network interface naming functions.
var nifnamerByNetworkType = map[string]nifnamer{
	"bridge":  bridgeNifName,
	"ipvlan":  xvlanNifName,
	"macvlan": xvlanNifName,
}

// bridgeNifName returns the name of the bridge network interface for the
// specified Docker bridge network. Docker bridge networks typically don't
// explicitly store the bridge interface name in their configurations, but
// instead the bridge name is derived implicitly from part of the networks
// unique ID hex string. However, it is possible to assign an already existing
// manually created bridge to a bridge network by specifing its name in the
// network's options.
func bridgeNifName(net *network.Network) string {
	if brname, ok := net.Options[BridgeNameOptionName]; ok {
		return brname // ...explicitly configured bridge nif name.
	}
	return "br-" + net.ID[0:12] // ...auto-generated nif name.
}

// xvlanNifName returns the name of either the master network interface or the
// dummy-type replacement interface for the specified Docker macvlan network.
func xvlanNifName(net *network.Network) string {
	if mastername, ok := net.Options[XvlanParentOptionName]; ok {
		return mastername // ...explicitly configured bridge nif name.
	}
	return "dm-" + net.ID[0:12] // ...auto-generated nif name for the dummy master.
}
