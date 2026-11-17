# Downstream Distributions

Downstream builds (distributions) can compose their own Gostwire service
binaries and container images from this repository **without forking it**: they
keep their additions – such as further decorators, HTTP endpoints, and a
differently themed and branded web UI – in their own Go module and web UI
sources, and then build using the upstream `Dockerfile`.

The [example downstream module](https://github.com/siemens/ghostwire/tree/main/examples/downstream)
in `examples/downstream/` gets built and smoke-tested by CI, so the extension
points described here won't rot.

## Go Side

### Main Package

The Gostwire service is implemented by the reusable
[`cmd/gostwire/app`](https://pkg.go.dev/github.com/siemens/ghostwire/v2/cmd/gostwire/app)
package; the upstream `cmd/gostwire` main package is just a thin wrapper around
`app.Main(app.WithUI(webui.FS()))`. The `app` package doesn't serve any web UI
on its own: `webui.FS()` from the `github.com/siemens/ghostwire/v2/webui`
package returns the web UI that the upstream `Dockerfile` builds and embeds. A
downstream main package looks like this:

```go
package main

import (
    "github.com/siemens/ghostwire/v2/cmd/gostwire/app"
    "github.com/siemens/ghostwire/v2/webui"

    _ "example.org/downstream/decorator/foo" // additional decorator(s)
    "example.org/downstream/endpoint"
)

var Version = "0.0.0-dev" // stamp using -ldflags "-X main.Version=..."

func main() {
    app.Main(
        app.WithName("foo"),
        app.WithDescription("foo discovery service"),
        app.WithVersion(Version),
        app.WithUI(webui.FS()),
        app.WithRoutes(endpoint.Register),
    )
}
```

Importing `app` takes care of all process-level setup, so a downstream main
package cannot forget about it: locking the initial OS thread in `init()`,
pulling in the C constructor of lxkns' mount namespace sandbox (so cgo must be
enabled), and registering the CLI flag plugins, such as `--http`, `--brand`,
`--initialcgroup`, `--debug`, and `--silent`.

The available options are:

| Option | Purpose |
| --- | --- |
| `WithName(name)` | command name in usage, `--version` output, and logs; defaults to `gostwire`. |
| `WithDescription(text)` | short description in the help, logged at start. |
| `WithVersion(version)` | version in the `--version` output and logged at start, defaults to the Ghostwire semantic version; when different, the Ghostwire version gets logged too. Discovery results always report the Ghostwire version as their creator version. |
| `WithUI(fsys)` | serves the web UI from the complete web UI build in `fsys`; `webui.FS()` returns the web UI that the upstream `Dockerfile` builds from `webappsrc` and embeds. Without `WithUI`, the service serves only its HTTP endpoints. |
| `WithRoutes(fn)` | `fn(router, cmd, containerizer)` registers additional HTTP handlers with the [gorilla/mux](https://github.com/gorilla/mux) router; can be passed multiple times. |

Instead of `app.Main` (which handles SIGINT/SIGTERM/SIGQUIT and the exit code),
downstream main packages can also use `app.Run(ctx, opts...)` or
`app.NewRootCmd(opts...)`.

### Decorators

All upstream decorators (`decorator/all`) are always included. Additional
decorators register themselves as `decorator.Decorate` plugins in their package
`init()` functions, so the downstream main package only needs to blank-import
them:

```go
func init() {
    plugger.Group[decorator.Decorate]().Register(
        Decorate, plugger.WithPlugin("foo"))
}
```

### HTTP Endpoints

Additional endpoints get registered either using `app.WithRoutes`, or as
`app.RouteHandler` plugins. Additional endpoints never shadow the upstream
endpoints (`/json`, `/mobyshark`, `/mobydig`), but always take precedence over
the web UI. Endpoints run discoveries using the passed containerizer:

```go
func Register(r *mux.Router, cmd *cobra.Command, cizer containerizer.Containerizer) {
    r.HandleFunc("/api/v1/foo", func(w http.ResponseWriter, req *http.Request) {
        result := gostwire.Discover(req.Context(), cizer, nil)
        // ...
    }).Methods(http.MethodGet)
}
```

Websocket endpoints should use
[`cmd/gostwire/wsconn`](https://pkg.go.dev/github.com/siemens/ghostwire/v2/cmd/gostwire/wsconn),
which applies the same origin checks as the upstream endpoints, including the
`--allowed-origin` CLI flag.

### Go Module

The downstream `go.mod` requires the Ghostwire module at the upstream release
to build against:

```text
module example.org/downstream

go 1.26.6

require github.com/siemens/ghostwire/v2 v2.9.0
```

For local development against a Ghostwire checkout, use a `go.work` (or a
`replace` directive):

```bash
go work init ./ghostwire ./downstream
```

## Container Image

The upstream `deployments/gostwire/Dockerfile` builds downstream images using
these build arguments and contexts:

| Build arg/context | Purpose |
| --- | --- |
| `--build-context bctxN=…` | additional Go module source (N = 1…9). |
| `MODN` | directory of module N inside the build's Go workspace, such as `./downstream`. |
| `WSDISKPATHS` | all workspace module directories, such as `./ghostwire ./downstream`. |
| `MAINPKG` | main package to build, relative to the Ghostwire module directory; defaults to `./cmd/gostwire`. For instance, `../downstream/cmd/gostwire`. |
| `--build-context webappsrc=…` | web UI sources. |
| `VERSION` | version stamped into the Ghostwire module and the web UI. |
| `LDFLAGS`, `TAGS` | Go linker flags and build tags; when overriding `LDFLAGS`, keep `-s -w -extldflags=-static`. |

Inside the build, the Ghostwire module lives in `/ws/ghostwire` and each
additional module in `/ws/<MODN>`. When an additional module requires (or even
`replace`s) a version of another workspace module – such as the Ghostwire module
itself – the build pins that version to the workspace module. This way, the
build neither fetches the Ghostwire module nor follows `replace` directives
pointing outside the build, and the downstream module can require a Ghostwire
version that hasn't been released yet.

Everything else stays as in the upstream image: the static binary
(`xx-verify --static`), its file capabilities, the minimal root filesystem,
entrypoint, default command, and labels.

### Using `docker-build.sh`

`scripts/docker-build.sh` derives the `bctxN`, `MODN`, and `WSDISKPATHS`
arguments from the `go.work` file (use the `GOWORK` environment variable to
point to a particular one), as well as `MAINPKG` from the host path of the main
package in the `MAINPKG` environment variable. It must be run from inside the
Ghostwire module directory.

For instance, with the downstream module in `../downstream` and a workspace
file `../go.work` using `./ghostwire` and `./downstream`:

```bash
cd ghostwire
MAINPKG=../downstream/cmd/gostwire \
  scripts/docker-build.sh deployments/gostwire/Dockerfile \
    --load -t downstream \
    --build-arg VERSION=$(git describe --tags --always) \
    --build-arg LDFLAGS="-s -w -extldflags=-static -X main.Version=1.2.3" \
    --build-context webappsrc=./webui
```

The CI job building the in-repository example does:

```bash
cd ghostwire
repo="$PWD"
mkdir -p /tmp/ws
(cd /tmp/ws && go work init "${repo}" "${repo}/examples/downstream")
GOWORK=/tmp/ws/go.work MAINPKG=./examples/downstream/cmd/gostwire \
  scripts/docker-build.sh deployments/gostwire/Dockerfile \
    --load -t gostwire:downstream \
    --build-context webappsrc=./webui
examples/downstream/smoketest.sh gostwire:downstream
```

### Plain `docker buildx`

Without the script, pass all nine `bctxN` contexts, using an empty directory for
the unused ones:

```bash
mkdir -p .emptyctx
docker buildx build -f deployments/gostwire/Dockerfile \
  --build-context bctx1=../downstream --build-arg MOD1=./downstream \
  $(for n in 2 3 4 5 6 7 8 9; do echo --build-context bctx$n=.emptyctx; done) \
  --build-arg WSDISKPATHS="./ghostwire ./downstream" \
  --build-arg MAINPKG=../downstream/cmd/gostwire \
  --build-context webappsrc=./webui \
  --load -t downstream .
```
