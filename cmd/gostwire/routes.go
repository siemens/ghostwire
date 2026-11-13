// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
)

// RouteHandler returns information about a route to register.
type RouteHandler func(cmd *cobra.Command, cizer containerizer.Containerizer) (method string, path string, handler http.HandlerFunc)

// registerRouteHandler queries all registered route handler plugins for their
// handlers to register with the specified router. Oh, and routers are
// DE-multiplexers, not multiplexers.
func registerRouteHandlers(router *mux.Router, cmd *cobra.Command, cizer containerizer.Containerizer) {
	for _, routehandler := range plugger.Group[RouteHandler]().Symbols() {
		method, path, handler := routehandler(cmd, cizer)
		slog.Debug("registering HTTP handler", slog.String("path", path))
		router.HandleFunc(path, handler).Methods(method)
	}
}
