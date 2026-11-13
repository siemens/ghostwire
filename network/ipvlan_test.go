// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"log/slog"
	"os"
	"time"

	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("IPVLAN network interfaces", func() {

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

	It("discovers IPVLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a dummy netdev in a transient network namespace")
		masterNetns := netns.NewTransient()
		masterNetnsID := species.NamespaceIDfromInode(netns.Ino(masterNetns))
		dummy := dummy.NewTransient(dummy.InNamespace(masterNetns))

		By("creating an IPVLAN netdev in a different transient network namespace")
		ipvlanNetns := netns.NewTransient()
		ipvlanNetnsID := species.NamespaceIDfromInode(netns.Ino(ipvlanNetns))
		ipvl := link.NewTransient(&netlink.IPVlan{
			LinkAttrs: netlink.LinkAttrs{
				ParentIndex: dummy.Attrs().Index,
			},
			Mode: netlink.IPVLAN_MODE_L3,
			Flag: netlink.IPVLAN_FLAG_PRIVATE,
		}, "ipvl-",
			link.InNamespace(ipvlanNetns),
			link.WithLinkNamespace(masterNetns))

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(masterNetnsID))
		Expect(allnetns).To(HaveKey(ipvlanNetnsID))

		By("ensuring IPVLAN attributes and master relation")
		testnetns := allnetns[ipvlanNetnsID]
		Expect(testnetns.Nifs).To(HaveLen(2), testnetns.NifsString())
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceKindAndName("", "lo"),
			HaveInterfaceKindAndName("ipvlan", ipvl.Attrs().Name),
		), testnetns.NifsString())
		ipvlan := AssignableTo[Ipvlan](testnetns.NamedNifs[ipvl.Attrs().Name]).Ipvlan()

		Expect(ipvlan.Mode.String()).To(Equal("L3"))
		Expect(ipvlan.Flag.String()).To(Equal("private"))

		Expect(ipvlan.Master).NotTo(BeNil())
		master := ipvlan.Master.Nif()
		Expect(master.Name).To(Equal(dummy.Attrs().Name))
		Expect(master.Netns).NotTo(BeIdenticalTo(ipvlan.Netns))

		By("ensuring castability")
		ipvlans := master.Slaves.OfKind("ipvlan")
		Expect(ipvlans).To(HaveLen(1))
		_ = AssignableTo[Ipvlan](ipvlans[0])
		_ = AssignableTo[*IpvlanAttrs](ipvlans[0])
	})

})
