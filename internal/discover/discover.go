// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package discover

import (
	"context"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/turtlefinder"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
)

// Discover returns the discovered network stacks, virtual network topology, and
// network-related configuration. Labels optionally control certain aspects of
// "decorating" (that is, enriching) the discovery results for some decorator
// plugins supporting labels (such as the ieappicon decorator plugin).
func Discover(ctx context.Context, cizer containerizer.Containerizer, labels map[string]string) (network.NetworkNamespaces, *lxknsdiscover.Result) {
	// First phase: run a Linux-kernel namespace (+container) discovery,
	// courtesy of lxkns.
	discoverednetns := lxknsdiscover.Namespaces(
		lxknsdiscover.FromProcs(),
		lxknsdiscover.FromBindmounts(),
		lxknsdiscover.WithNamespaceTypes(
			species.CLONE_NEWNET|species.CLONE_NEWPID|species.CLONE_NEWNS|species.CLONE_NEWUTS),
		lxknsdiscover.WithHierarchy(),
		lxknsdiscover.WithContainerizer(cizer),
		lxknsdiscover.WithPIDMapper(),
		lxknsdiscover.WithLabels(labels),
	)
	// Second phase: create the Gostwire-specific information model based on the
	// lxkns discovery and augment the model with additional network-related
	// details, such as tenant DNS resolver configuration, Docker network names,
	// et cetera.
	log.Debugf("discovering network namespace details (interfaces, address, routes, ...)")
	allnetns := network.NewNetworkNamespaces(
		discoverednetns.Namespaces[model.NetNS],
		discoverednetns.Processes,
		discoverednetns.Containers)
	engines := []*model.ContainerEngine{}
	if overseer, ok := cizer.(turtlefinder.Overseer); ok {
		engines = overseer.Engines()
	}
	log.Debugf("running gostwire decorators")
	for _, decorateur := range plugger.Group[decorator.Decorate]().Symbols() {
		decorateur(ctx, allnetns, discoverednetns.Processes, engines)
	}
	log.Debugf("gostwire discovery finished")
	return allnetns, discoverednetns
}
