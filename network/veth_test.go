// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/bridge"
	"github.com/thediveo/notwork/veth"
	"github.com/thediveo/spacetest/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

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

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	It("discovers VETH pairs", func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a test bridge and VETH pair")
		dupondNetns := netns.NewTransient()
		dupontNetns := netns.NewTransient()
		br := bridge.NewTransient(bridge.InNamespace(dupondNetns))
		dupond, dupont := veth.NewTransient(veth.InNamespace(dupondNetns), veth.WithPeerNamespace(dupontNetns))
		netns.Execute(dupondNetns, func() {
			bridge.AddPort(br, dupond)
		})

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).NotTo(BeEmpty())

		// Expect a bridge to be present, with a nif name derived from the
		// network ID.
		brNetnsID := species.NamespaceIDfromInode(netns.Ino(dupondNetns))
		brnetns := allnetns[brNetnsID]
		Expect(brnetns).NotTo(BeNil())
		Expect(brnetns.NamedNifs).To(HaveKey(br.Attrs().Name))
		nif := brnetns.NamedNifs[br.Attrs().Name]
		Expect(nif.Nif().Kind).To(Equal("bridge"))

		// Expect a single port: an veth linked with another end in our test
		// container.
		ports := nif.(Bridge).Bridge().Ports
		Expect(ports).To(HaveLen(1))

		// Now for the checks centrol to the VETH peer relation: do we correctly
		// got two VETHs and do they refer to each other?
		Expect(ports[0].Nif().Kind).To(Equal("veth"))
		Expect(ports[0].Nif().Name).To(Equal(dupond.Attrs().Name))

		veth := AssignableTo[Veth](ports[0])
		peer := veth.Veth().Peer
		Expect(peer.Nif().Kind).To(Equal("veth"))
		Expect(peer.Nif().Name).To(Equal(dupont.Attrs().Name))

		// Our peer's peer must be us.
		vethpeer := AssignableTo[Veth](peer)
		Expect(vethpeer.Veth().Peer).To(BeIdenticalTo(veth))
	})

})
