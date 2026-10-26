// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"errors"
	"fmt"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/nonstd/sets"
	"github.com/thediveo/nonstd/xslices"
	"github.com/thediveo/whalewatcher/v2/engineclient/moby"

	"github.com/siemens/ghostwire/v2/decorator/dockernet"
	"github.com/siemens/ghostwire/v2/network"
)

const ComposerServiceLabel = "com.docker.compose.service"

// NeighborhoodServices returns the (deduplicated) list of services (including
// stand-alone containers) DNS addressable from a specific container. If the
// specific container to be taken as the "viewpoint" doesn't exist, an error is
// returned instead.
func NeighborhoodServices(m network.NetworkNamespaces, startContainer *model.Container) (Services, error) {
	networks, err := attachedNetworks(m, startContainer)
	if err != nil {
		return nil, err
	}
	services := Services{}
	// scan all attached networks for services and containers attached to them.
	// This even includes our starting point cntr.
	for _, netw := range networks {
		for _, cntrNif := range attachedNifs(netw) {
			for _, tenant := range cntrNif.Nif().Netns.Tenants {
				// Make sure to work only on the Docker container tenants; other
				// tenants might be present due to network namespace sharing.
				cntr := tenant.Process.Container
				if cntr == nil || cntr.Type != moby.Type {
					continue
				}
				services = append(services, ServiceOnNetworks{
					Name:          cntr.Labels[ComposerServiceLabel],
					Servants:      Containers{cntr},
					NetworkLabels: []string{netw.Nif().Labels[dockernet.GostwireNetworkNameKey]},
				})
			}
		}
	}
	services = services.Deduplicate()
	// We might have collected services connected to more networks than actually
	// are connected to our start position container, so we now prune the
	// network lists of the services to those networks our starting container is
	// attached to.
	for _, service := range services {
		if service.Servants.Contains(startContainer) {
			services.FilterNetworkLabels(service.NetworkLabels)
			break
		}
	}
	return services, nil
}

// attachedNetworks returns the (Docker) user-defined networks attached to the
// specified container, in form of their bridge or MACVLAN/IPVLAN master network
// interfaces. Returns an error if specified container isn't a Docker container.
// Docker's default bridge network is never returned.
func attachedNetworks(m network.NetworkNamespaces, cntr *model.Container) ([]network.Interface, error) {
	if cntr == nil {
		return nil, errors.New("no Docker container identified to discover attached networks from")
	}
	if cntr.Type != moby.Type {
		return nil, fmt.Errorf(
			"container %q is not a Docker container to discover attached networks from", cntr.Name)
	}
	netns := m.ByContainer(cntr)
	if netns == nil {
		//lint:ignore ST1005 because the type of container is important and
		//Docker is a real name.
		return nil, fmt.Errorf("Docker container %q does not exist", cntr.Name)
	}
	networks := sets.New[network.Interface]()
	for _, nif := range netns.Nifs {
		netw := networkOf(nif)
		if netw == nil ||
			netw.Nif().Labels[dockernet.GostwireNetworkNameKey] == "" ||
			netw.Nif().Labels[dockernet.GhostwireNetworkDefaultBridge] == "true" {
			continue
		}
		networks.Add(netw)
	}
	return networks.Elements(), nil
}

// networkOf returns the network interface representing the Docker network the
// specified container network interface is attached to: either the bridge of
// its veth/netkit peer, or its MACVLAN/IPVLAN master. Otherwise, nil.
func networkOf(nif network.Interface) network.Interface {
	if peer := theOtherEnd(nif); peer != nil {
		return peer.Nif().Bridge
	}
	return masterOf(nif)
}

// masterOf returns the master of a MACVLAN or IPVLAN network interface,
// otherwise nil. Returns nil as well when master and MACVLAN/IPVLAN are in the
// same network namespace, as Docker always moves its MACVLANs/IPVLANs into
// container network namespaces, so these are host-side "shim" interfaces
// instead.
func masterOf(nif network.Interface) network.Interface {
	var master network.Interface
	switch vlan := nif.(type) {
	case network.Macvlan:
		master = vlan.Macvlan().Master
	case network.Ipvlan:
		master = vlan.Ipvlan().Master
	}
	if master == nil || master.Nif().Netns == nif.Nif().Netns {
		return nil
	}
	return master
}

// attachedNifs returns the container-side network interfaces attached to the
// specified Docker network, given in form of its bridge or MACVLAN/IPVLAN master
// network interface.
func attachedNifs(netw network.Interface) []network.Interface {
	if bridge, ok := netw.(network.Bridge); ok {
		return xslices.Filter(
			xslices.Map(bridge.Bridge().Ports, theOtherEnd),
			func(nif network.Interface) bool { return nif != nil })
	}
	return xslices.Filter(netw.Nif().Slaves,
		func(nif network.Interface) bool { return masterOf(nif) == netw })
}

// theOtherEnd returns the peer side of a veth or netkit network interface,
// otherwise nil.
func theOtherEnd(nif network.Interface) network.Interface {
	return network.PeerOf(nif)
}
