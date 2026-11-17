// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

package app

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
	"github.com/siemens/ghostwire/v2/cmd/gostwire/cli/brand"
	_ "github.com/siemens/ghostwire/v2/cmd/gostwire/cli/initialcgroup"
	"github.com/siemens/ghostwire/v2/cmd/gostwire/cli/service"
	"github.com/siemens/ghostwire/v2/cmd/gostwire/wsconn"
)

// NewRootCmd returns the Cobra "root" command, thus the Gostwire service
// itself, configured using the specified options.
func NewRootCmd(opts ...Option) (rootCmd *cobra.Command) {
	cfg := newConfig(opts...)
	rootCmd = &cobra.Command{
		Use:     cfg.name,
		Short:   cfg.description,
		Version: cfg.version,
		Args:    cobra.NoArgs,
		// See: https://github.com/spf13/cobra/issues/340
		SilenceUsage: true,
		PersistentPreRunE: func(*cobra.Command, []string) error {
			return clippy.BeforeCommand(rootCmd)
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return serve(cmd, cfg)
		},
	}
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	clippy.AddFlags(rootCmd) // ...runs all registered SetupCLI plugins.
	// Apply WithBrand by changing the default of the "--brand" flag registered
	// by the brand CLI plugin, also in the help. Setting the flag's value
	// directly doesn't mark the flag as changed, so an explicit "--brand" on
	// the command line still takes precedence.
	if cfg.brand != "" {
		if flag := rootCmd.PersistentFlags().Lookup(brand.BrandFlag); flag != nil {
			_ = flag.Value.Set(cfg.brand)
			flag.DefValue = cfg.brand
		}
	}
	return
}

// serve runs the service until the command's context is done.
func serve(cmd *cobra.Command, cfg *config) error {
	versionAttrs := []any{slog.String("version", cfg.version)}
	if cfg.version != gostwire.SemVersion {
		versionAttrs = append(versionAttrs, slog.String("ghostwire", gostwire.SemVersion))
	}
	slog.Info(cfg.banner, versionAttrs...)
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
		Handler:           newHandler(cmd, cizer, cfg),
		ReadHeaderTimeout: 10 * time.Second,
	}
	slog.Info("starting "+cfg.name+" server",
		slog.String("addr", listener.Addr().String()))
	served := make(chan error, 1)
	go func() { served <- server.Serve(listener) }()
	select {
	case err := <-served:
		return err
	case <-ctx.Done():
	}

	maxwait := service.ShutdownTimeout(cmd)
	slog.Info("gracefully shutting down "+cfg.name+" server, waiting...",
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
	slog.Info(cfg.name + " server stopped")
	return nil
}
