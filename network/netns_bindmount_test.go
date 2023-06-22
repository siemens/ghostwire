// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"
	"time"

	"github.com/thediveo/lxkns/nstest"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/testbasher"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

const testNetnsName = "gostwire-testnetns"

var _ = Describe("bind-mounted network namespaces", func() {

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

	It("discovers network interfaces in process-less network namespaces correctly", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a bind-mounted network namespace with some network interfaces")
		scripts := testbasher.Basher{}
		defer scripts.Done()
		scripts.Common(nstest.NamespaceUtilsScript)
		scripts.Common("netnsname=" + testNetnsName)
		scripts.Script("main", `
ip netns del ${netnsname} || true
ip link del wallace || true
ip netns add ${netnsname}
ip -netns ${netnsname} link add wallace type veth peer name gromit
namespaceid /run/netns/${netnsname}
read # wait for test to proceed
ip netns del ${netnsname}
`)
		cmd := scripts.Start("main")
		defer cmd.Close()

		realnetnsid := nstest.CmdDecodeNSId(cmd)
		testnetnsid, err := ops.NamespacePath("/proc/1/root/run/netns/" + testNetnsName).ID()
		Expect(err).NotTo(HaveOccurred())
		Expect(testnetnsid).To(Equal(realnetnsid))

		By("running a discovery")
		allnetns, _ := discoverRedux()
		Expect(allnetns).To(HaveKey(realnetnsid))

		testnetns := allnetns[realnetnsid]
		Expect(testnetns.Nifs).To(HaveLen(3))
		for _, nif := range testnetns.Nifs {
			By(nif.Nif().Name)
		}
		Expect(testnetns.Nifs).To(ContainElements(
			HaveInterfaceOfKindWithName("", "lo"),
			HaveInterfaceOfKindWithName("veth", "wallace"),
			HaveInterfaceOfKindWithName("veth", "gromit"),
		))
	})

})
