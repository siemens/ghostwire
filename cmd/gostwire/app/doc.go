/*
Package app implements the Gostwire service as a reusable application, so that
downstream builds can compose their own service binaries from it without
forking the Ghostwire repository.

The Gostwire service binary itself is just a thin wrapper, serving the web UI
built and embedded by the upstream Dockerfile:

	package main

	import (
		"github.com/siemens/ghostwire/v2/cmd/gostwire/app"
		"github.com/siemens/ghostwire/v2/webui"
	)

	func main() {
		app.Main(app.WithUI(webui.FS()))
	}

Downstream builds use the same [Main] (or [Run] and [NewRootCmd]), passing
options to customize the service, such as [WithName], [WithVersion], [WithUI],
and [WithRoutes]:

	package main

	import (
		"github.com/siemens/ghostwire/v2/cmd/gostwire/app"
		"github.com/siemens/ghostwire/v2/webui"

		_ "example.org/downstream/decorator/foo" // additional decorator
	)

	func main() {
		app.Main(
			app.WithName("foo"),
			app.WithVersion("1.2.3"),
			app.WithUI(webui.FS()),
			app.WithRoutes(registerFooRoutes),
		)
	}

# Process-Level Setup

Importing this package takes care of the process-level setup required by the
Gostwire discovery engine, so downstream main packages cannot forget about it:

  - locking the initial goroutine to the initial OS thread (the task group
    leader), see also golang/go issue #53210,
  - pulling in the C constructor of lxkns' mount namespace sandbox, which
    requires building with cgo enabled,
  - registering the Gostwire CLI flag plugins, such as “--http”, “--brand”,
    and “--initialcgroup”, as well as the lxkns and clippy plugins for
    container engine discovery, debug logging, and silencing.

# Web UI

This package doesn't serve any web UI on its own: main packages pass the web
UI to serve using [WithUI]. [github.com/siemens/ghostwire/v2/webui.FS] returns
the web UI that the upstream Dockerfile builds and embeds using the “webui”
build tag; without this build tag, it returns the “webui/build” directory
relative to the current working directory instead, such as for development.
Without [WithUI], the service serves only its HTTP endpoints.

# Decorators

The Gostwire decorators in [github.com/siemens/ghostwire/v2/decorator/all] are
always included. Additional decorators simply register themselves as
[github.com/siemens/ghostwire/v2/decorator.Decorate] plugins in their package
init functions, so downstream main packages only need to blank-import them.

# HTTP Endpoints

Gostwire's own HTTP endpoints register themselves as [RouteHandler] plugins.
Downstream builds add their own endpoints by passing [WithRoutes] options, or
as further [RouteHandler] plugins.

For each request, the service uses the first registered route matching the
request's method and path. It registers the routes in this order:

  - all [RouteHandler] plugins, ordered by their plugin names and optional
    placements,
  - the routes of the [WithRoutes] options, in the order given,
  - the web UI, if any, serving all remaining requests.

Thus, a [WithRoutes] route with the same method and path as one of Gostwire's
own endpoints never gets used: it cannot shadow the upstream endpoint, neither
by accident nor on purpose. Use [WithRouteOverride] or [WithoutRoutes] instead,
see below. Requests not matching an upstream route still reach [WithRoutes]
routes, such as requests for the same path but a different method, or for
paths below the path of a prefix route.

In turn, [WithRoutes] routes always take precedence over the web UI. Please
note that this includes paths of the web UI's own views, such as “/about”:
such a view then still works when navigating inside the web UI, but not when
directly loading or reloading its URL.

Additional [RouteHandler] plugins get ordered among Gostwire's own route
plugins, so they must not register the same method and path as an upstream
endpoint, as they might shadow it, depending on their plugin names.

Gostwire's own endpoints can be explicitly overridden using [WithRouteOverride]
by the plugin names of their route handler plugins, such as [RoutePluginJSON]
for the “/json” endpoint, getting passed the upstream handler for optionally
post-processing its responses:

	app.WithRouteOverride(app.RoutePluginJSON,
		func(upstream http.HandlerFunc, cmd *cobra.Command, cizer containerizer.Containerizer) http.HandlerFunc {
			return func(w http.ResponseWriter, req *http.Request) {
				w.Header().Set("X-Foo", "bar")
				upstream(w, req)
			}
		})

Or they can be removed using [WithoutRoutes], so that their paths can be
reimplemented using [WithRoutes]. Overriding or removing route plugins that
don't exist makes the service fail to start, so that upstream route changes
don't go unnoticed.
*/
package app
