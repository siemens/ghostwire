// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"
	"time"

	"github.com/thediveo/deferrer"
	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

const testNsidNetnsName = "gostwire-testnsid"

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
	})

	Context("with an almost lonely network namespace", func() {

		var outer deferrer.Deferrer
		var allnetns NetworkNamespaces
		var scripts testbasher.Basher
		var realnetnsid species.NamespaceID
		var testNetns *NetworkNamespace

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			outer = deferrer.Deferrer{}

			scripts = testbasher.Basher{}
			outer.Defer(scripts.Done)

			scripts.Common(nstest.NamespaceUtilsScript)
			scripts.Common("netnsname=" + testNsidNetnsName)
			scripts.Script("main", `
ip netns del ${netnsname} || true
ip link del wallace || true
ip netns add ${netnsname}
ip link add wallace type veth peer gromit
ip link set gromit netns ${netnsname}
namespaceid /run/netns/${netnsname}
read # wait for test to proceed
ip netns del ${netnsname}
`)
			cmd := scripts.Start("main")
			outer.Defer(cmd.Close)

			realnetnsid = nstest.CmdDecodeNSId(cmd)
			testnetnsid, err := ops.NamespacePath("/proc/1/root/run/netns/" + testNsidNetnsName).ID()
			Expect(err).NotTo(HaveOccurred())
			Expect(testnetnsid).To(Equal(realnetnsid))

			allnetns, _ = discoverRedux()
			Expect(allnetns).To(HaveKey(realnetnsid))

			testNetns = allnetns[realnetnsid]
			Expect(testNetns).NotTo(BeNil())
		})

		AfterEach(outer.Cleanup)

		It("founds the related NetworkNamespace's via their discovered NSIDs", func() {
			initnetnsid, err := ops.NamespacePath("/proc/1/ns/net").ID()
			Expect(err).NotTo(HaveOccurred())
			initialNetns := allnetns[initnetnsid]
			Expect(initialNetns).NotTo(BeNil())

			Expect(initialNetns.peerNetns).To(ContainElement(testNetns))
			Expect(testNetns.peerNetns).To(ContainElement(initialNetns))
		})

		It("lists nifs in new network namespace", func() {
			Expect(testNetns.NifList()).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName("gromit")))
		})

		When("sorting", func() {

			It("sorts the initial netns first", func() {
				initnetnsid, err := ops.NamespacePath("/proc/1/ns/net").ID()
				Expect(err).NotTo(HaveOccurred())
				initialNetns := allnetns[initnetnsid]
				Expect(initialNetns).NotTo(BeNil())

				var anotherNetns *NetworkNamespace
				for _, netns := range allnetns {
					if netns == initialNetns {
						continue
					}
					anotherNetns = netns
					break
				}
				Expect(anotherNetns).NotTo(BeNil())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, anotherNetns})(0, 1)).To(BeTrue())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{anotherNetns, initialNetns})(0, 1)).To(BeFalse())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, initialNetns})(0, 1)).To(BeFalse())
			})

		})

	})

})
