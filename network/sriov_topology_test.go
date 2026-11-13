// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("SR-IOV topology", func() {

	var netns *NetworkNamespace

	BeforeEach(func() {
		tmpdir := Successful(os.MkdirTemp("", "ghostwire-sysfs-"))
		DeferCleanup(os.RemoveAll, tmpdir)
		DeferCleanup(func(orig string) { sysfsPCIDevicesPath = orig }, sysfsPCIDevicesPath)
		sysfsPCIDevicesPath = tmpdir

		Expect(os.Mkdir(filepath.Join(tmpdir, "0000:01:00.0"), 0o755)).To(Succeed())
		Expect(os.WriteFile(filepath.Join(tmpdir, "0000:01:00.0", "sriov_numvfs"), []byte("1\n"), 0o644)).To(Succeed())
		for _, vf := range []string{"0000:01:10.0", "0000:02:10.0"} {
			Expect(os.Mkdir(filepath.Join(tmpdir, vf), 0o755)).To(Succeed())
		}
		Expect(os.Symlink("../0000:01:00.0", filepath.Join(tmpdir, "0000:01:10.0", "physfn"))).To(Succeed())
		Expect(os.Symlink("../0000:02:00.0", filepath.Join(tmpdir, "0000:02:10.0", "physfn"))).To(Succeed())
		Expect(os.Mkdir(filepath.Join(tmpdir, "0000:03:00.0"), 0o755)).To(Succeed())

		netns = ownNetworkNamespace()
	})

	It("resolves PFs and their VFs from sysfs", func() {
		nif := func(index int, name, busaddr string) *NifAttrs {
			return &NifAttrs{Index: index, Name: name, Netns: netns, DriverInfo: NifDriverInfo{BusInfo: busaddr}}
		}
		pf := nif(1, "pf", "0000:01:00.0")
		vf := nif(2, "vf", "0000:01:10.0")
		orphanvf := nif(3, "orphan", "0000:02:10.0")
		plain := nif(4, "plain", "0000:03:00.0")
		virtual := nif(5, "virtual", "")
		netns.Nifs = map[int]Interface{1: pf, 2: vf, 3: orphanvf, 4: plain, 5: virtual}

		resolveSRIOVTopology(NetworkNamespaces{netns.ID(): netns})

		Expect(pf.SRIOVRole).To(Equal(PCI_SRIOV_PF))
		Expect(pf.Slaves).To(ConsistOf(BeIdenticalTo(vf)))
		Expect(vf.SRIOVRole).To(Equal(PCI_SRIOV_VF))
		Expect(vf.PF).To(BeIdenticalTo(pf))
		Expect(orphanvf.SRIOVRole).To(Equal(PCI_SRIOV_VF))
		Expect(orphanvf.PF).To(BeNil())
		Expect(plain.SRIOVRole).To(Equal(PCI_NIC))
		Expect(virtual.SRIOVRole).To(Equal(PCI_NIC))
	})

})
