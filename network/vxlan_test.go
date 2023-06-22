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

const testVxlanNetnsName = "gostwire-testvxlan"
const testVxlanNifName = "gwtestvxlan"

var _ = Describe("VXLAN network interfaces", func() {

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

	It("discovers VXLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		outer := deferrer.Deferrer{}
		defer outer.Cleanup()

		var allnetns NetworkNamespaces
		var scripts testbasher.Basher
		var realnetnsid species.NamespaceID

		By("creating a bind-mounted network namespace with VXLAN connected to initial netns lo", func() {
			scripts = testbasher.Basher{}
			outer.Defer(scripts.Done)

			scripts.Common(nstest.NamespaceUtilsScript)
			scripts.Common("netnsname=" + testVxlanNetnsName)
			scripts.Common("testvxlannif=" + testVxlanNifName)
			scripts.Script("main", `
ip netns del ${netnsname} || true
ip netns add ${netnsname}
ip link add ${testvxlannif} type vxlan id 666 dstport 4789 dev lo ttl 2
ip link set ${testvxlannif} netns ${netnsname} 
namespaceid /run/netns/${netnsname}
read # wait for test to proceed
ip netns del ${netnsname}
`)
			cmd := scripts.Start("main")
			outer.Defer(cmd.Close)

			realnetnsid = nstest.CmdDecodeNSId(cmd)
			testnetnsid, err := ops.NamespacePath("/proc/1/root/run/netns/" + testVxlanNetnsName).ID()
			Expect(err).NotTo(HaveOccurred())
			Expect(testnetnsid).To(Equal(realnetnsid))
		})

		By("running a discovery", func() {
			allnetns, _ = discoverRedux()
			Expect(allnetns).To(HaveKey(realnetnsid))
		})

		var master *NifAttrs

		By("ensuring VXLAN attributes and master relation", func() {
			testnetns := allnetns[realnetnsid]
			Expect(testnetns.Nifs).To(HaveLen(2))
			for _, nif := range testnetns.Nifs {
				By(nif.Nif().Name)
			}
			Expect(testnetns.Nifs).To(ContainElements(
				HaveInterfaceOfKindWithName("", "lo"),
				HaveInterfaceOfKindWithName("vxlan", testVxlanNifName),
			))
			vxlannif := testnetns.NamedNifs[testVxlanNifName]
			vxlan := vxlannif.(Vxlan).Vxlan()
			Expect(vxlan.Master).NotTo(BeNil())
			master = vxlan.Master.Nif()
			Expect(master.Name).To(Equal("lo"))
			Expect(vxlan.Netns).NotTo(BeIdenticalTo(master.Netns))

			Expect(vxlan.VID).To(Equal(uint32(666)))
			Expect(vxlan.DestinationPort).To(Equal(uint16(4789)))
		})

		By("ensuring castability", func() {
			Expect(func() {
				vxlans := master.Nif().Slaves.OfKind("vxlan")
				Expect(vxlans).NotTo(BeEmpty())
				for _, vxlan := range vxlans {
					_ = vxlan.(Vxlan)
					_ = vxlan.(*VxlanAttrs)
				}
			}).NotTo(Panic())

		})
	})

})
