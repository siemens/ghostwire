// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	nl "github.com/mdlayher/netlink"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/notwork/netdevsim"
	"github.com/thediveo/notwork/netdevsim/ensure"
	"github.com/thediveo/spacetest"
	"github.com/thediveo/spacetest/mntns"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"

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

			if !ensure.Netdevsim() {
				Skip("wants netdevsim")
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

		It("lists netdev NAPIs", func() {
			// This is going to be real fun now: we want to use a throw-away
			// network namespace, this is actually quite simple. But we later
			// are forced to tinker with our virtual throw-away network
			// interface through the sysfs and this is where it gets ugly. We
			// need a new sysfs instance that correctly reflects our throw-away
			// network namespace...
			netnsfd := netns.NewTransient()
			mntnsfd, procfsroot := mntns.NewTransient()
			// ...welcome to the very, very dark side. And in case you wonder,
			// there's no process attached to this transient mount namespace,
			// except for the following brief moment where we quickly get in,
			// remount /sysfs, and then leave. From then on, we simply travel
			// through the procfs root wormholes to access the VFS elements
			// inside the mount namespace.
			spacetest.Execute(func() {
				Expect(unix.Mount(
					"none", "/sys", "sysfs",
					unix.MS_NODEV|unix.MS_NOEXEC|unix.MS_NOSUID|unix.MS_RELATIME,
					"")).To(Succeed(),
					"cannot mount new sysfs instance on /sys")
			}, netnsfd, mntnsfd)
			// We need a network interface that registers NAPIs and we want it
			// preferably on virtual network interfaces. Regardless of the many
			// goblin hallucinations there is only one pseudo-hardware virtual
			// network device that does the trick, the "netdevsim", and they
			// still tell me it's not. Now, netdevsim is intended for kernel
			// developers, not for us user spacers. Doesn't really repel us and
			// the trick to know here is that we need to create a netdevsim with
			// more than one queue and we have to bring it up.
			_, ndsims := netdevsim.NewTransient(
				netdevsim.WithRxTxQueueCountEach(2),
				netdevsim.InNamespace(netnsfd))
			ndsim := ndsims[0]
			netns.Execute(netnsfd, func() {
				Expect(netlink.LinkSetUp(ndsim)).To(Succeed())
			})

			// okay, things now get really dirty.
			Expect(os.WriteFile(
				fmt.Sprintf(filepath.Join(procfsroot+"/sys/class/net/%s/threaded"), ndsim.Attrs().Name),
				[]byte("1\n"), 0)).To(Succeed())

			c := spacetest.Call(func() *Conn {
				return Successful(Dial(nil))
			}, netnsfd)
			defer func() { _ = c.Close() }()
			nifnapis := Successful(c.nAPIs())
			Expect(nifnapis).To(HaveKeyWithValue(
				uint32(ndsim.Attrs().Index),
				ContainElement(And(
					HaveField("ID", Not(BeZero())),
					HaveField("IfIndex", uint32(ndsim.Attrs().Index)),
					HaveField("PID", Not(BeZero()))))))

			hwnicnapis := Successful(c.napis(uint32(ndsim.Attrs().Index)))
			Expect(hwnicnapis).NotTo(BeEmpty())
			var ndsimNapis napis
			Expect(hwnicnapis).To(ContainElement(HaveField("PID", Not(BeZero())), &ndsimNapis))
			for _, napi := range ndsimNapis {
				kthread := model.NewProcess(model.PIDType(napi.PID), false)
				Expect(kthread).NotTo(BeNil())
				Expect(kthread.Name).To(And(
					HavePrefix("napi/"),
					HaveSuffix(fmt.Sprintf("-%d", napi.ID))))
			}
		})

	})

})
