// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

const testVethNetworkName = "gostwire-test-veth"
const testVethWorkloadName = "gostwire-test-veth-workload"

var _ = Describe("VETH nif", func() {

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

	It("discovers VETH pairs", NodeTimeout(30*time.Second), func(_ context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a test bridge network in order to get VETHs")
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		testnet, err := pool.CreateNetwork(testVethNetworkName)
		Expect(err).NotTo(HaveOccurred(), "network %s", testVethNetworkName)
		defer testnet.Close()

		By("creating a test workload and connecting it to the test network")
		testwl, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       testVethWorkloadName,
			Cmd:        []string{"/bin/sleep", "120s"},
			Networks:   []*dockertest.Network{testnet},
		})
		Expect(err).NotTo(HaveOccurred(), "container %s", testVethWorkloadName)
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

		// Expect a single port: an veth linked with another end in our test
		// container.
		ports := nif.(Bridge).Bridge().Ports
		Expect(ports).To(HaveLen(1))

		// Now for the checks centrol to the VETH peer relation: do we correctly
		// got two VETHs and do they refer to each other?
		Expect(ports[0].Nif().Kind).To(Equal("veth"))
		var veth Veth
		Expect(func() { veth = ports[0].(Veth) }).NotTo(Panic())

		peer := veth.Veth().Peer
		Expect(peer.Nif().Kind).To(Equal("veth"))
		var vethpeer Veth
		Expect(func() { vethpeer = peer.(Veth) }).NotTo(Panic())

		// Our peer's peer must be us.
		Expect(vethpeer.Veth().Peer).To(BeIdenticalTo(veth))
	})

})
