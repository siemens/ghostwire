// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package main

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
)

// RouteHandler returns information about a route to register.
type RouteHandler func() (method string, path string, handler http.HandlerFunc)

// registerRouteHandler queries all registered route handler plugins for their
// handlers to register with the specified router. Oh, and routers are
// DE-multiplexers, not multiplexers.
func registerRouteHandlers(router *mux.Router) {
	for _, rh := range plugger.Group[RouteHandler]().Symbols() {
		method, path, handler := rh()
		log.Debugf("registering HTTP handler for path %s", path)
		router.HandleFunc(path, handler).Methods(method)
	}
}
