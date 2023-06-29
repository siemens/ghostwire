// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctlnet

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/decorator/dockernet"
	"github.com/siemens/ghostwire/v2/network"

	"github.com/containernetworking/cni/libcni"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/whalewatcher/watcher/containerd"
)

// NetworkConfigurationsGlob specifies the location only of the CNI network
// configuration list files.
const NetworkConfigurationsDir = "/etc/cni/net.d"

// NetworkConfigurationsGlob specifies the pattern of the CNI network
// configuration list files.
const NetworkConfigurationsGlob = "nerdctl-*.conflist"

// GostwireNetworkNameKey defines the label key for storing the nerdctl "Docker"
// network name of bridge networks.
const GostwireNetworkNameKey = dockernet.BridgeNameOptionName

// Register this Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("nerdctlnet"))
}

// nerdctlNetworks describes the networks managed by nerdctl for a containerd
// engine, from Gostwire's perspective: the nerdctl/CNI-related information, as
// well as the NetworkNamespace these networks are managed in, which are the
// network namespace of the particular managing containerd engine.
type nerdctlNetworks struct {
	networks    []nerdctlNetwork          // nerdctl-managed network information.
	engineNetns *network.NetworkNamespace // ...of the managing containerd engine.
}

// nerdctlNetwork describes the nerdctl-managed CNI configuration, which is a
// superset of CNI's ordered list of networks.
//
// Note: recent nerdctl versions do not create the underlying bridge netdev when
// creating a custom network, but only allocate a name. The bridge netdev will
// only be created later when the first sandbox needs to be wired up to the
// custom network.
//
// Note: as nerdctl has no official API we have to define our own versions of
// how nerdctl stores network configuration information, to some extend
// mirroring some things. Sigh.
type nerdctlNetwork struct {
	*libcni.NetworkConfigList
	ID     string       // nerdctl network ID.
	Labels model.Labels // optionally attached labels ("nerdctlLabels").
}

// Plugin returns the plugin configuration information for the (first) plugin of
// specified type. Otherwise returns nil, if a plugin of the specified type
// cannot be found.
func (n *nerdctlNetwork) Plugin(typ string) *libcni.NetworkConfig {
	for _, plugin := range n.Plugins {
		if plugin.Network.Type == typ {
			return plugin
		}
	}
	return nil
}

// Plugin returns the string value of the specified field of the (first) plugin
// of the specified type. Otherwise, it returns "".
//
// A typical example is to ask for the field "bridge" of the CNI plugin of type
// "bridge": the bridge field then is the name of the virtual bridge netdev.
func (n *nerdctlNetwork) PluginField(typ string, field string) string {
	plugin := n.Plugin(typ)
	if plugin == nil {
		return "" // no such type of plugin.
	}
	rawFields := map[string]interface{}{}
	if json.Unmarshal(plugin.Bytes, &rawFields) != nil {
		return "" // something's rotten here.
	}
	if val, ok := rawFields[field]; ok {
		if sval, ok := val.(string); ok {
			return sval
		}
	}
	return "" // either no such field or it ain't contain a string.
}

// newNerdctlNetworks returns configuration information about the
// nerdctl-managed networks for the specified containerd engine.
func newNerdctlNetworks(ctx context.Context, engine *model.ContainerEngine, allnetns network.NetworkNamespaces) nerdctlNetworks {
	netnsid, _ := ops.NamespacePath(fmt.Sprintf("/proc/%d/ns/net", engine.PID)).ID()
	nerdynets := nerdctlNetworks{
		engineNetns: allnetns[netnsid],
		networks:    []nerdctlNetwork{},
	}
	mntneer, err := mountineer.New(model.NamespaceRef{fmt.Sprintf("/proc/%d/ns/mnt", engine.PID)}, nil)
	if err != nil {
		log.Errorf("cannot access mount namespace of nerdctl engine, reason: %w", err)
		return nerdynets
	}
	defer mntneer.Close()
	netwConfigsDir, err := mntneer.Resolve(NetworkConfigurationsDir)
	if err != nil {
		log.Errorf("cannot resolve CNI plugins configuration path, reason: %w", err)
		return nerdynets
	}
	configFilenames, err := filepath.Glob(filepath.Join(netwConfigsDir, NetworkConfigurationsGlob))
	if err != nil {
		return nerdynets
	}
	for _, configFilename := range configFilenames {
		log.Debugf("found CNI configuration file %q", configFilename)
		nerdynetworkconf, err := libcni.ConfListFromFile(configFilename)
		if err != nil {
			log.Errorf("invalid CNI configuration file %q, reason: %w", configFilename, err)
			continue
		}
		// Oh well ... libcni puts the original raw JSON into the "Bytes"
		// rucksack and we now try to extract the additional nerdctl-related
		// fields out of it.
		rawFields := struct {
			ID     string            `json:"nerdctlID"`
			Labels map[string]string `json:"nerdctlLabels"`
		}{}
		if json.Unmarshal(nerdynetworkconf.Bytes, &rawFields) != nil {
			continue
		}
		nerdynets.networks = append(nerdynets.networks, nerdctlNetwork{
			NetworkConfigList: nerdynetworkconf,
			ID:                rawFields.ID,
			Labels:            rawFields.Labels,
		})
	}
	return nerdynets
}

// Decorate decorates bridge network interfaces with alias names that are the
// names of their corresponding nerdctl-managed CNI "bridge" networks, where
// applicable (a copy is stored also in the labels in Gostwire's key namespace).
// Additionally, it copies over any user-defined network labels.
func Decorate(
	ctx context.Context,
	allnetns network.NetworkNamespaces,
	allprocs model.ProcessTable,
	engines []*model.ContainerEngine,
) {
	log.Debugf("discovering nerdctl-managed CNI networks")
	// As some container engines currently might not manage any container
	// workload, we will prime the container engine networks cache with the
	// networks discovered from then engines we're told are under supervision.
	// This way, we ensure to discover networks even for engines without any
	// workload, because otherwise we won't see them at all in the containers
	// attached to the network namespaces.
	nerdctlNets := map[model.PIDType]nerdctlNetworks{}
	for _, engine := range engines {
		if engine.Type != containerd.Type {
			continue
		}
		nerdctlNets[engine.PID] = newNerdctlNetworks(ctx, engine, allnetns)
	}
	// Now that we know about the nerdctl-managed CNI networks, try to locate
	// the matching Linux-kernel network interfaces so we can set/override the
	// alias names of the interfaces.
	for _, nerdynets := range nerdctlNets {
		for _, netw := range nerdynets.networks {
			nifname := netw.PluginField("bridge", "bridge")
			if nifname == "" {
				// silently ignore if this ain't a bridge-based network.
				continue
			}
			log.Debugf("nerdctl bridge network %q (ID %q) uses netdev %q",
				netw.Name, netw.ID, nifname)
			netif, ok := nerdynets.engineNetns.NamedNifs[nifname]
			if !ok {
				// hmm, no such Linux bridge (yet), so skip it as it won't show
				// up in the discovery anyway without a Linux bridge network
				// interface.
				continue
			}
			nif := netif.Nif()
			nif.Alias = netw.Name
			nif.Labels[GostwireNetworkNameKey] = netw.Name
			nif.AddLabels(netw.Labels)
			// TODO: support internal when --internal gets supported by nerdctl.
		}
	}
}
