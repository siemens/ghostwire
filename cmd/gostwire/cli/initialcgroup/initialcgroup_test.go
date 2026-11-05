// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package initialcgroup

import (
	"slices"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("initial cgroup flags", func() {

	It("registers as CLI plugins, running after the debug plugin", func() {
		Expect(plugger.Group[cliplugin.SetupCLI]().Plugins()).To(ContainElement("ghostwire/initialcgroup"))
		plugins := plugger.Group[cliplugin.BeforeCommand]().Plugins()
		Expect(plugins).To(ContainElements("clippy/debug", "ghostwire/initialcgroup"))
		Expect(slices.Index(plugins, "clippy/debug")).To(
			BeNumerically("<", slices.Index(plugins, "ghostwire/initialcgroup")))
	})

	It("hides the re-execution flag", func() {
		root := &cobra.Command{Use: "root"}
		setupCLI(root)
		Expect(root.PersistentFlags().Lookup(InitialCgroupFlag)).To(HaveField("Hidden", BeFalse()))
		Expect(root.PersistentFlags().Lookup(CgroupSwitchedFlag)).To(HaveField("Hidden", BeTrue()))
	})

	It("does nothing unless asked to", func() {
		root := &cobra.Command{
			Use:  "root",
			RunE: func(cmd *cobra.Command, _ []string) error { return beforeCommand(cmd) },
		}
		root.SetOut(GinkgoWriter)
		root.SetErr(GinkgoWriter)
		setupCLI(root)
		root.SetArgs([]string{})
		Expect(root.Execute()).To(Succeed())
	})

})
