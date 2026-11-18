// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

package app

import (
	"context"
	"io/fs"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/spf13/cobra"
	"github.com/thediveo/lxkns/containerizer"
	_ "github.com/thediveo/lxkns/ops/mountineer/mntnssandbox" // pre-Go runtime C constructor; requires cgo.
	"github.com/thediveo/nonstd/sets"

	gostwire "github.com/siemens/ghostwire/v2"
)

func init() {
	// lock the initial OS thread (that is, the Linux task group leader
	// representing the whole process) "M0" to the initial goroutine "G0". This
	// avoids that M0 ever gets scheduled to one of the goroutines that call
	// ops.Exec in order to execute a function in a different network namespace
	// and afterwards throw away the thread M scheduled and locked to such a
	// goroutine. Locking in any package init function makes the main function
	// run on the initial OS thread.
	//
	// For the details and background information, please see:
	//
	// golang/go issue 53210: "runtime: on Linux, better do not treat the
	// initial thread/task group leader as any other thread/task",
	// https://github.com/golang/go/issues/53210
	//
	// Google Groups go-nuts discussion "LockOSThread, switching (Linux kernel)
	// namespaces: what happens to the main thread...?",
	// https://groups.google.com/g/golang-nuts/c/dx-jweSVxHk
	runtime.LockOSThread()
}

// Defaults of the Gostwire service.
const (
	DefaultName        = "gostwire"
	DefaultDescription = "gostwire virtual network topology and configuration discovery service"
	defaultBanner      = "Gostwire \"The Sequel\" virtual network topology and configuration discovery service"
)

// RoutesFunc registers additional HTTP handlers with the specified router. It
// gets passed the root command, so it can query CLI flags, and the
// containerizer, so it can run discoveries using [gostwire.Discover].
type RoutesFunc func(r *mux.Router, cmd *cobra.Command, cizer containerizer.Containerizer)

// Option configures the Gostwire service created by [NewRootCmd], [Run], and
// [Main].
type Option func(*config)

type config struct {
	name           string
	description    string
	banner         string
	version        string
	brand          string
	uifs           fs.FS
	routes         []RoutesFunc
	routeOverrides map[string][]RouteOverride // by route plugin name
	removedRoutes  sets.Set[string]           // route plugin names
}

func newConfig(opts ...Option) *config {
	c := &config{
		name:        DefaultName,
		description: DefaultDescription,
		banner:      defaultBanner,
		version:     gostwire.SemVersion,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithName sets the name of the service command, as shown in the usage, the
// “--version” output, and the service logs. It defaults to “gostwire”.
func WithName(name string) Option {
	return func(c *config) {
		c.name = name
	}
}

// WithDescription sets the short description of the service, as shown in the
// help output and logged when the service starts.
func WithDescription(description string) Option {
	return func(c *config) {
		c.description = description
		c.banner = description
	}
}

// WithVersion sets the version of the service, as shown in the “--version”
// output and logged when the service starts. It defaults to the Gostwire
// semantic version. Please note that the discovery results always report the
// Gostwire semantic version as their creator version.
func WithVersion(version string) Option {
	return func(c *config) {
		c.version = version
	}
}

// WithBrand sets the default of the “--brand” CLI flag, that is, the brand
// name the service passes to the web UI. As the passed brand name takes
// precedence over the brand name the web UI has been built with, downstream
// builds with a differently branded web UI should set the same brand name here.
func WithBrand(name string) Option {
	return func(c *config) {
		c.brand = name
	}
}

// WithUI serves the web UI from the specified file system, which must contain
// a complete web UI build with an “index.html” at its root. Use
// [github.com/siemens/ghostwire/v2/webui.FS] for the web UI built and embedded
// by the upstream Dockerfile, or a downstream module's own embedded web UI
// build. Without WithUI, the service serves only its HTTP endpoints, but no
// web UI.
func WithUI(uifs fs.FS) Option {
	return func(c *config) {
		if uifs != nil {
			c.uifs = uifs
		}
	}
}

// WithRoutes registers additional HTTP handlers using the specified function.
// It can be passed multiple times, registering in the order given. Additional
// handlers cannot shadow the handlers registered as [RouteHandler] plugins,
// but always take precedence over serving the web UI. Use [WithRouteOverride]
// and [WithoutRoutes] to override and remove [RouteHandler] plugin routes.
func WithRoutes(fn RoutesFunc) Option {
	return func(c *config) {
		if fn != nil {
			c.routes = append(c.routes, fn)
		}
	}
}

// WithRouteOverride overrides the handler of the route served by the route
// handler plugin with the specified plugin name, such as [RoutePluginJSON],
// keeping the route's method and path. The override gets passed the upstream
// handler, which it can call, such as for post-processing its response, or
// completely reimplement the route. It can be passed multiple times for the
// same route plugin, with each override wrapping the previous one. The service
// fails to start when there is no route plugin with the specified name.
func WithRouteOverride(plugin string, override RouteOverride) Option {
	return func(c *config) {
		if override == nil {
			return
		}
		if c.routeOverrides == nil {
			c.routeOverrides = map[string][]RouteOverride{}
		}
		c.routeOverrides[plugin] = append(c.routeOverrides[plugin], override)
	}
}

// WithoutRoutes removes the routes served by the route handler plugins with the
// specified plugin names, such as [RoutePluginMobydig]. Their paths then get
// served by any additional handlers registered using [WithRoutes], or
// otherwise by the web UI. The service fails to start when there is no route
// plugin with a specified name, or when it is also overridden.
func WithoutRoutes(plugins ...string) Option {
	return func(c *config) {
		if c.removedRoutes == nil {
			c.removedRoutes = sets.New[string]()
		}
		for _, plugin := range plugins {
			c.removedRoutes.Add(plugin)
		}
	}
}

// Run runs the service with the CLI arguments from [os.Args] until the passed
// context is done or the service fails.
func Run(ctx context.Context, opts ...Option) error {
	return NewRootCmd(opts...).ExecuteContext(ctx)
}

// Main runs the service until it receives SIGINT, SIGTERM, or SIGQUIT, and
// exits the process with a non-zero exit code if the service fails.
func Main(opts ...Option) {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	// Cobra already reports any error, see also:
	// https://github.com/spf13/cobra/issues/304
	if err := Run(ctx, opts...); err != nil {
		stop()
		os.Exit(1)
	}
}
