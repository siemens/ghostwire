// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package cliplugin

import "github.com/spf13/cobra"

// SetupCLI defines an exposed plugin symbol type for adding things to a cobra
// root command.
type SetupCLI func(*cobra.Command)

// BeforeCommand defines an exposed plugin symbol type for running checks after
// the command line args have been processed and before running the command.
type BeforeCommand func(*cobra.Command) error
