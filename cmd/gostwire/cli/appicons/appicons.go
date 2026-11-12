// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package appicons

import (
	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
)

// PreferShowingAppIconsFlag is the name of the CLI flag provided in this
// package.
const PreferShowingAppIconsFlag = "prefer-showing-app-icons"

func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		setupCLI, plugger.WithPlugin("ghostwire/appicons"))
}

// setupCLI adds the app icons flag.
func setupCLI(cmd *cobra.Command) {
	cmd.PersistentFlags().Bool(PreferShowingAppIconsFlag, false,
		"default to showing Industrial Edge app icons when available, unless overridden by user's browser setting")
}

// PreferShowing returns true if the web UI should default to showing app icons
// (when available), as specified for the passed command.
func PreferShowing(cmd *cobra.Command) bool {
	prefer, _ := cmd.Flags().GetBool(PreferShowingAppIconsFlag)
	return prefer
}
