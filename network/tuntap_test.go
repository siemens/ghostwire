// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"time"

	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/notwork/link"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

const tapNamePrefix = "tap-"

var _ = Describe("TAPs and TUNs", func() {

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

	It("discovers a TAP", func() {
		By("creating a TAP")
		tap := link.NewTransient(&netlink.Tuntap{
			Mode:   netlink.TUNTAP_MODE_TAP,
			Queues: 1,
		}, tapNamePrefix).(*netlink.Tuntap)
		defer tap.Fds[0].Close()

		By("discovering the TAP")
		allnetns, _ := discoverRedux()
		currentNetns := Successful(ops.NamespacePath("/proc/self/ns/net").ID())
		Expect(allnetns).To(HaveKey(currentNetns))
		Expect(allnetns[currentNetns].NamedNifs).To(HaveKey(tap.Attrs().Name))
		gwtap := allnetns[currentNetns].NamedNifs[tap.Attrs().Name].(TunTap)
		Expect(gwtap.Nif().Kind).To(Equal("tuntap"))
		Expect(gwtap.TunTap().Mode).To(Equal(TunTapModeTap))
	})

})
