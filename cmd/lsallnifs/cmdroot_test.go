// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"io"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
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

})
