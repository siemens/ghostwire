// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"github.com/thediveo/clippy"
	_ "github.com/thediveo/clippy/debug"
	_ "github.com/thediveo/lxkns/cmd/cli/silent"
	"github.com/thediveo/lxkns/cmd/cli/turtles"
	"github.com/thediveo/lxkns/ops/mountineer"

	gostwire "github.com/siemens/ghostwire/v2"
	_ "github.com/siemens/ghostwire/v2/cmd/cli/initialcgroup"
	"github.com/siemens/ghostwire/v2/cmd/cli/service"
	"github.com/siemens/ghostwire/v2/cmd/internal/wsconn"
)

// newRootCmd returns the Cobra "root" command, thus the Gostwire service
// itself.
func newRootCmd() (rootCmd *cobra.Command) {
	rootCmd = &cobra.Command{
		Use:     "gostwire",
		Short:   "gostwire virtual network topology and configuration discovery service",
		Version: gostwire.SemVersion,
		Args:    cobra.NoArgs,
		// See: https://github.com/spf13/cobra/issues/340
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return clippy.BeforeCommand(rootCmd)
		},
		RunE: serve,
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	clippy.AddFlags(rootCmd) // ...runs all registered SetupCLI plugins.
	return
}

// serve runs the Gostwire service until the command's context is done.
func serve(cmd *cobra.Command, _ []string) error {
	slog.Info("Gostwire \"The Sequel\" virtual network topology and configuration discovery service",
		slog.String("version", gostwire.SemVersion))
	slog.Info("Copyright (c) Siemens AG 2018-2026")

	if pausebin := mountineer.StandaloneSandboxBinary(); pausebin != "" {
		slog.Info("using optimized pandora's sandbox binary", slog.String("path", pausebin))
	}

	ctx := cmd.Context()
	slog.Debug("using container engine \"Turtles Anywhere\" technology")
	cizer := turtles.Containerizer(ctx, cmd)
	defer cizer.Close()

	slog.Debug("priming list of discovered container engines in background")
	go func() {
		_ = gostwire.Discover(ctx, cizer, nil)
	}()

	wsconn.AllowOrigins(service.AllowedOrigins(cmd)...)
	listener, err := net.Listen("tcp", service.Address(cmd))
	if err != nil {
		return fmt.Errorf("cannot start service: %w", err)
	}
	server := &http.Server{
		Handler:           newHandler(cmd, cizer),
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("starting gostwire server",
		slog.String("addr", listener.Addr().String()))
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}

	maxwait := service.ShutdownTimeout(cmd)
	slog.Info("gracefully shutting down gostwire server, waiting...",
		slog.Duration("maxwait", maxwait))
	shutdownCtx, cancel := context.WithTimeout(context.Background(), maxwait)
	defer cancel()
	// Shutdown neither closes nor waits for hijacked websocket connections.
	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("cannot gracefully shut down service: %w", err)
	}
	if err := <-served; !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	slog.Info("gostwire server stopped")
	return nil
}
