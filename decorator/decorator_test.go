// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package decorator

import (
	"context"

	"github.com/siemens/ghostwire/v2/network"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// We don't rely on any for-production decorator plugins, but instead bring in
// our own testing decorator.
const testDecoratorPluginName = "testdecoratore"

func init() {
	plugger.Group[Decorate]().Register(
		func(ctx context.Context, allnetns network.NetworkNamespaces, allprocs model.ProcessTable, engines []*model.ContainerEngine) {
			allnetns[species.NoneID] = nil // canary
		}, plugger.WithPlugin(testDecoratorPluginName))
}

var _ = Describe("decorator plugin infrastructure", func() {

	It("returns the names of registered decorators", func() {
		Expect(plugger.Group[Decorate]().Plugins()).To(ConsistOf(testDecoratorPluginName))
	})

})
