// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"errors"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/netkit"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("netkit network interfaces", func() {

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

	It("discovers netkit pairs", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a netkit pair spanning two transient network namespaces")
		primaryNetns := netns.NewTransient()
		primaryNetnsID := species.NamespaceIDfromInode(netns.Ino(primaryNetns))
		var probeErr error
		netns.Execute(primaryNetns, func() {
			probeErr = netlink.LinkAdd(&netlink.Netkit{LinkAttrs: netlink.LinkAttrs{Name: "nkprobe"}})
		})
		if errors.Is(probeErr, unix.EOPNOTSUPP) {
			Skip("kernel lacks netkit support")
		}
		peerNetns := netns.NewTransient()
		peerNetnsID := species.NamespaceIDfromInode(netns.Ino(peerNetns))
		primary, peer := netkit.NewTransient(
			netkit.InNamespace(primaryNetns),
			netkit.WithPeerNamespace(peerNetns),
			netkit.WithIPMode(),
			netkit.WithPeerPolicy(netlink.NETKIT_POLICY_BLACKHOLE))

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(primaryNetnsID))
		Expect(allnetns).To(HaveKey(peerNetnsID))

		By("ensuring netkit attributes and peer relation")
		primaryNif := AssignableTo[Netkit](allnetns[primaryNetnsID].NamedNifs[primary.Attrs().Name])
		Expect(primaryNif.Nif().Kind).To(Equal("netkit"))
		nk := primaryNif.Netkit()
		Expect(nk.Mode.String()).To(Equal("L3"))
		Expect(nk.Primary).To(BeTrue())
		Expect(nk.Policy.String()).To(Equal("forward"))
		Expect(nk.PeerPolicy.String()).To(Equal("blackhole"))

		peerNif := AssignableTo[Netkit](allnetns[peerNetnsID].NamedNifs[peer.Attrs().Name])
		Expect(peerNif.Netkit().Primary).To(BeFalse())
		Expect(peerNif.Netkit().Policy.String()).To(Equal("blackhole"))

		Expect(nk.Peer).To(BeIdenticalTo(peerNif))
		Expect(peerNif.Netkit().Peer).To(BeIdenticalTo(primaryNif))
		Expect(PeerOf(primaryNif)).To(BeIdenticalTo(peerNif))
		Expect(nk.Single).To(BeFalse())
		Expect(peerNif.Netkit().Single).To(BeFalse())
	})

	When("not depending on kernel netkit support", func() {

		var netns *NetworkNamespace

		// netkit returns a fake netkit nif with the specified ifindex and peer
		// ifindex in our netns.
		netkit := func(index, peerindex int) *NetkitAttrs {
			nk := &NetkitAttrs{NifAttrs: NifAttrs{
				Name:  "nk" + strconv.Itoa(index),
				Index: index,
				Netns: netns,
				Link: &netlink.Netkit{LinkAttrs: netlink.LinkAttrs{
					Index: index, ParentIndex: peerindex, NetNsID: -1,
				}},
			}}
			netns.Nifs[index] = nk
			return nk
		}

		BeforeEach(func() {
			netns = ownNetworkNamespace()
		})

		It("initializes from netkit link details", func() {
			nlh := Successful(netlink.NewHandle())
			DeferCleanup(func() { _ = nlh.Close() })
			nif := NewInterface(nlh, netns, &netlink.Netkit{
				LinkAttrs:  netlink.LinkAttrs{Name: "nk-fake", Index: 4242, ParentIndex: 4243, NetNsID: -1},
				Mode:       netlink.NETKIT_MODE_L2,
				Policy:     netlink.NETKIT_POLICY_FORWARD,
				PeerPolicy: netlink.NETKIT_POLICY_BLACKHOLE,
			})
			nk := AssignableTo[Netkit](nif).Netkit()
			Expect(nk.Nif()).To(BeIdenticalTo(&nk.NifAttrs))
			Expect(nk.Kind).To(Equal("netkit"))
			Expect(nk.Name).To(Equal("nk-fake"))
			Expect(nk.Mode.String()).To(Equal("L2"))
			Expect(nk.Policy.String()).To(Equal("forward"))
			Expect(nk.PeerPolicy.String()).To(Equal("blackhole"))
			Expect(nk.Primary).To(BeFalse())
			Expect(nk.Single).To(BeFalse())
		})

		It("relates the ends of a netkit pair", func() {
			nk1 := netkit(1, 2)
			nk2 := netkit(2, 1)
			allns := NetworkNamespaces{netns.ID(): netns}

			nk1.ResolveRelations(allns)
			nk2.ResolveRelations(allns)
			Expect(nk1.Peer).To(BeIdenticalTo(nk2))
			Expect(nk2.Peer).To(BeIdenticalTo(nk1))
		})

		It("ignores a peer that isn't a netkit", func() {
			nk := netkit(1, 2)
			netns.Nifs[2] = &NifAttrs{Index: 2, Netns: netns, Link: &netlink.Dummy{}}

			nk.ResolveRelations(NetworkNamespaces{netns.ID(): netns})
			Expect(nk.Peer).To(BeNil())
		})

		It("doesn't overwrite an inconsistent peer of its peer", func() {
			nk1 := netkit(1, 2)
			nk2 := netkit(2, 1)
			nk3 := netkit(3, 2)
			nk2.Peer = nk3

			nk1.ResolveRelations(NetworkNamespaces{netns.ID(): netns})
			Expect(nk1.Peer).To(BeIdenticalTo(nk2))
			Expect(nk2.Peer).To(BeIdenticalTo(nk3))
		})

	})

	DescribeTable("detecting single-device mode",
		func(primary bool, parentIndex int, netnsid int, expected bool) {
			Expect(isSingleNetkit(primary, &netlink.LinkAttrs{
				ParentIndex: parentIndex,
				NetNsID:     netnsid,
			})).To(Equal(expected))
		},
		Entry("single device", true, 0, -1, true),
		Entry("primary with peer in same netns", true, 42, -1, false),
		Entry("primary with peer in other netns", true, 42, 1, false),
		Entry("primary with same-ifindex peer in other netns", true, 0, 1, false),
		Entry("non-primary peer", false, 0, -1, false),
	)

})
