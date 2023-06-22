// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

//go:build pprof
// +build pprof

package main

import (
	"net/http"
	"net/http/pprof"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
)

// Automatically register a pprof HTTP handler on "/debug/pprof/" for several of
// the standard pprof topics/themes.
func init() {
	log.Infof("pprof handler enabled")
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
		route := route // sic! closure over value, not loop variable.
		plugger.Group[RouteHandler]().Register(
			func() (string, string, http.HandlerFunc) {
				return "GET", "/debug/pprof/" + route.profile, route.handler
			}, plugger.WithPlugin("pprof"+route.profile))
	}
}
