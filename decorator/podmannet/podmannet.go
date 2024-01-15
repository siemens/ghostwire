// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package podmannet

import (
	"context"
	"fmt"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/decorator/dockernet"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/turtlefinder/activator/podman"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
)

// GostwireNetworkNameKey defines the label key for storing the Docker network
// name of bridge networks.
const GostwireNetworkNameKey = dockernet.GostwireNetworkNameKey

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("podmannet-v4+"))
}

// podmanNetworks describes the networks managed by a podman engine from
// Gostwire's perspective: the Docker-related information, as well as the
// NetworkNamespace these networks are managed in, which are the network
// namespace of the particular managing podman engine.
type podmanNetworks struct {
	networks    []NetworkResource         // podman-managed network information.
	engine      *model.ContainerEngine    // corresponding podman engine.
	engineNetns *network.NetworkNamespace // ...of the managing podman engine.
}

// makePodmanNetworks returns a podmanNetworks object with the networks managed
// by the specified Docker engine. If discovery failed, a zero-valued
// podmanNetworks object will be returned instead, to be used in the engine map
// to signal that we asked the engine, but it failed, so no more attempts to
// talk to it, please.
func makePodmanNetworks(ctx context.Context, engine *model.ContainerEngine, allnetns network.NetworkNamespaces) (
	podmannets podmanNetworks,
) {
	libpodclient, err := newLibpodClient(engine.API)
	if err != nil {
		log.Warnf("cannot discover podman-managed networks from API %s, reason: %s",
			engine.API, err.Error())
		return
	}
	info, err := libpodclient.info(ctx)
	if err != nil {
		log.Warnf("cannot discover podman-managed networks from API %s, reason: %s",
			engine.API, err.Error())
		return
	}
	libpodclient.libpodVersion = info.Version.APIVersion
	networks, _ := libpodclient.NetworkList(ctx)
	_ = libpodclient.Close()
	netnsid, _ := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/net", engine.PID)).ID()
	podmannets.networks = networks
	podmannets.engine = engine
	podmannets.engineNetns = allnetns[netnsid]
	log.Infof("found %d podman networks related to net:[%d] %s",
		len(networks), podmannets.engineNetns.ID().Ino, podmannets.engineNetns.DisplayName())
	return
}

// Decorate decorates bridge and macvlan master network interfaces with alias
// names that are the names of their corresponding Docker “bridge” or “macvlan”
// networks, where applicable (a copy is stored also in the labels in Gostwire's
// key namespace). Additionally, it copies over any user-defined network labels.
func Decorate(
	ctx context.Context,
	allnetns network.NetworkNamespaces,
	allprocs model.ProcessTable,
	engines []*model.ContainerEngine,
) {
	log.Debugf("discovering podman-managed networks")
	// As some container engines currently might not manage any container
	// workload, we will prime the container engine networks cache with the
	// networks discovered from then engines we're told are under supervision.
	// This way, we ensure to discover networks even for engines without any
	// workload, because otherwise we won't see them at all in the containers
	// attached to the network namespaces.
	podmanNets := map[model.PIDType]podmanNetworks{}
	for _, engine := range engines {
		if engine.Type != podman.Type {
			continue
		}
		podmanNets[engine.PID] = makePodmanNetworks(ctx, engine, allnetns)
	}
	// Now that we know about the podman networks, try to locate the matching
	// Linux-kernel network interfaces so we can set/override the alias names of
	// the interfaces.
	for _, podmannet := range podmanNets {
		for _, netw := range podmannet.networks {
			var nifname string
			switch netw.Driver {
			case "bridge", "macvlan":
				nifname = netw.NetworkInterface
			default:
				continue
			}
			// Try to locate the Linux network interface related to this podman
			// network, and if successful, set its alias name.
			netif, ok := podmannet.engineNetns.NamedNifs[nifname]
			if !ok {
				continue
			}
			nif := netif.Nif()
			nif.Alias = netw.Name
			// We additionally also label the network interface with the Docker
			// network name.
			nif.Labels[GostwireNetworkNameKey] = netw.Name
			nif.AddLabels(netw.Labels)
			// In case this is an "internal" podman bridge network, then label
			// it as being internal. While this only applies to "bridge"-driver
			// networks, the flag is always present in libpod's API structure,
			// so we don't need to differentiate here.
			if netw.Internal {
				nif.Labels[network.GostwireInternalBridgeKey] = "true"
			}
			// Is there something like Docker's default network...?
		}
	}
}
