// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package nerdctlnet

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/test/nerdctl"
	"github.com/siemens/ghostwire/v2/turtlefinder"
	"github.com/siemens/ghostwire/v2/util"

	"github.com/onsi/gomega/gexec"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/containerd"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

const testNetworkName = "gostwire-test-nerdctlnet"
const testWorkloadName = "gostwire-test-nerdctlnet-workload"

var _ = Describe("nerdctlnet decorator", func() {

	It("registers correctly", func() {
		Expect(plugger.Group[decorator.Decorate]().Plugins()).To(
			ContainElement("nerdctlnet"))
	})

	Context("when looking for nerdctl-managed CNI networks", func() {

		BeforeEach(func() {
			goodfds := Filedescriptors()
			goodgos := Goroutines() // avoid other failed goroutine tests to spill over

			nerdctl.NerdctlIgnore("rm", testWorkloadName)
			nerdctl.NerdctlIgnore("network", "rm", testNetworkName)

			DeferCleanup(func() {
				gexec.KillAndWait()
				nerdctl.NerdctlIgnore("rm", testWorkloadName)
				nerdctl.NerdctlIgnore("network", "rm", testNetworkName)

				Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))

			})
		})

		It("discovers the name of a bridge network and decorates its Linux-kernel bridge network interface", NodeTimeout(30*time.Second), func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			nerdctl.SkipWithout()

			By(fmt.Sprintf("creating a test bridge network %q", testNetworkName))
			nerdctl.Nerdctl("network", "create", "--label=foo=bar", testNetworkName)

			By("creating a test workload and connecting it to the test network")
			nerdctl.NerdctlIgnore("rm", "-f", testWorkloadName)
			nerdctl.Nerdctl(
				"run", "-d",
				"--name", testWorkloadName,
				"--network", testNetworkName,
				"busybox", "/bin/sleep", "120s")

			By("running a discovery")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()
			allnetns, _ := discover.Discover(ctx, cizer, nil)
			Expect(allnetns).NotTo(BeEmpty())

			testwlc := util.FindContainer(allnetns, testWorkloadName, containerd.Type)
			Expect(testwlc).NotTo(BeNil())

			// Expect eth0 inside container to have a veth peer.
			wlnetnsid := testwlc.Process.Namespaces[model.NetNS].ID()
			Expect(allnetns).To(HaveKey(wlnetnsid))
			wlnetns := allnetns[wlnetnsid]
			Expect(wlnetns.NamedNifs).To(network.ContainInterfaceWithName("eth0"))
			eth0 := wlnetns.NamedNifs["eth0"]
			Expect(eth0.Nif().Kind).To(Equal("veth"))
			veth, _ := eth0.(network.Veth)
			Expect(veth).NotTo(BeNil())
			Expect(veth.Veth().Peer).NotTo(BeNil())

			// Expect a bridge with the alias and label of the test network.
			bridge := veth.Veth().Peer.Nif().Bridge
			Expect(bridge).NotTo(BeNil())
			Expect(bridge).To(network.HaveInterfaceAlias(testNetworkName))
			Expect(bridge.Nif().Labels).To(HaveKeyWithValue(GostwireNetworkNameKey, testNetworkName))
			Expect(bridge.Nif().Labels).To(HaveKeyWithValue("foo", "bar"))

		})

	})

})
