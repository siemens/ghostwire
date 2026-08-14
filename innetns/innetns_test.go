// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package innetns

import (
	"errors"
	"log/slog"
	"os"
	"runtime"

	"github.com/thediveo/caps/v2"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

type mockedNamespace struct {
	model.Namespace // "inherit" all the methods
	typ             species.NamespaceType
	ref             []string
}

func (m *mockedNamespace) Type() species.NamespaceType { return m.typ }
func (m *mockedNamespace) Ref() model.NamespaceRef     { return m.ref }

var _ = Describe("in a netns", func() {

	BeforeEach(func() {
		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	It("reports a visitation error when powerless", func() {
		if os.Getuid() != 0 {
			Skip("need root to become powerless (sic!)")
		}

		netnsfd := netns.NewTransient()
		DeferCleanup(unix.Close, netnsfd)

		allns := discover.Namespaces(discover.WithStandardDiscovery())
		newnetns := allns.Namespaces[model.NetNS][species.NamespaceIDfromInode(netns.Ino(netnsfd))]
		Expect(newnetns).NotTo(BeNil())

		runtime.LockOSThread() // never unlock; throw-away OS-level thread
		Expect(Successful(caps.OfCurrentTask()).Effective().Clear().ApplyToCurrentTask()).
			Error().NotTo(HaveOccurred())

		Expect(Run(newnetns, func() error { return nil })).To(MatchError(
			ContainSubstring("cannot enter namespace, operation not permitted")))
	})

	When("having the power", func() {

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
		})

		It("rejects nil network namespaces", func() {
			Expect(Run(nil, func() error { return nil })).To(MatchError("nil namespace"))
		})

		It("rejects non-net namespaces", func() {
			Expect(Run(&mockedNamespace{
				typ: species.CLONE_NEWPID,
			}, func() error { return nil })).To(
				MatchError(ContainSubstring("invalid non-netns namespace of type CLONE_NEWPID")))
		})

		It("rejects namespaces without a reference", func() {
			Expect(Run(&mockedNamespace{
				typ: species.CLONE_NEWNET,
			}, func() error { return nil })).To(MatchError("invalid empty netns reference"))
		})

		It("runs in the original network namespace", func() {
			allns := discover.Namespaces(discover.WithStandardDiscovery())
			ourproc := allns.Processes[model.PIDType(os.Getpid())]
			Expect(ourproc).NotTo(BeNil())
			ournetns := ourproc.Namespaces[model.NetNS]
			Expect(ournetns).NotTo(BeNil())

			dmy := dummy.NewTransient()
			Expect(Run(ournetns, func() error {
				_, err := netlink.LinkByName(dmy.Attrs().Name)
				return err
			})).To(Succeed())
		})

		It("runs in another network namespace", func() {
			netnsfd := netns.NewTransient()
			dmy := dummy.NewTransient(dummy.InNamespace(netnsfd))
			Expect(netlink.LinkByName(dmy.Attrs().Name)).Error().To(HaveOccurred())

			allns := discover.Namespaces(discover.WithStandardDiscovery())
			othernetns := allns.Namespaces[model.NetNS][species.NamespaceIDfromInode(netns.Ino(netnsfd))]
			Expect(Run(othernetns, func() error {
				_, err := netlink.LinkByName(dmy.Attrs().Name)
				return err
			})).To(Succeed())

			canary := errors.New("DO'H!")
			origino := netns.CurrentIno()
			Expect(Run(othernetns, func() error {
				Expect(netns.CurrentIno()).NotTo(Equal(origino))
				return canary
			})).To(BeIdenticalTo(canary))

		})

	})

})
