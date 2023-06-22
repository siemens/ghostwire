// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"errors"
	"fmt"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/engineclient/moby"

	"github.com/siemens/ghostwire/v2/decorator/dockernet"
	"github.com/siemens/ghostwire/v2/network"
)

const ComposerServiceLabel = "com.docker.compose.service"

// NeighborhoodServices returns the (deduplicated) list of services (including
// stand-alone containers) DNS addressable from a specific container. If the
// specific container to be taken as the "viewpoint" doesn't exist, an error is
// returned instead.
func NeighborhoodServices(m network.NetworkNamespaces, startContainer *model.Container) (Services, error) {
	bridgeNetworks, err := attachedNetworks(m, startContainer)
	if err != nil {
		return nil, err
	}
	services := Services{}
	// scan all attached networks (bridges) for services and containers attached
	// to them. This even includes our starting point cntr.
	for _, netw := range bridgeNetworks {
		for _, port := range netw.(network.Bridge).Bridge().Ports {
			cntrNif := theOtherEnd(port)
			if cntrNif == nil {
				continue // not a veth or "incomplete" veth where we didn't find the other end.
			}
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
// specified container, in form of network interfaces. Returns an error if
// specified container isn't a Docker container. Only non-default bridge Docker
// networks are returned.
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
	bridges := map[network.Interface]struct{}{}
	for _, nif := range netns.Nifs {
		bridgeport := theOtherEnd(nif)
		if bridgeport == nil {
			continue
		}
		bridge := bridgeport.Nif().Bridge
		if bridge == nil ||
			bridge.Nif().Labels[dockernet.GostwireNetworkNameKey] == "" ||
			bridge.Nif().Labels[dockernet.GhostwireNetworkDefaultBridge] == "true" {
			continue
		}
		bridges[bridge] = struct{}{}
	}
	networks := make([]network.Interface, 0, len(bridges))
	for bridge := range bridges {
		networks = append(networks, bridge)
	}
	return networks, nil
}

// theOtherEnd returns the peer side of a veth network interface, otherwise nil.
func theOtherEnd(nif network.Interface) network.Interface {
	if nif.Nif().Kind != "veth" {
		return nil
	}
	return nif.(network.Veth).Veth().Peer
}
