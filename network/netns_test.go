// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"log/slog"
	"maps"
	"os"
	"time"

	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/nonstd/xiter"
	"github.com/thediveo/notwork/veth"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("network namespace", func() {

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

	Context("with an almost lonely network namespace", func() {

		var allnetns NetworkNamespaces
		var testNetns *NetworkNamespace
		var gromit netlink.Link

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By("creating a new transient network namespace and a VETH pair from here to there")
			tmpNetns := netns.NewTransient()
			_, gromit = veth.NewTransient(veth.WithPeerNamespace(tmpNetns))

			allnetns, _ = discoverRedux()
			Expect(allnetns).To(HaveKey(species.NamespaceIDfromInode(netns.Ino(tmpNetns))))

			testNetns = allnetns[species.NamespaceIDfromInode(netns.Ino(tmpNetns))]
			Expect(testNetns).NotTo(BeNil())
		})

		It("found the related NetworkNamespace via their discovered NSIDs", func() {
			initnetnsid, err := ops.NamespacePath("/proc/self/ns/net").ID()
			Expect(err).NotTo(HaveOccurred())
			initialNetns := allnetns[initnetnsid]
			Expect(initialNetns).NotTo(BeNil())

			Expect(initialNetns.peerNetns).To(ContainElement(testNetns))
			Expect(testNetns.peerNetns).To(ContainElement(initialNetns))
		})

		It("lists nifs in new network namespace", func() {
			Expect(testNetns.NifList()).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName(gromit.Attrs().Name)))
		})

		When("sorting", func() {

			It("sorts the initial netns first", func() {
				initnetnsid, err := ops.NamespacePath("/proc/1/ns/net").ID()
				Expect(err).NotTo(HaveOccurred())
				initialNetns := allnetns[initnetnsid]
				Expect(initialNetns).NotTo(BeNil())

				anotherNetns := Allright(
					xiter.FirstOk(
						xiter.Filter(maps.Values(allnetns),
							func(n *NetworkNamespace) bool { return n != initialNetns })))
				Expect(anotherNetns).NotTo(BeNil())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, anotherNetns})(0, 1)).To(BeTrue())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{anotherNetns, initialNetns})(0, 1)).To(BeFalse())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, initialNetns})(0, 1)).To(BeFalse())
			})

		})

	})

})
