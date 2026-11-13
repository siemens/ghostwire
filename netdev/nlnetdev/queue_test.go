// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"log/slog"
	"os"
	"time"

	nl "github.com/mdlayher/netlink"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netdev queues", func() {

	Context("Queue attribute decoding", func() {

		It("returns an error if queue information is not decodable", func() {
			Expect(newQueue(nil)).Error().To(HaveOccurred())
		})

		It("rejects insufficient NETLINK Queue information", func() {
			Expect(newQueue([]byte{42, 6, 66})).Error().To(HaveOccurred())
		})

		It("rejects invalid NETLINK Queue information", func() {
			enc := nl.NewAttributeEncoder()
			enc.String(rxtxlayout.NETDEV_A_QUEUE_ID, "foobar!")
			Expect(newQueue(Successful(enc.Encode()))).Error().To(MatchError(
				ContainSubstring("cannot decode netdev queue information")))
		})

		It("rejects incomplete Queue information", func() {
			enc := nl.NewAttributeEncoder()
			enc.Uint32(rxtxlayout.NETDEV_A_QUEUE_ID, 42)
			Expect(newQueue(Successful(enc.Encode()))).Error().To(MatchError(
				ContainSubstring("incomplete netdev queue information")))
		})

		It("grabs all Queue attributes correctly", func() {
			enc := nl.NewAttributeEncoder()
			enc.Uint32(rxtxlayout.NETDEV_A_QUEUE_ID, 42)
			enc.Uint32(rxtxlayout.NETDEV_A_QUEUE_TYPE, uint32(rxtxlayout.NETDEV_QUEUE_TYPE_TX))
			enc.Uint32(rxtxlayout.NETDEV_A_QUEUE_IFINDEX, 666)
			enc.Uint32(rxtxlayout.NETDEV_A_QUEUE_NAPI_ID, 123)
			q := Successful(newQueue(Successful(enc.Encode())))
			Expect(q).To(And(
				HaveField("ID", uint32(42)),
				HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX),
				HaveField("IfIndex", uint32(666)),
				HaveField("NapiID", uint32(123))))
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

			DeferCleanup(slog.SetDefault, slog.Default())
			slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))

		})

		It("returns an empty map if everything is down", func() {
			defer netns.EnterTransient()()
			c := Successful(Dial(nil))
			DeferCleanup(c.Close)
			Expect(Successful(c.netdevsQueues())).To(BeEmpty())
		})

		It("lists queues and organizes them by network interface", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransientUp()
			mcvlan := macvlan.NewTransient(dmy)
			link.EnsureUp(mcvlan)

			c := Successful(Dial(nil))
			DeferCleanup(c.Close)
			nifsqs := Successful(c.netdevsQueues())
			Expect(nifsqs).To(HaveLen(2)) // without lo, as it is down
			Expect(nifsqs).To(HaveKeyWithValue(
				uint32(dmy.Attrs().Index),
				ConsistOf(
					And(
						HaveField("IfIndex", uint32(dmy.Attrs().Index)),
						HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX)),
					And(
						HaveField("IfIndex", uint32(dmy.Attrs().Index)),
						HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX)))))
			Expect(nifsqs).To(HaveKeyWithValue(
				uint32(mcvlan.Attrs().Index),
				ConsistOf(
					And(
						HaveField("IfIndex", uint32(mcvlan.Attrs().Index)),
						HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX)),
					And(
						HaveField("IfIndex", uint32(mcvlan.Attrs().Index)),
						HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX)))))
		})

		It("lists queues", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransientUp()

			c := Successful(Dial(nil))
			DeferCleanup(c.Close)
			nicqueues := Successful(c.Queues(uint32(dmy.Attrs().Index)))
			Expect(nicqueues).To(HaveLen(2))
			Expect(nicqueues).To(ConsistOf(
				And(
					HaveField("IfIndex", uint32(dmy.Attrs().Index)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_RX)),
				And(
					HaveField("IfIndex", uint32(dmy.Attrs().Index)),
					HaveField("Type", rxtxlayout.NETDEV_QUEUE_TYPE_TX))))
		})

		It("returns an error when there is no matching network interface", func() {
			defer netns.EnterTransient()()
			c := Successful(Dial(nil))
			DeferCleanup(c.Close)
			Expect(c.Queues(42)).Error().To(MatchError(
				ContainSubstring("no such device")))
		})

		It("returns an error when specifying the zero interface index", func() {
			c := Successful(Dial(nil))
			DeferCleanup(c.Close)
			Expect(c.Queues(0)).Error().To(MatchError(
				ContainSubstring("invalid ifindex")))
		})
	})

})
