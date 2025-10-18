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
	"github.com/thediveo/notwork/vlan"
	"github.com/thediveo/spacetest/netns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("VLAN network interfaces", func() {

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

	It("discovers VLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a transient network namespace with VLAN netdev connected to initial netns lo")
		dmyNetns := netns.NewTransient()
		dmy := dummy.NewTransient(dummy.InNamespace(dmyNetns))

		tmpNetns := netns.NewTransient()
		tmpNetnsID := species.NamespaceIDfromInode(netns.Ino(tmpNetns))
		wielahm := vlan.NewTransient(999, dmy,
			vlan.InNamespace(tmpNetns),
			vlan.WithLinkNamespace(dmyNetns),
			vlan.WithLooseBinding())

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(tmpNetnsID))

		By("ensuring VLAN attributes and master relation")
		testnetns := allnetns[tmpNetnsID]
		Expect(testnetns.Nifs).To(HaveLen(2 /*lo+vlan*/), testnetns.NifsString())
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceKindAndName("", "lo"),
			HaveInterfaceKindAndName("vlan", wielahm.Attrs().Name),
		), testnetns.NifsString())
		vlannif := testnetns.NamedNifs[wielahm.Attrs().Name]
		vlan := AssignableTo[Vlan](vlannif).Vlan()

		Expect(vlan.Master).NotTo(BeNil())
		master := vlan.Master.Nif()
		Expect(master.Name).To(Equal(dmy.Attrs().Name))
		Expect(vlan.Netns).NotTo(BeIdenticalTo(master.Netns))

		By("ensuring castability")
		vlans := master.Nif().Slaves.OfKind("vlan")
		Expect(vlans).NotTo(BeEmpty())
		Expect(func() {
			for _, slave := range vlans {
				_ = AssignableTo[Vlan](slave)
				_ = AssignableTo[*VlanAttrs](slave)
			}
		}).NotTo(Panic())
	})

})
