// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"context"
	"fmt"
	"log/slog"

	mobynet "github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/nonstd/sets"
	"github.com/thediveo/nonstd/xslices"
	"github.com/thediveo/nonstd/xslog"
	"github.com/thediveo/whalewatcher/v2/engineclient/moby"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/passedthrough"
)

// GostwireNetworkNameKey defines the label key for storing the Docker network
// name.
const GostwireNetworkNameKey = "gostwire/network/name"

// GostwireNetworkIDKey defines the labey key for storing the Docker network ID.
const GostwireNetworkIDKey = "gostwire/network/id"

// GhostwireNetworkDefaultBridge defines the label key for signalling that a
// particular network is Docker's "default bridge", better known to many as
// simply "docker0". For instance, service and container DNS name resolution is
// disabled on the default bridge. Allowed values are "true" and "false". A
// non-existing key means "false".
const GhostwireNetworkDefaultBridge = "gostwire/network/default-bridge"

// GhostwireInterfaceOrigNameKey defines the label key for storing the original
// name (when it was in the “host” network namespace) of a network interface.
const GhostwireInterfaceOrigNameKey = "gostwire/interface/orig-name"

// DefaultBridgeOptionName optionally identifies a Docker network as the default
// (legacy) network, also known as simply "docker0".
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
	networks    []mobynet.Network         // Docker-managed network information
	engine      *model.ContainerEngine    // corresponding Docker engine.
	engineNetns *network.NetworkNamespace // ...of the managing Docker engine.
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
	slog.Debug("discovering Docker-managed networks")
	// As some container engines currently might not manage any container
	// workload, we will prime the container engine networks cache with the
	// networks discovered from the engines we're told are under supervision.
	// This way, we ensure to discover networks even for engines without any
	// workload, because otherwise we won't see them at all in the containers
	// attached to the network namespaces.
	engineDockerNetworks := map[model.PIDType]dockerNetworks{}
	for _, engine := range engines {
		if engine.Type != moby.Type {
			continue
		}
		engineDockerNetworks[engine.PID] = dockerNetworksOfEngine(ctx, engine, allnetns)
	}

	// Now that we know about the Docker networks, try to locate the matching
	// Linux-kernel network interfaces so we can set/override the alias names of
	// the interfaces, as well as additional tidbits of intelligence.
	for _, dockernets := range engineDockerNetworks {
		processDockernetInterfaces(dockernets)
		processPassedThroughInterfaces(dockernets, allnetns)
	}
}

// process the Docker networks of a particular Docker engine, mapping the
// network interfaces in the same network namespace as the Docker engine to
// their corresponding Docker custom networks.
//
//   - sets the interface's Alias detail to the name of the Docker network.
//   - adds labels for Docker custom network name and ID.
//   - adds any Docker custom network labels to the interface details.
//   - if the Docker network is configured as "internal", it adds dedicated
//     marker label.
//   - if the Docker network is the default "docker0" network, it adds a
//     dedicated marker label.
func processDockernetInterfaces(dockernets dockerNetworks) {
	nifNamesToDockerNets := mapNifNamesToDockerNetworks(dockernets.networks)
	for nifname, dnet := range nifNamesToDockerNets {
		netif, ok := dockernets.engineNetns.NamedNifs[nifname]
		if !ok {
			continue
		}
		nif := netif.Nif()
		// Set the interface's alias to the name of the Docker network.
		nif.Alias = dnet.Name
		// We additionally also label the network interface with the Docker
		// network name.
		nif.Labels[GostwireNetworkNameKey] = dnet.Name
		nif.Labels[GostwireNetworkIDKey] = dnet.ID
		nif.AddLabels(dnet.Labels)
		// In case this is an "internal" Docker network, then label it as being
		// internal. The flag is always present in Docker's API structure and
		// actually applies to other network types than just bridge networks.
		if dnet.Internal {
			nif.Labels[network.GostwireInternalBridgeKey] = "true"
		}
		// In case of Docker's default network/bridge, label it so.
		if dnet.Options[DefaultBridgeOptionName] == "true" {
			nif.Labels[GhostwireNetworkDefaultBridge] = "true"
		}
	}
}

// process the network interfaces inside the Docker containers of a particular
// Docker engine, mapping passed-through interfaces to their Docker networks and
// adding information their original name before they were passed-through.
func processPassedThroughInterfaces(dockernets dockerNetworks, allnetns network.NetworkNamespaces) {
	// Since Docker containers can share network namespaces, keep a map of the
	// network namespaces already seen, do skip duplicate work.
	dockernetsByID := map[string]*mobynet.Network{}
	for idx := range dockernets.networks {
		dockernet := &dockernets.networks[idx]
		dockernetsByID[dockernet.ID] = dockernet
	}

	visitedNetns := sets.New[*network.NetworkNamespace]()
	for _, cntr := range dockernets.engine.Containers {
		netns := allnetns.ByContainer(cntr)
		if visitedNetns.Contains(netns) {
			continue
		}
		visitedNetns.Add(netns)
		// nota bene: if there are containers with "net:host" then we will
		// eventually be here working on the "host" network namespace, or
		// rather: the engine's network namespace. Now, as network interfaces in
		// that network namespace are never passed-through they won't have a
		// Docker network ID attached (we also rely on libnetwork giving the
		// boot to any attempt to attach a Docker network with a custom
		// passthrough driver to a container with "net:host"). Thus, such
		// network interfaces won't trigger.
		for _, netif := range netns.Nifs {
			link := netif.Nif().Link
			origNifname := passedthrough.OriginalIfname(link)
			dockernetID := passedthrough.DockernetID(link)
			if origNifname == "" || dockernetID == "" {
				// skip this interface while in the engine's ("host") network
				// namespace, as it isn't currently passed-through.
				continue
			}
			nif := netif.Nif()
			dnet := dockernetsByID[dockernetID]
			if dnet == nil {
				continue
			}
			// Set the interface's alias to the name of the Docker network.
			slog.Warn("passed-through network interface",
				slog.String("nifname", nif.Name),
				slog.String("alias", dnet.Name),
			)
			nif.Alias = dnet.Name
			// Label the original interface name.
			nif.Labels[GhostwireInterfaceOrigNameKey] = origNifname
			// We additionally also label the network interface with the Docker
			// network name.
			nif.Labels[GostwireNetworkNameKey] = dnet.Name
			nif.Labels[GostwireNetworkIDKey] = dnet.ID
			nif.AddLabels(dnet.Labels)
		}
	}
}

type PassedThrough struct {
	Alias     string
	HwAddress string
}

// dockerNetworksOfEngine returns a dockerNetworks object with the networks managed
// by the specified Docker engine. If discovery failed, a zero-value
// dockerNetworks object will be returned instead, to be used in the engine map
// to signal that we asked the engine, but it failed, so no more attempts to
// talk to it, please.
func dockerNetworksOfEngine(ctx context.Context, engine *model.ContainerEngine, allnetns network.NetworkNamespaces) (
	docknets dockerNetworks,
) {
	moby, err := client.New(client.WithHost(engine.API))
	if err != nil {
		slog.Warn("cannot discover Docker-managed networks",
			slog.String("api", engine.API), xslog.Error(err))
		return
	}
	networks, _ := moby.NetworkList(ctx, client.NetworkListOptions{})
	_ = moby.Close()
	netnsid, _ := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/net", engine.PID)).ID()
	docknets.networks = xslices.Map(networks.Items,
		func(s mobynet.Summary) mobynet.Network { return s.Network })
	docknets.engine = engine
	docknets.engineNetns = allnetns[netnsid]
	slog.Info("found Docker custom networks",
		slog.Int("count", len(networks.Items)),
		slog.Uint64("netns", docknets.engineNetns.ID().Ino),
		slog.String("name", docknets.engineNetns.DisplayName()))
	return
}
