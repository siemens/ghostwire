// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"
	"time"

	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/vxlan"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("VXLAN network interfaces", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	It("discovers VXLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a transient network namespace with VXLAN netdev connected to initial netns lo")
		tmpNetns := netns.NewTransient()
		tmpNetnsID := species.NamespaceIDfromInode(netns.Ino(tmpNetns))
		overlay := vxlan.NewTransient(Successful(netlink.LinkByName("lo")),
			vxlan.InNamespace(tmpNetns),
			vxlan.WithID(666),
			vxlan.WithTTL(2),
			vxlan.WithDestinationPort(4789))

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(tmpNetnsID))

		By("ensuring VXLAN attributes and master relation")
		testnetns := allnetns[tmpNetnsID]
		Expect(testnetns.Nifs).To(HaveLen(2))
		for _, nif := range testnetns.Nifs {
			By(nif.Nif().Name)
		}
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceKindAndName("", "lo"),
			HaveInterfaceKindAndName("vxlan", overlay.Attrs().Name),
		))
		vxlannif := testnetns.NamedNifs[overlay.Attrs().Name]
		vxlan := AssignableTo[Vxlan](vxlannif).Vxlan()
		Expect(vxlan.Master).NotTo(BeNil())
		master := vxlan.Master.Nif()
		Expect(master.Name).To(Equal("lo"))
		Expect(vxlan.Netns).NotTo(BeIdenticalTo(master.Netns))

		Expect(vxlan.VID).To(Equal(uint32(666)))
		Expect(vxlan.DestinationPort).To(Equal(uint16(4789)))

		By("ensuring castability")
		Expect(func() {
			vxlans := master.Nif().Slaves.OfKind("vxlan")
			Expect(vxlans).NotTo(BeEmpty())
			for _, vxlan := range vxlans {
				_ = AssignableTo[Vxlan](vxlan)
				_ = AssignableTo[*VxlanAttrs](vxlan)
			}
		}).NotTo(Panic())
	})

})
