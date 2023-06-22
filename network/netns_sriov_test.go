// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("SR-IOV", func() {

	It("discovers SR-IOV topology", func() {
		By("running a network namespace discovery")
		allnetns, _ := discoverRedux()

		By("looking for a PF with VFs")
		var pf Interface
		var numvfs uint64
	PFScan:
		for _, netns := range allnetns {
			for _, nif := range netns.Nifs {
				if !nif.Nif().Physical {
					continue
				}
				contents, err := os.ReadFile(nif.Nif().SysfsBusPath() + "/sriov_numvfs")
				if err != nil {
					continue
				}
				Expect(nif.Nif().SRIOVRole).To(Equal(PCI_SRIOV_PF))
				numvfs, err = strconv.ParseUint(strings.TrimSuffix(string(contents), "\n"), 10, 16)
				Expect(err).NotTo(HaveOccurred(), "invalid sriov_numvfs %q", string(contents))
				if numvfs == 0 {
					continue // not a candidate
				}
				pf = nif
				break PFScan
			}
		}
		if pf == nil {
			Skip("no PF with at least one VF available")
		}

		var vfs []Interface
		Expect(pf.Nif().Slaves).To(ContainElement(HaveField("SRIOVRole", PCI_SRIOV_VF), &vfs))
		Expect(vfs).To(HaveLen(int(numvfs)))
		Expect(vfs).To(HaveEach(HaveField("PF", BeIdenticalTo(pf))))
	})

})
