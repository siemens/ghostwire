// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package gostwire

import (
	"context"

	_ "github.com/siemens/ghostwire/v2/decorator/all" // activate all Gostwire-specific decorators.
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/turtlefinder"
	"github.com/thediveo/lxkns/containerizer"
	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
)

// DiscoveryResult contains the network topology and configuration discovery
// results, as well as the Linux-kernel namespace, process, and container
// discovery results.
type DiscoveryResult struct {
	Netns   network.NetworkNamespaces // network discovery
	Lxkns   *lxknsdiscover.Result     // namespaces, process, and containers discovery
	Engines []*model.ContainerEngine  // discovered container engines, even if without workload
}

// Discover returns the discovered network stacks, virtual network topology, and
// network-related configuration. Labels optionally control certain aspects of
// "decorating" (that is, enriching) the discovery results for some decorator
// plugins supporting labels (such as the ieappicon decorator plugin).
func Discover(ctx context.Context, cizer containerizer.Containerizer, labels map[string]string) DiscoveryResult {
	// break the vicious import cycle which otherwise happens for some unit test
	// needing discovery.
	allnetns, nsdisco := discover.Discover(ctx, cizer, labels)
	var engines []*model.ContainerEngine
	if overseer, ok := cizer.(turtlefinder.Overseer); ok {
		engines = overseer.Engines()
	}
	return DiscoveryResult{Netns: allnetns, Lxkns: nsdisco, Engines: engines}
}
