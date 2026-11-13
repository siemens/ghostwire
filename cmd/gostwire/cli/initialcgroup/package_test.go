// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package initialcgroup

import (
	"testing"

	"github.com/spf13/cobra"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCmdCliInitialCgroup(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/cmd/gostwire/cli/initialcgroup package")
}

// runBeforeCommand runs beforeCommand from a command with this package's flags
// and the specified CLI args.
func runBeforeCommand(args ...string) error {
	root := &cobra.Command{
		Use:  "root",
		RunE: func(cmd *cobra.Command, _ []string) error { return beforeCommand(cmd) },
	}
	root.SetOut(GinkgoWriter)
	root.SetErr(GinkgoWriter)
	setupCLI(root)
	root.SetArgs(args)
	return root.Execute()
}
