// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package gostwire

import (
	"context"
	"os"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("discovery", func() {

	It("discovers our own network namespace, process, and container engines", func(ctx context.Context) {
		cizerctx, cancel := context.WithCancel(ctx)
		DeferCleanup(cancel)
		cizer := turtlefinder.New(func() context.Context { return cizerctx })
		DeferCleanup(cizer.Close)

		result := Discover(ctx, cizer, nil)
		Expect(result.Lxkns).NotTo(BeNil())
		Expect(result.Netns).To(HaveKey(Successful(ops.NamespacePath("/proc/self/ns/net").ID())))
		Expect(result.Lxkns.Processes).To(HaveKey(model.PIDType(os.Getpid())))
		Expect(result.Engines).To(Equal(result.Lxkns.ContainerEngines))
	})

})
