// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package service

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("service flags", func() {

	It("registers as a CLI setup plugin", func() {
		Expect(plugger.Group[cliplugin.SetupCLI]().Plugins()).To(ContainElement("ghostwire/service"))
	})

	DescribeTable("getting the service settings",
		func(args []string, addr string, timeout time.Duration, origins []string) {
			var actualAddr string
			var actualTimeout time.Duration
			var actualOrigins []string
			root := &cobra.Command{
				Use: "root",
				Run: func(cmd *cobra.Command, _ []string) {
					actualAddr = Address(cmd)
					actualTimeout = ShutdownTimeout(cmd)
					actualOrigins = AllowedOrigins(cmd)
				},
			}
			root.SetOut(GinkgoWriter)
			root.SetErr(GinkgoWriter)
			setupCLI(root)
			root.SetArgs(append([]string{}, args...))
			Expect(root.Execute()).To(Succeed())
			Expect(actualAddr).To(Equal(addr))
			Expect(actualTimeout).To(Equal(timeout))
			Expect(actualOrigins).To(Equal(origins))
		},
		Entry("defaults", nil, DefaultAddress, DefaultShutdownTimeout, []string{}),
		Entry("explicit",
			[]string{
				"--" + HTTPFlag, "127.0.0.1:5005",
				"--" + ShutdownFlag, "2s",
				"--" + AllowedOriginFlag, "https://a.example,https://b.example",
				"--" + AllowedOriginFlag, "*",
			},
			"127.0.0.1:5005", 2*time.Second, []string{"https://a.example", "https://b.example", "*"}),
	)

	It("rejects an invalid shutdown duration", func() {
		root := &cobra.Command{Use: "root", Run: func(*cobra.Command, []string) {}}
		root.SetOut(GinkgoWriter)
		root.SetErr(GinkgoWriter)
		setupCLI(root)
		root.SetArgs([]string{"--" + ShutdownFlag, "foo"})
		Expect(root.Execute()).To(MatchError(ContainSubstring("invalid argument")))
	})

})
