// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"log/slog"
	"os"
	"time"

	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("MACVAN network interfaces", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	It("discovers MACVLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a dummy netdev in a transient network namespace")
		masterNetns := netns.NewTransient()
		masterNetnsID := species.NamespaceIDfromInode(netns.Ino(masterNetns))
		dummy := dummy.NewTransient(dummy.InNamespace(masterNetns))

		By("creating a MACVLAN netdev in a different transient network namespace")
		macvlanNetns := netns.NewTransient()
		macvlanNetnsID := species.NamespaceIDfromInode(netns.Ino(macvlanNetns))
		mcvlan := macvlan.NewTransient(dummy,
			macvlan.InNamespace(macvlanNetns),
			macvlan.WithLinkNamespace(masterNetns), // ouch.
		)

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(masterNetnsID))
		Expect(allnetns).To(HaveKey(macvlanNetnsID))

		By("ensuring MACVLAN attributes and master relation")
		testnetns := allnetns[macvlanNetnsID]
		Expect(testnetns.Nifs).To(HaveLen(2), testnetns.NifsString())
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceKindAndName("", "lo"),
			HaveInterfaceKindAndName("macvlan", mcvlan.Attrs().Name),
		), testnetns.NifsString())
		macvlannif := testnetns.NamedNifs[mcvlan.Attrs().Name]
		macvlan := macvlannif.(Macvlan).Macvlan()

		Expect(macvlan.Macvlan().Mode.String()).To(Equal("bridge"))

		Expect(macvlan.Master).NotTo(BeNil())
		master := macvlan.Master.Nif()
		Expect(master).NotTo(BeNil())
		Expect(master.Name).To(Equal(dummy.Attrs().Name))
		Expect(master.Netns).NotTo(BeIdenticalTo(macvlan.Netns))

		By("ensuring castability")
		macvlans := master.Nif().Slaves.OfKind("macvlan")
		Expect(macvlans).NotTo(BeEmpty())
		Expect(func() {
			for _, slave := range macvlans {
				_ = AssignableTo[Macvlan](slave)
				_ = AssignableTo[*MacvlanAttrs](slave)
			}
		}).NotTo(Panic())
	})

})
