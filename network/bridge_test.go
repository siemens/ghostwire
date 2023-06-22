// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

const testBridgeNetworkName = "gostwire-test-bridge"
const testBridgeWorkloadName = "gostwire-test-bridge-workload"

func discoverRedux() (NetworkNamespaces, *discover.Result) {
	discoverednetns := discover.Namespaces(
		discover.FromProcs(),
		discover.FromBindmounts(),
		discover.WithNamespaceTypes(
			species.CLONE_NEWNET|species.CLONE_NEWPID|species.CLONE_NEWNS|species.CLONE_NEWUTS),
		discover.WithHierarchy(),
		discover.WithPIDMapper(),
	)
	allnetns := NewNetworkNamespaces(
		discoverednetns.Namespaces[model.NetNS],
		discoverednetns.Processes,
		discoverednetns.Containers)
	return allnetns, discoverednetns
}

var _ = Describe("bridge nif", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	It("discovers bridge", NodeTimeout(30*time.Second), func(_ context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a test bridge network")
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		testnet, err := pool.CreateNetwork(testBridgeNetworkName)
		Expect(err).NotTo(HaveOccurred(), "network %s", testBridgeNetworkName)
		defer testnet.Close()

		By("creating a test workload and connecting it to the test network")
		testwl, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       testBridgeWorkloadName,
			Cmd:        []string{"/bin/sleep", "120s"},
			Networks:   []*dockertest.Network{testnet},
		})
		Expect(err).NotTo(HaveOccurred(), "container %s", testBridgeWorkloadName)
		defer testwl.Close()

		By("running a discovery")
		allnetns, lxknsdisco := discoverRedux()
		Expect(allnetns).NotTo(BeEmpty())

		// Expect a bridge to be present, with a nif name derived from the
		// network ID.
		brname := "br-" + testnet.Network.ID[0:12]
		hostnetnsid := lxknsdisco.Processes[1].Namespaces[model.NetNS].ID()
		hostnetns := allnetns[hostnetnsid]
		Expect(hostnetns).NotTo(BeNil())
		Expect(hostnetns.NamedNifs).To(HaveKey(brname))
		nif := hostnetns.NamedNifs[brname]
		Expect(nif.Nif().Kind).To(Equal("bridge"))

		// Expect a single port: an veth. We leave more precise checks up to the
		// VETH unit tests.
		ports := nif.(Bridge).Bridge().Ports
		Expect(ports).To(HaveLen(1))
		port := ports[0]
		Expect(port.Nif().Kind).To(Equal("veth"))
	})

})
