// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cli

import (
	"github.com/siemens/ghostwire/v2/cmd/internal/cli/cliplugin"
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
)

// AddFlags runs all registered SetupCLI plugin functions (in group "cli") in
// order to register CLI flags for the specified root command.
func AddFlags(rootCmd *cobra.Command) {
	for _, setupCLI := range plugger.Group[cliplugin.SetupCLI]().Symbols() {
		setupCLI(rootCmd)
	}
}

// BeforeCommand runs all registered BeforeRun plugin functions (in group "cli")
// just before the selected command runs; it terminates as soon as the first
// plugin function returns a non-nil error.
func BeforeCommand(cmd *cobra.Command) error {
	for _, beforeCmd := range plugger.Group[cliplugin.BeforeCommand]().Symbols() {
		if err := beforeCmd(cmd); err != nil {
			return err
		}
	}
	return nil
}
