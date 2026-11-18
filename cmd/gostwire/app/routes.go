// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

package app

import (
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/nonstd/sets"
	"github.com/thediveo/nonstd/xslices"
)

// Plugin names of the Gostwire route handler plugins (not their URL paths),
// such as for overriding or removing them using [WithRouteOverride] and
// [WithoutRoutes].
const (
	RoutePluginJSON      = "json"      // serves GET /json: discovery results.
	RoutePluginMobyshark = "mobyshark" // serves GET /mobyshark: capture target discovery results.
	RoutePluginMobydig   = "mobydig"   // serves GET /mobydig: neighborhood services websocket.
)

// RouteHandler returns information about a route to register. Route handlers
// are plugins registered with the plugger.Group[RouteHandler] plugin group, so
// that packages can register their routes in their init functions.
type RouteHandler func(cmd *cobra.Command, cizer containerizer.Containerizer) (method string, path string, handler http.HandlerFunc)

// RouteOverride returns the handler to use for a route instead of the passed
// upstream handler. The returned handler can call the upstream handler, such
// as for post-processing its response, or reimplement the route completely.
type RouteOverride func(upstream http.HandlerFunc, cmd *cobra.Command, cizer containerizer.Containerizer) http.HandlerFunc

// routePlugins returns the plugin names of all registered route handler
// plugins.
func routePlugins() sets.Set[string] {
	return sets.New(xslices.Map(plugger.Group[RouteHandler]().PluginsSymbols(),
		func(symbol plugger.Symbol[RouteHandler]) string { return symbol.Plugin })...)
}

// checkRoutes returns an error if any route override or removal refers to a
// route plugin that doesn't exist, or if a route plugin is to be both
// overridden and removed.
func (c *config) checkRoutes(plugins sets.Set[string]) error {
	known := func() string {
		elements := plugins.Elements()
		slices.Sort(elements)
		return strings.Join(elements, ", ")
	}
	for plugin := range c.routeOverrides {
		if !plugins.Contains(plugin) {
			return fmt.Errorf("cannot override unknown route plugin %q, known route plugins: %s", plugin, known())
		}
	}
	for plugin := range c.removedRoutes {
		if !plugins.Contains(plugin) {
			return fmt.Errorf("cannot remove unknown route plugin %q, known route plugins: %s", plugin, known())
		}
		if _, ok := c.routeOverrides[plugin]; ok {
			return fmt.Errorf("cannot both override and remove route plugin %q", plugin)
		}
	}
	return nil
}

// registerRouteHandler queries all registered route handler plugins for their
// handlers to register with the specified router, applying the configured
// route overrides and removals. Oh, and routers are DE-multiplexers, not
// multiplexers.
func registerRouteHandlers(router *mux.Router, cmd *cobra.Command, cizer containerizer.Containerizer, cfg *config) {
	for _, symbol := range plugger.Group[RouteHandler]().PluginsSymbols() {
		if cfg.removedRoutes.Contains(symbol.Plugin) {
			slog.Debug("removing HTTP handler", slog.String("plugin", symbol.Plugin))
			continue
		}
		method, path, handler := symbol.S(cmd, cizer)
		for _, override := range cfg.routeOverrides[symbol.Plugin] {
			slog.Debug("overriding HTTP handler",
				slog.String("plugin", symbol.Plugin), slog.String("path", path))
			handler = override(handler, cmd, cizer)
		}
		slog.Debug("registering HTTP handler", slog.String("path", path))
		router.HandleFunc(path, handler).Methods(method)
	}
}
