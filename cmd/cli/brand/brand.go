// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package brand

import (
	"encoding/base64"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
)

// Names of the CLI flags provided in this package.
const (
	BrandFlag     = "brand"
	BrandIconFlag = "brandicon"
)

// DefaultBrand is the brand name shown in the UI unless specified otherwise.
const DefaultBrand = "Ghostwire"

func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		setupCLI, plugger.WithPlugin("ghostwire/brand"))
}

// setupCLI adds the branding flags.
func setupCLI(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.String(BrandFlag, DefaultBrand, "brand name to show in the UI")
	pf.String(BrandIconFlag, "", "brand icon SVG markup (optionally base64 encoded)")
}

// Name returns the brand name to show in the UI, as specified for the passed
// command.
func Name(cmd *cobra.Command) string {
	name, _ := cmd.Flags().GetString(BrandFlag)
	return name
}

// Icon returns the brand icon SVG markup, as specified for the passed command,
// decoding it first if it is base64 encoded.
func Icon(cmd *cobra.Command) string {
	icon, _ := cmd.Flags().GetString(BrandIconFlag)
	if decoded, err := base64.StdEncoding.DecodeString(icon); err == nil {
		return string(decoded)
	}
	return icon
}
