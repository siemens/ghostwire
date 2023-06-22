// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockernet

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/turtlefinder"

	"github.com/ory/dockertest/v3"
	dtclient "github.com/ory/dockertest/v3/docker"
	"github.com/thediveo/go-plugger/v3"
	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

const testNetworkName = "gostwire-test-dockernet"
const testWorkloadName = "gostwire-test-dockernet-workload"

var _ = Describe("dockernet decorator", func() {

	It("registers correctly", func() {
		Expect(plugger.Group[decorator.Decorate]().Plugins()).To(
			ContainElement("dockernet"))
	})

	Context("when looking for Docker-managed networks", func() {

		// We make the test more resilient against left-over created containers by
		// dockertest by first cleaning up before each test. Kind of brute force
		// method, me ponders.
		BeforeEach(func() {
			goodfds := Filedescriptors()
			goodgos := Goroutines()
			pool, err := dockertest.NewPool("")
			if err == nil {
				_ = pool.RemoveContainerByName(testWorkloadName)
			}
			DeferCleanup(func() {
				// dockertest has the slightly annoying behavior of leaving us with created
				// containers when it fails to connect them to a network. So we simply clean
				// up here, kind of brute force.
				_ = pool.RemoveContainerByName(testWorkloadName)
				pool.Client.HTTPClient.CloseIdleConnections()
				Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})
		})

		It("discovers the name of a bridge network and decorates its Linux-kernel bridge network interface", NodeTimeout(30*time.Second), func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By(fmt.Sprintf("creating a test bridge network %q", testNetworkName))
			pool, err := dockertest.NewPool("")
			Expect(err).NotTo(HaveOccurred())
			testnet, err := pool.CreateNetwork(testNetworkName, func(config *dtclient.CreateNetworkOptions) {
				config.Internal = true
				config.Labels = map[string]string{"foo": "bar"}
			})
			Expect(err).NotTo(HaveOccurred(), "bridge network %s", testNetworkName)
			defer testnet.Close()

			By("creating a test workload and connecting it to the test network")
			testwl, err := pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       testWorkloadName,
				Cmd:        []string{"/bin/sleep", "120s"},
				Networks:   []*dockertest.Network{testnet},
			})
			Expect(err).NotTo(HaveOccurred(), "container %", testWorkloadName)
			defer testwl.Close()

			By("running a discovery")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()
			allnetns, lxknsdisco := discover.Discover(ctx, cizer, nil)
			Expect(allnetns).NotTo(BeEmpty())

			testwlpid := model.PIDType(testwl.Container.State.Pid)
			Expect(lxknsdisco.Processes).To(HaveKey(testwlpid))
			testwlc := lxknsdisco.Processes[testwlpid].Container

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
			Expect(bridge.Nif().Labels).To(HaveKey(network.GostwireInternalBridgeKey))
			Expect(bridge.Nif().Labels).To(HaveKeyWithValue("foo", "bar"))
		})

		It("discovers the name of a MACVLAN network and decorates its Linkx-kernel master network interface", NodeTimeout(30*time.Second), func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By("looking for any existing MACVLAN network")
			pool, err := dockertest.NewPool("")
			Expect(err).NotTo(HaveOccurred())
			dockernets, err := pool.Client.ListNetworks()
			Expect(err).NotTo(HaveOccurred())
			var macvlanNetworkName string
			var masterName string
			for _, dcnet := range dockernets {
				if dcnet.Driver == "macvlan" {
					macvlanNetworkName = dcnet.Name
					masterName = dcnet.Options["parent"]
					break
				}
			}
			if macvlanNetworkName == "" {
				By("looking for a suitable master", func() {
					ctx, cancel := context.WithCancel(ctx)
					cizer := turtlefinder.New(func() context.Context { return ctx })
					defer cancel()
					defer cizer.Close()
					allnetns, lxknsdisco := discover.Discover(ctx, cizer, nil)
					Expect(allnetns).NotTo(BeEmpty())
					initialnetns := lxknsdisco.Processes[model.PIDType(1)].Namespaces[model.NetNS]
					Expect(allnetns).To(HaveKey(initialnetns.ID()))
					inetns := allnetns[initialnetns.ID()]
					for _, nif := range inetns.Nifs {
						if nif.Nif().Physical && nif.Nif().Name != "lo" {
							masterName = nif.Nif().Name
							break
						}
					}
					Expect(masterName).NotTo(BeEmpty(), "found no physical network interface in initial network namespace")
				})

				By(fmt.Sprintf("creating a test macvlan network %q with parent %q", testNetworkName, masterName))
				testnet, err := pool.CreateNetwork(testNetworkName, func(config *dtclient.CreateNetworkOptions) {
					config.Driver = "macvlan"
					config.Options = map[string]interface{}{
						"parent": masterName,
					}
					config.IPAM = &dtclient.IPAMOptions{
						Config: []dtclient.IPAMConfig{
							{Subnet: "192.168.253.0/24"},
						},
					}
					config.Labels = map[string]string{"foo": "bar"}
				})
				Expect(err).NotTo(HaveOccurred(), "MACVLAN network %s", testNetworkName)
				defer testnet.Close()
				macvlanNetworkName = testNetworkName
			}

			By(fmt.Sprintf("creating a test workload and connecting it to the test network %q", macvlanNetworkName))
			testwl, err := pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       testWorkloadName,
				Cmd:        []string{"/bin/sleep", "120s"},
				NetworkID:  macvlanNetworkName,
			})
			Expect(err).NotTo(HaveOccurred(), "container %s", testWorkloadName)
			defer testwl.Close()

			By("running a discovery and waiting things to settle")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()

			var allnetns network.NetworkNamespaces
			var lxknsdisco *lxknsdiscover.Result
			var testwlc *model.Container
			Eventually(func() model.Namespace {
				allnetns, lxknsdisco = discover.Discover(ctx, cizer, nil)
				testwlpid := model.PIDType(testwl.Container.State.Pid)
				Expect(lxknsdisco.Processes).To(HaveKey(testwlpid))
				testwlc = lxknsdisco.Processes[testwlpid].Container
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
			Expect(master).To(network.HaveInterfaceName(masterName))
			Expect(master).To(network.HaveInterfaceAlias(macvlanNetworkName))
			Expect(master.Nif().Labels).To(HaveKeyWithValue(GostwireNetworkNameKey, macvlanNetworkName))
			if macvlanNetworkName == testNetworkName {
				Expect(master.Nif().Labels).To(HaveKeyWithValue("foo", "bar"))
			}
		})

	})

})
