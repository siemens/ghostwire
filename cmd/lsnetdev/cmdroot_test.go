// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"
	"os"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("lsnetdev command", func() {

	It("shows help with all flags", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(SatisfyAll(
			MatchRegexp(`--mode mode .*\(default sysfs\)`),
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
		Expect(out.String()).To(Equal("lsnetdev version " + gostwire.SemVersion + "\n"))
	})

	It("rejects an invalid mode", func() {
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--" + ModeFlag, "foo"})
		rootCmd.SetOut(GinkgoWriter)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(MatchError(ContainSubstring("invalid argument")))
	})

	It("rejects an unsupported mode", func(ctx context.Context) {
		rootCmd := newRootCmd()
		rootCmd.SetContext(ctx)
		Expect(discoverNetdevs(rootCmd, Mode(42))).To(
			MatchError(ContainSubstring("unsupported discovery mode")))
	})

	DescribeTable("lists the netdevs of network namespaces",
		func(ctx context.Context, mode string, source string) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			var out safe.Buffer
			rootCmd := newRootCmd()
			rootCmd.SetContext(ctx)
			debug.SetWriter(rootCmd, GinkgoWriter)
			rootCmd.SetArgs([]string{"--" + ModeFlag, mode})
			rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
			rootCmd.SetErr(GinkgoWriter)
			Expect(rootCmd.Execute()).To(Succeed())
			Expect(out.String()).To(MatchRegexp(
				`(?s)^net:\[\d+\]\n  leader process\(es\)\n.*  netdev\(s\)\n.*    🕳 "lo" \(ifindex 1\) source %s\n        📥 RX queue 0`,
				source))
		},
		Entry(nil, "sysfs", "sysfs"),
		Entry(nil, "netlink", "netlink"),
		Entry(nil, "both", "netlink"),
	)

})
