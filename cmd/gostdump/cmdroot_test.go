// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"encoding/json"
	"io"

	"github.com/thediveo/clippy/debug"
	"github.com/thediveo/safe"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("gostdump command", func() {

	It("shows help with all flags", func() {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"--help"})
		rootCmd.SetOut(io.MultiWriter(&out, GinkgoWriter))
		Expect(rootCmd.Execute()).To(Succeed())
		Expect(out.String()).To(SatisfyAll(
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
		Expect(out.String()).To(Equal("gostdump version " + gostwire.SemVersion + "\n"))
	})

	It("rejects arguments", func() {
		rootCmd := newRootCmd()
		rootCmd.SetArgs([]string{"foo"})
		rootCmd.SetOut(GinkgoWriter)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(MatchError(ContainSubstring(`unknown command "foo"`)))
	})

	It("dumps the discovery results as JSON", func(ctx context.Context) {
		var out safe.Buffer
		rootCmd := newRootCmd()
		rootCmd.SetContext(ctx)
		debug.SetWriter(rootCmd, GinkgoWriter)
		rootCmd.SetArgs([]string{})
		rootCmd.SetOut(&out)
		rootCmd.SetErr(GinkgoWriter)
		Expect(rootCmd.Execute()).To(Succeed())
		var result map[string]any
		Expect(json.Unmarshal([]byte(out.String()), &result)).To(Succeed())
		Expect(result).To(HaveKey("metadata"))
		Expect(result).To(HaveKeyWithValue("network-namespaces", Not(BeEmpty())))
	})

})
