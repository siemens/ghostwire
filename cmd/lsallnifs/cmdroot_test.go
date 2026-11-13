// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"
	stdnet "net"
	"os"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/notwork/nlhandle"
	"github.com/thediveo/notwork/vxlan"
	"github.com/thediveo/safe"
	"github.com/thediveo/spacetest"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("lsallnifs command", func() {

	It("shows help with all flags", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(SatisfyAll(
			ContainSubstring("-x, --all"),
			ContainSubstring("-t, --tenants"),
			ContainSubstring("-p, --ports"),
			ContainSubstring("-a, --addresses"),
			ContainSubstring("-r, --routes"),
			ContainSubstring("--debug"),
			ContainSubstring("--silent"),
			ContainSubstring("--version"),
		))
	})

	It("shows its version", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--version"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(Equal("lsallnifs version " + gostwire.SemVersion + "\n"))
	})

	It("rejects arguments", func() {
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"foo"})
		rootCmd.SetOut(GinkgoWriter)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(MatchError(ContainSubstring(`unknown command "foo"`)))
	})

	It("lists network namespaces with their network interfaces", func(ctx context.Context) {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetContext(ctx)
		debug.SetWriter(rootCmd, GinkgoWriter)
		rootCmd.SetArgs([]string{"--all"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(MatchRegexp(
			`(?s)^lsallnifs\n.*net:\[\d+\] with .*:\n  tenants:\n.*  network interfaces:\n.* lo\(1\):`))
	})

	It("lists addresses, ports, as well as MACVLAN, IPVLAN, and VXLAN relations", func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		netnsfd := spacetest.NewTransient(unix.CLONE_NEWNET)
		// MACVLANs and IPVLANs cannot share the same parent.
		mcvparent := dummy.NewTransientUp(dummy.InNamespace(netnsfd))
		Expect(nlhandle.New(netnsfd).AddrAdd(mcvparent,
			Successful(netlink.ParseAddr("192.0.2.1/24")))).To(Succeed())
		mcv := macvlan.NewTransient(mcvparent,
			macvlan.InNamespace(netnsfd), macvlan.WithLinkNamespace(netnsfd))
		ipvparent := dummy.NewTransient(dummy.InNamespace(netnsfd))
		ipv := link.NewTransient(&netlink.IPVlan{
			LinkAttrs: netlink.LinkAttrs{ParentIndex: ipvparent.Attrs().Index},
			Mode:      netlink.IPVLAN_MODE_L2,
		}, "ipvl-", link.InNamespace(netnsfd), link.WithLinkNamespace(netnsfd))
		vxl := vxlan.NewTransient(mcvparent, vxlan.WithID(42),
			vxlan.InNamespace(netnsfd), vxlan.WithLinkNamespace(netnsfd))
		var listener stdnet.Listener
		spacetest.Execute(func() {
			listener = Successful(stdnet.Listen("tcp4", "0.0.0.0:0"))
		}, netnsfd)
		DeferCleanup(listener.Close)

		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetContext(ctx)
		debug.SetWriter(rootCmd, GinkgoWriter)
		rootCmd.SetArgs([]string{"-t", "-p", "-a"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(SatisfyAll(
			ContainSubstring("        192.0.2.1/24\n"),
			MatchRegexp(` TCP 0\.0\.0\.0:%d .* (\S+, ){4}\S+\n`,
				listener.Addr().(*stdnet.TCPAddr).Port),
			MatchRegexp(` MACVLAN: %s\(\d+\) in `, mcv.Attrs().Name),
			MatchRegexp(` master %s\(\d+\) in `, mcvparent.Attrs().Name),
			MatchRegexp(` IPVLAN: %s\(\d+\) in `, ipv.Attrs().Name),
			MatchRegexp(`\n      L2 mode, bridge\n\s+\S+\s+master %s\(\d+\) in `, ipvparent.Attrs().Name),
			MatchRegexp(` VXLAN overlay ID 42: %s\(\d+\) in `, vxl.Attrs().Name),
			MatchRegexp(` underlay %s\(\d+\) in `, mcvparent.Attrs().Name),
		))
	})

})
