// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"log/slog"
	"os"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/netdevsim"
	"github.com/thediveo/notwork/netdevsim/ensure"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("netdev NETLINK discovery", Ordered, func() {

	BeforeAll(func() {
		if !ensure.Netdevsim() {
			Skip("needs root and netdevsim kernel module")
		}
	})

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

	It("discovers netdev information", func() {
		netnsfd := netns.NewTransient()
		_, ndevsim := netdevsim.NewTransient(
			netdevsim.WithRxTxQueueCountEach(8),
			netdevsim.InNamespace(netnsfd))
		netns.Execute(netnsfd, func() {
			ndevsim := Successful(netlink.LinkByName(ndevsim[0].Attrs().Name))
			link.EnsureUp(ndevsim)
		})

		allnetns := discover.Namespaces(
			discover.WithStandardDiscovery(),
			discover.FromTasks())
		Expect(allnetns.Namespaces[model.NetNS]).To(HaveKey(species.NamespaceIDfromInode(netns.Ino(netnsfd))))

		netdevsByNetns := Successful(Discover(allnetns))
		transientnetns := allnetns.Namespaces[model.NetNS][species.NamespaceIDfromInode(netns.Ino(netnsfd))]
		Expect(netdevsByNetns[transientnetns]).To(ContainElement(And(
			HaveField("Name", ndevsim[0].Attrs().Name),
			HaveField("Queues", HaveLen(2*8)))))
	})

})
