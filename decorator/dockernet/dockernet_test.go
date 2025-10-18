// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/go-plugger/v3"
	lxkns "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/ipam"
	"github.com/thediveo/morbyd/v2/net"
	"github.com/thediveo/morbyd/v2/net/macvlan"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/notwork/dummy"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	testNetworkBaseName    = "gostwire-decorator-dockernet-test"
	testBridgeNetworkName  = testNetworkBaseName + "-bridge"
	testMACVLANNetworkName = testNetworkBaseName + "-macvlan"

	testWorkloadName = "gostwire-decorator-docketnet-test-workload"
)

var _ = Describe("dockernet decorator", func() {

	It("registers correctly", func() {
		Expect(plugger.Group[decorator.Decorate]().Plugins()).To(
			ContainElement("dockernet"))
	})

	Context("when looking for Docker-managed networks", func() {

		BeforeEach(func() {
			goodfds := Filedescriptors()
			goodgos := Goroutines()
			DeferCleanup(func() {
				Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})

			DeferCleanup(slog.SetDefault, slog.Default())
			slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		})

		It("discovers the name of a bridge network and decorates its Linux-kernel bridge network interface", func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By(fmt.Sprintf("creating a test Docker custom bridge network %q", testBridgeNetworkName))
			sess := Successful(morbyd.NewSession(ctx,
				session.WithAutoCleaning("test.decorator.dockernet=")))
			DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })
			testnet := Successful(sess.CreateNetwork(ctx, testBridgeNetworkName,
				net.WithInternal(),
				net.WithLabel("foo=bar")))

			By("creating a test workload and connecting it to the test network")
			testwl := Successful(sess.Run(ctx, "busybox:latest",
				run.WithName(testWorkloadName),
				run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
				run.WithNetwork(testnet.ID),
			))
			testwlPID := model.PIDType(Successful(testwl.PID(ctx)))

			By("running a discovery")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()
			allnetns, lxknsdisco := discover.Discover(ctx, cizer, nil)
			Expect(allnetns).NotTo(BeEmpty())

			Expect(lxknsdisco.Processes).To(HaveKey(testwlPID))
			testwlc := lxknsdisco.Processes[testwlPID].Container

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
			Expect(bridge).To(network.HaveInterfaceAlias(testBridgeNetworkName))
			Expect(bridge.Nif().Labels).To(HaveKeyWithValue(GostwireNetworkNameKey, testBridgeNetworkName))
			Expect(bridge.Nif().Labels).To(HaveKey(network.GostwireInternalBridgeKey))
			Expect(bridge.Nif().Labels).To(HaveKeyWithValue("foo", "bar"))
		})

		It("discovers the name of a MACVLAN network and decorates its Linkx-kernel master network interface", func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By("setting up our own isolated MACVLAN network")
			dmymaster := dummy.NewTransient()

			sess := Successful(morbyd.NewSession(ctx, session.WithAutoCleaning("test.decorator.dockernet=")))
			DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })
			testnet := Successful(sess.CreateNetwork(ctx, testMACVLANNetworkName,
				net.WithDriver("macvlan"),
				macvlan.WithParent(dmymaster.Attrs().Name),
				net.WithIPAM(ipam.WithPool("192.168.253.0/24")),
				net.WithLabel("foo=bar")))

			By(fmt.Sprintf("creating a test workload and connecting it to the test network %q", testMACVLANNetworkName))
			testwl := Successful(sess.Run(ctx, "busybox:latest",
				run.WithName(testWorkloadName),
				run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
				run.WithNetwork(testnet.ID),
				run.WithAutoRemove()))
			testwlPID := model.PIDType(Successful(testwl.PID(ctx)))

			By("running a discovery and waiting things to settle")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()

			var allnetns network.NetworkNamespaces
			var lxknsdisco *lxkns.Result
			var testwlc *model.Container
			Eventually(func() model.Namespace {
				allnetns, lxknsdisco = discover.Discover(ctx, cizer, nil)
				Expect(lxknsdisco.Processes).To(HaveKey(testwlPID))
				testwlc = lxknsdisco.Processes[testwlPID].Container
				return testwlc.Process.Namespaces[model.NetNS]
			}, "5s", "0.25s").ShouldNot(BeNil())

			// Expect eth0 inside container be a MACVLAN with a specific master.
			wlnetnsid := testwlc.Process.Namespaces[model.NetNS].ID()
			Expect(allnetns).To(HaveKey(wlnetnsid))
			wlnetns := allnetns[wlnetnsid]
			Expect(wlnetns.NamedNifs).To(network.ContainInterfaceWithName("eth0"))
			eth0 := wlnetns.NamedNifs["eth0"]
			Expect(eth0.Nif().Kind).To(Equal("macvlan"))
			macvlan, _ := eth0.(network.Macvlan)
			Expect(macvlan).NotTo(BeNil())
			Expect(macvlan.Macvlan().Master).NotTo(BeNil())

			// Expect a physical network interface with the alias and label of the test network.
			master := macvlan.Macvlan().Master
			Expect(master).To(network.HaveInterfaceName(dmymaster.Attrs().Name))
			Expect(master).To(network.HaveInterfaceAlias(testMACVLANNetworkName))
			Expect(master.Nif().Labels).To(HaveKeyWithValue(GostwireNetworkNameKey, testMACVLANNetworkName))
			Expect(master.Nif().Labels).To(HaveKeyWithValue("foo", "bar"))
		})

	})

})
