// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"context"
	"fmt"
	"strings"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/network"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

// GostwireNetworkNameKey defines the label key for storing the Docker network
// name of bridge networks.
const GostwireNetworkNameKey = "gostwire/network/name"

// BridgeNameOptionName optionally specifies the name of the Linux-kernel bridge
// for a Docker "bridge" network. If missing, then the default naming scheme
// applies, taking the first 12 hex digits of the network's ID and prepending
// them with "br-".
const BridgeNameOptionName = "com.docker.network.bridge.name"

// PassthroughHostIfnameOptionName specifies the name of a host network
// interface to be passed through into a single network namespace (sandbox).
const PassthroughHostIfnameOptionName = "ifname"

// GhostwireNetworkDefaultBridge defines the label key for signalling that a
// particular network is Docker's "default bridge". For instance, service and
// container DNS name resolution is disabled on the default bridge. Allowed
// values are "true" and "false". A non-existing key means "false".
const GhostwireNetworkDefaultBridge = "gostwire/network/default-bridge"

// DefaultBridgeOptionName optionally identifies a Docker network as the default
// network.
const DefaultBridgeOptionName = "com.docker.network.bridge.default_bridge"

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("dockernet"))
}

// dockerNetworks describes the networks managed by a Docker engine from
// Gostwire's perspective: the Docker-related information, as well as the
// NetworkNamespace these networks are managed in, which are the network
// namespace of the particular managing Docker engine.
type dockerNetworks struct {
	networks    []types.NetworkResource   // Docker-managed network information.
	engine      *model.ContainerEngine    // corresponding Docker engine.
	engineNetns *network.NetworkNamespace // ...of the managing Docker engine.
}

// makeDockerNetworks returns a dockerNetworks object with the networks managed
// by the specified Docker engine. If discovery failed, a zero-value
// dockerNetworks object will be returned instead, to be used in the engine map
// to signal that we asked the engine, but it failed, so no more attempts to
// talk to it, please.
func makeDockerNetworks(ctx context.Context, engine *model.ContainerEngine, allnetns network.NetworkNamespaces) (
	docknets dockerNetworks,
) {
	dockerclient, err := client.NewClientWithOpts(
		client.WithHost(engine.API),
		client.WithAPIVersionNegotiation())
	if err != nil {
		log.Warnf("cannot discover Docker-managed networks from API %s, reason: %s",
			engine.API, err.Error())
		return
	}
	networks, _ := dockerclient.NetworkList(ctx, types.NetworkListOptions{})
	_ = dockerclient.Close()
	netnsid, _ := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/net", engine.PID)).ID()
	docknets.networks = networks
	docknets.engine = engine
	docknets.engineNetns = allnetns[netnsid]
	log.Infof("found %d Docker networks related to net:[%d] %s",
		len(networks), docknets.engineNetns.ID().Ino, docknets.engineNetns.DisplayName())
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
	log.Debugf("discovering Docker-managed networks")
	// As some container engines currently might not manage any container
	// workload, we will prime the container engine networks cache with the
	// networks discovered from then engines we're told are under supervision.
	// This way, we ensure to discover networks even for engines without any
	// workload, because otherwise we won't see them at all in the containers
	// attached to the network namespaces.
	dockerNets := map[model.PIDType]dockerNetworks{}
	for _, engine := range engines {
		if engine.Type != moby.Type {
			continue
		}
		dockerNets[engine.PID] = makeDockerNetworks(ctx, engine, allnetns)
	}
	// Now that we know about the Docker networks, try to locate the matching
	// Linux-kernel network interfaces so we can set/override the alias names of
	// the interfaces.
	for _, docknet := range dockerNets {
		for _, netw := range docknet.networks {
			var nifname string
			switch netw.Driver {
			case "bridge":
				nifname = linuxBridgeName(netw) // no explicit name, but always implicit
			case "macvlan":
				nifname = netw.Options["parent"]
			default:
				continue
			}
			// Try to locate the Linux network interface related to this Docker
			// network, and if successful, set its alias name.
			netif, ok := docknet.engineNetns.NamedNifs[nifname]
			if !ok {
				continue
			}
			nif := netif.Nif()
			nif.Alias = netw.Name
			// We additionally also label the network interface with the Docker
			// network name.
			nif.Labels[GostwireNetworkNameKey] = netw.Name
			nif.AddLabels(netw.Labels)
			// In case this is an "internal" Docker bridge network, then label
			// it as being internal. While this only applies to "bridge"-driver
			// networks, the flag is always present in Docker's API structure,
			// so we don't need to differentiate here.
			if netw.Internal {
				nif.Labels[network.GostwireInternalBridgeKey] = "true"
			}
			// In case of Docker's default network/bridge, label it so.
			if netw.Options[DefaultBridgeOptionName] == "true" {
				nif.Labels[GhostwireNetworkDefaultBridge] = "true"
			}
		}
	}
	// Next up: process any passthrough networks present...
	for _, docknet := range dockerNets {
		throughpassers := map[string]PassedThrough{}
		for _, netw := range docknet.networks {
			retrievedDetails := false
			// when using the managed plugin, the driver name contains a tag,
			// such as ":latest". No Docker, this isn't thought out well.
			if netw.Driver != "passthrough" && !strings.HasPrefix(netw.Driver, "passthrough:") {
				continue
			}
			nifname := netw.Options[PassthroughHostIfnameOptionName]
			if nifname == "" {
				continue
			}
			log.Debugf("found passthrough network %s with network interface %s",
				netw.Name, nifname)
			// Try to locate the Linux network interface related to this Docker
			// network, and if successful, set its alias name. We even try this
			// for passthrough network interfaces as they might be in transition
			// between the endpoint created but not yet joined to a container's
			// sandbox.
			netif, ok := docknet.engineNetns.NamedNifs[nifname]
			if !ok {
				// Okay, the network interface currently is passed through into
				// container, as we don't find it in the Docker engine's
				// ("host") network namespace.
				//
				// We now need more details in order to locate the passed
				// through network interface ... that has been renamed and
				// Docker won't tell us the new name. Oh, well.
				//
				// Before, we only "listed" the available Docker networks and
				// this will not return the connected container/endpoint
				// details. We thus now inspect the current passthrough network
				// more closely, in order to get the connected container
				// (endpoint) information. We only need to do it once per
				// passthrough network.
				if !retrievedDetails {
					retrievedDetails = true
					dockerclient, err := client.NewClientWithOpts(
						client.WithHost(docknet.engine.API),
						client.WithAPIVersionNegotiation())
					if err == nil {
						netwDetails, err := dockerclient.NetworkInspect(ctx, netw.ID, types.NetworkInspectOptions{})
						_ = dockerclient.Close()
						if err == nil {
							netw = netwDetails
						} else {
							log.Errorf("cannot inspect Docker network %s more closely, reason: %s",
								netw.Name, err.Error())
						}
					}
				}
				// passthrough can only have at most a single sandbox attached,
				// so take the first container with its endpoint information we
				// can find.
				for cntrID, cntr := range netw.Containers {
					log.Debugf("network interface %s currently passed into container with ID %s",
						nifname, cntrID)
					// Just get the first and only connected sandbox/container.
					throughpassers[cntrID] = PassedThrough{
						Alias:     netw.Name,
						HwAddress: cntr.MacAddress,
					}
					break
				}
				continue
			}
			log.Debugf("network interface %s not yet passed through", nifname)
			nif := netif.Nif()
			nif.Alias = netw.Name
			// We additionally also label the network interface with the Docker
			// network name.
			nif.Labels[GostwireNetworkNameKey] = netw.Name
			nif.AddLabels(netw.Labels)
		}
		// Did we find any currently passed-through network interfaces that
		// weren't in their home position? Then we now need to find the
		// containers that "contain" these passed-through network interfaces. We
		// can identify them only by the endpoint associated container and the
		// hardware MAC address.
		if len(throughpassers) > 0 {
			log.Debugf("scanning containers for passed-through nifs...")
			for _, netns := range allnetns {
				for _, tenant := range netns.Tenants {
					if tenant.Process == nil {
						continue
					}
					cntr := tenant.Process.Container
					if cntr == nil || cntr.Type != moby.Type {
						continue
					}
					passedthrough, ok := throughpassers[cntr.ID]
					if !ok {
						continue
					}
					log.Debugf("found network interface %s (MAC %s) inside container %s",
						passedthrough.Alias, passedthrough.HwAddress, cntr.Name)
					for _, netif := range netns.Nifs {
						hwaddr := netif.Nif().Link.Attrs().HardwareAddr.String()
						if hwaddr == passedthrough.HwAddress && hwaddr != "" {
							nif := netif.Nif()
							nif.Alias = passedthrough.Alias
							// We additionally also label the network interface with the Docker
							// network name.
							nif.Labels[GostwireNetworkNameKey] = passedthrough.Alias
						}
					}
				}
			}
		}
	}
}

type PassedThrough struct {
	Alias     string
	HwAddress string
}

// linuxBridgeName returns the name of the bridge network interface for the
// given Docker bridge network. Docker bridge networks don't explicitly store
// the bridge interface name in their configurations, but instead the bridge
// name is derived implicitly from part of the networks unique ID hex string.
func linuxBridgeName(netw types.NetworkResource) string {
	if brname, ok := netw.Options[BridgeNameOptionName]; ok {
		return brname // ...explicitly configured bridge nif name.
	}
	return "br-" + netw.ID[0:12] // ...auto-generated nif name.
}
