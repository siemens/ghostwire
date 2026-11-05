// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package main

import (
	"encoding/json"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"

	gostwire "github.com/siemens/ghostwire/v2"
	apiv1 "github.com/siemens/ghostwire/v2/api/v1"
)

// newRootCmd returns the Cobra "root" command, thus the gostdump tool itself.
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "gostdump",
		Short:   "dumps (full) discovery results in JSON REST v1 format",
		Version: gostwire.SemVersion,
		Args:    cobra.NoArgs,
		// See: https://github.com/spf13/cobra/issues/340
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return clippy.BeforeCommand(rootCmd)
		},
		RunE: dumpDiscovery,
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	clippy.AddFlags(rootCmd) // ...runs all registered SetupCLI plugins.
	return
}

// dumpDiscovery discovers and then writes the discovery results in JSON REST
// API v1 format to the command's output.
func dumpDiscovery(cmd *cobra.Command, _ []string) error {
	cizer := turtles.Containerizer(cmd.Context(), cmd)
	defer cizer.Close()
	discovery := gostwire.Discover(cmd.Context(), cizer, nil)
	result := apiv1.NewDiscoveryResult(discovery)
	return json.NewEncoder(cmd.OutOrStdout()).Encode(&result)
}
