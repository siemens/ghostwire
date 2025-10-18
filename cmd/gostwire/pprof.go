// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build pprof

package main

import (
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
)

// Automatically register a pprof HTTP handler on "/debug/pprof/" for several of
// the standard pprof topics/themes.
func init() {
	slog.Info("pprof handler enabled")
	for _, route := range []struct {
		profile string
		handler http.HandlerFunc
	}{
		{"", pprof.Index},
		{"allocs", pprof.Handler("allocs").ServeHTTP},
		{"block", pprof.Handler("block").ServeHTTP},
		{"cmdline", pprof.Cmdline},
		{"goroutine", pprof.Handler("goroutine").ServeHTTP},
		{"heap", pprof.Handler("heap").ServeHTTP},
		{"mutex", pprof.Handler("mutex").ServeHTTP},
		{"profile", pprof.Profile},
		{"threadcreate", pprof.Handler("threadcreate").ServeHTTP},
		{"trace", pprof.Trace},
	} {
		plugger.Group[RouteHandler]().Register(
			func(cmd *cobra.Command, cizer containerizer.Containerizer) (string, string, http.HandlerFunc) {
				return "GET", "/debug/pprof/" + route.profile, route.handler
			}, plugger.WithPlugin("pprof"+route.profile))
	}
}
