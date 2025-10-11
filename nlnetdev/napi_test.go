// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"fmt"
	"os"
	"time"

	nl "github.com/mdlayher/netlink"
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/notwork/macvlan"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netdev NAPIs", func() {

	Context("Queue attribute decoding", func() {

		It("returns an error if NAPI information is not decodable", func() {
			Expect(newNAPI(nil)).Error().To(HaveOccurred())
		})

		It("rejects insufficient NETLINK NAPI information", func() {
			Expect(newNAPI([]byte{42, 6, 66})).Error().To(HaveOccurred())
		})

		It("rejects invalid NETLINK NAPI information", func() {
			enc := nl.NewAttributeEncoder()
			enc.String(rxtxlayout.NETDEV_A_NAPI_ID, "foobar!")
			Expect(newNAPI(Successful(enc.Encode()))).Error().To(MatchError(
				ContainSubstring("cannot decode netdev NAPI information")))
		})

		It("rejects incomplete NAPI information", func() {
			enc := nl.NewAttributeEncoder()
			enc.Uint32(rxtxlayout.NETDEV_A_NAPI_ID, 42)
			Expect(newNAPI(Successful(enc.Encode()))).Error().To(MatchError(
				ContainSubstring("incomplete netdev NAPI information")))
		})

		It("grabs all NAPI attributes correctly", func() {
			enc := nl.NewAttributeEncoder()
			enc.Uint32(rxtxlayout.NETDEV_A_NAPI_ID, 42)
			enc.Uint32(rxtxlayout.NETDEV_A_NAPI_IFINDEX, 666)
			enc.Uint32(rxtxlayout.NETDEV_A_NAPI_IRQ, 123)
			enc.Uint32(rxtxlayout.NETDEV_A_NAPI_PID, 321)
			q := Successful(newNAPI(Successful(enc.Encode())))
			Expect(q).To(And(
				HaveField("ID", uint32(42)),
				HaveField("IfIndex", uint32(666)),
				HaveField("HasIRQ", BeTrue()),
				HaveField("IRQ", uint32(123)),
				HaveField("PID", uint32(321))))
		})

	})

	When("using the real mccoy", func() {

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			goodgos := Goroutines()
			goodfds := Filedescriptors()
			DeferCleanup(func() {
				Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})
		})

		It("lists netdev NAPIs", func() {
			// We need a "real" "hardware" network interface, as otherwise we cannot
			// get any NAPI information.
			hwnic := macvlan.LocateHWParent()
			Expect(os.WriteFile(
				fmt.Sprintf("/sys/class/net/%s/threaded", hwnic.Attrs().Name),
				[]byte("1\n"), 0)).To(Succeed())

			c := Successful(Dial(nil))
			defer c.Close()
			nifnapis := Successful(c.nAPIs())
			Expect(nifnapis).To(HaveKeyWithValue(
				uint32(hwnic.Attrs().Index),
				ContainElement(And(
					HaveField("ID", Not(BeZero())),
					HaveField("IfIndex", uint32(hwnic.Attrs().Index)),
					HaveField("PID", Not(BeZero()))))))

			hwnicnapis := Successful(c.napis(uint32(hwnic.Attrs().Index)))
			Expect(hwnicnapis).NotTo(BeEmpty())
			var hwnapis napis
			Expect(hwnicnapis).To(ContainElement(HaveField("PID", Not(BeZero())), &hwnapis))
			for _, hwnapi := range hwnapis {
				kthread := model.NewProcess(model.PIDType(hwnapi.PID), false)
				Expect(kthread).NotTo(BeNil())
				Expect(kthread.Name).To(And(
					HavePrefix("napi/"),
					HaveSuffix(fmt.Sprintf("-%d", hwnapi.ID))))
			}
		})

	})

})
