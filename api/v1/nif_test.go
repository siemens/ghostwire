// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package v1

import (
	"os"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/bridge"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/notwork/veth"
	"github.com/thediveo/notwork/vlan"
	"github.com/thediveo/notwork/vxlan"
	"github.com/thediveo/spacetest/netns"

	"github.com/siemens/ghostwire/v2/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("v1 network interfaces", func() {

	It("represents virtual network interface relations and configurations", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
		defer netns.EnterTransient()()

		By("creating virtual network interfaces")
		parent := dummy.NewTransientUp()
		mcvl := macvlan.NewTransient(parent)
		vl := vlan.NewTransient(42, parent)
		vx := vxlan.NewTransient(parent, vxlan.WithID(666), vxlan.WithDestinationPort(4789))
		br := bridge.NewTransient()
		dupond, dupont := veth.NewTransient()
		bridge.AddPort(br, dupond)

		By("discovering")
		// Discovery must not run on this thread, as it is attached to the transient netns.
		discovered := make(chan network.NetworkNamespaces)
		go func() {
			lxknsdisco := discover.Namespaces(discover.WithStandardDiscovery(), discover.FromTasks())
			discovered <- network.NewNetworkNamespaces(
				lxknsdisco.Namespaces[model.NetNS], lxknsdisco.Processes, lxknsdisco.Containers,
				network.WithSocketProcesses(lxknsdisco.SocketProcessMap))
		}()
		var allnetns network.NetworkNamespaces
		Eventually(discovered).Within(30 * time.Second).Should(Receive(&allnetns))
		ournetns := allnetns[species.NamespaceIDfromInode(netns.CurrentIno())]
		Expect(ournetns).NotTo(BeNil())
		nif := func(name string) networkInterface {
			GinkgoHelper()
			n := ournetns.NamedNifs[name]
			Expect(n).NotTo(BeNil(), "missing network interface %q", name)
			return newNif(n)
		}
		withName := func(name string) OmegaMatcher {
			return HaveValue(HaveField("Name", name))
		}

		By("checking the MACVLAN and its master")
		Expect(nif(mcvl.Attrs().Name).MacvlanMaster).To(withName(parent.Attrs().Name))
		parentnif := nif(parent.Attrs().Name)
		Expect(parentnif.Macvlans).To(ConsistOf(withName(mcvl.Attrs().Name)))

		By("checking the VLAN")
		vlannif := nif(vl.Attrs().Name)
		Expect(vlannif.Master).To(withName(parent.Attrs().Name))
		Expect(vlannif.Vlan).To(HaveValue(HaveField("VID", uint16(42))))

		By("checking the VXLAN")
		vxlannif := nif(vx.Attrs().Name)
		Expect(vxlannif.Vxlan).To(HaveValue(And(
			HaveField("VID", uint32(666)),
			HaveField("RemotePort", uint16(4789)),
			HaveField("UnderlayID", parentnif.ID),
		)))

		By("checking the slaves of the parent")
		Expect(parentnif.Slaves).To(ContainElements(
			withName(vl.Attrs().Name), withName(vx.Attrs().Name)))

		By("checking the bridge, its port, and the port's peer")
		Expect(nif(br.Attrs().Name).Slaves).To(ConsistOf(withName(dupond.Attrs().Name)))
		portnif := nif(dupond.Attrs().Name)
		Expect(portnif.Master).To(withName(br.Attrs().Name))
		Expect(portnif.Peer).To(HaveValue(HaveField("Name", dupont.Attrs().Name)))
	})

})
