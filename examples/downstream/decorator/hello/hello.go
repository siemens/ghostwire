// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package hello

import (
	"context"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
)

// HelloLabel is the label added to all discovered containers.
const HelloLabel = "example.org/hello"

func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("example-hello"))
}

// Decorate labels all containers of all discovered container engines.
func Decorate(
	_ context.Context,
	_ network.NetworkNamespaces,
	_ model.ProcessTable,
	engines []*model.ContainerEngine,
) {
	for _, engine := range engines {
		for _, container := range engine.Containers {
			if container.Labels == nil {
				container.Labels = model.Labels{}
			}
			container.Labels[HelloLabel] = "world"
		}
	}
}
