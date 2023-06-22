// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"fmt"
	"os"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
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

const testVlanNetnsName = "gostwire-testvlan"
const testVlanNifName = "gwtestvlan"
const testVlanID = 666

var _ = Describe("VLAN network interfaces", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	It("discovers VLAN correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		var allnetns NetworkNamespaces
		var disco *discover.Result
		var masternif Interface

		By("finding a suitable hardware network interface for new VLAN")
		allnetns, disco = discoverRedux()
		initialnetns := disco.Processes[model.PIDType(1)].Namespaces[model.NetNS]
		for _, nif := range allnetns[initialnetns.ID()].Nifs {
			if nif.Nif().Physical && nif.Nif().State == Up {
				masternif = nif
				break
			}
		}
		Expect(masternif).NotTo(BeNil())

		var scripts testbasher.Basher
		var realnetnsid species.NamespaceID

		By(fmt.Sprintf("creating a bind-mounted network namespace with VLAN connected to initial netns nif %s", masternif.Nif().Name))
		scripts = testbasher.Basher{}
		defer scripts.Done()

		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Common("masternif=" + masternif.Nif().Name)
		scripts.Common("netnsname=" + testVlanNetnsName)
		scripts.Common("testvlannif=" + testVlanNifName)
		scripts.Common(fmt.Sprintf("testvlanid=%d", testVlanID))
		scripts.Script("main", `
ip netns del ${netnsname} || true
ip netns add ${netnsname}
ip link add ${testvlannif} link ${masternif} type vlan id ${testvlanid}
ip link set ${testvlannif} netns ${netnsname}
namespaceid /run/netns/${netnsname}
read # wait for test to proceed
ip netns del ${netnsname}
`)
		cmd := scripts.Start("main")
		defer cmd.Close()

		realnetnsid = nstest.CmdDecodeNSId(cmd)
		testnetnsid, err := ops.NamespacePath("/proc/1/root/run/netns/" + testVlanNetnsName).ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(testnetnsid).To(Equal(realnetnsid))

		By("running a discovery")
		allnetns, _ = discoverRedux()
		Expect(allnetns).To(HaveKey(realnetnsid),
			"did not discover %s netns in %s", testVlanNetnsName, allnetns.String())

		var master *NifAttrs

		By("ensuring VLAN attributes and master relation")
		testnetns := allnetns[realnetnsid]
		Expect(testnetns.Nifs).To(HaveLen(2), testnetns.NifsString())
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceOfKindWithName("", "lo"),
			HaveInterfaceOfKindWithName("vlan", testVlanNifName),
		), testnetns.NifsString())
		vlannif := testnetns.NamedNifs[testVlanNifName]
		vlan := vlannif.(Vlan).Vlan()

		Expect(vlan.Master).NotTo(BeNil())
		master = vlan.Master.Nif()
		Expect(master.Name).To(Equal(masternif.Nif().Name))
		Expect(master.Netns).NotTo(BeIdenticalTo(masternif.Nif().Netns))

		By("ensuring castability")
		vlans := master.Nif().Slaves.OfKind("vlan")
		Expect(vlans).NotTo(BeEmpty())
		Expect(func() {
			for _, slave := range vlans {
				_ = slave.(Vlan)
				_ = slave.(*VlanAttrs)
			}
		}).NotTo(Panic())
	})

})
