// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package service

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy/cliplugin"
	"github.com/thediveo/go-plugger/v3"
)

// Names of the CLI flags provided in this package.
const (
	HTTPFlag          = "http"
	ShutdownFlag      = "shutdown"
	AllowedOriginFlag = "allowed-origin"
)

// Defaults of the CLI flags provided in this package.
const (
	DefaultAddress         = "[::]:5000"
	DefaultShutdownTimeout = 15 * time.Second
)

func init() {
	plugger.Group[cliplugin.SetupCLI]().Register(
		setupCLI, plugger.WithPlugin("ghostwire/service"))
}

// setupCLI adds the service flags.
func setupCLI(cmd *cobra.Command) {
	pf := cmd.PersistentFlags()
	pf.String(HTTPFlag, DefaultAddress, "HTTP service address")
	pf.Duration(ShutdownFlag, DefaultShutdownTimeout, "graceful shutdown duration limit")
	pf.StringSlice(AllowedOriginFlag, nil,
		"additional cross-origin websocket origin(s) to accept, such as \"https://gw.example\"; \"*\" accepts any origin")
}

// Address returns the "host:port" address to serve on, as specified for the
// passed command.
func Address(cmd *cobra.Command) string {
	addr, _ := cmd.Flags().GetString(HTTPFlag)
	return addr
}

// ShutdownTimeout returns the maximum duration to wait for a graceful service
// shutdown, as specified for the passed command.
func ShutdownTimeout(cmd *cobra.Command) time.Duration {
	timeout, _ := cmd.Flags().GetDuration(ShutdownFlag)
	return timeout
}

// AllowedOrigins returns the additionally allowed cross-origin websocket
// origins, as specified for the passed command.
func AllowedOrigins(cmd *cobra.Command) []string {
	origins, _ := cmd.Flags().GetStringSlice(AllowedOriginFlag)
	return origins
}
