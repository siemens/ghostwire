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
| `WithBrand(name)` | default of the `--brand` CLI flag, that is, the brand name passed to the web UI; see [Web UI](#web-ui-themes-and-branding). |
| `WithUI(fsys)` | serves the web UI from the complete web UI build in `fsys`; `webui.FS()` returns the web UI that the upstream `Dockerfile` builds from `webappsrc` and embeds, see [Web UI](#web-ui-themes-and-branding). Without `WithUI`, the service serves only its HTTP endpoints. |
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
| `--build-context webappsrc=…` | web UI sources: a yarn project whose `yarn build` writes the web UI into `build/`, see [below](#web-ui-in-container-images). |
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

## Web UI: Themes and Branding

The web UI keeps all its MUI-based components, views, and mechanics, but the MUI
themes and the branding can be configured at build time.

The web UI in `webui/` is a yarn workspace consisting of:

- `packages/ghostwire-ui`: the `ghostwire-ui` library with all components,
  views, models, hooks, icons, the default themes, and `createGhostwireApp`,
- `packages/app`: the OSS web app, which builds into `webui/build` for embedding
  into the service binary.

The app is created using `createGhostwireApp(config)` from `ghostwire-ui`,
where all configuration elements default to the Ghostwire defaults:

```tsx
import type { GhostwireAppConfig } from 'ghostwire-ui'

const config: GhostwireAppConfig = {
    themes: {
        // MUI ThemeOptions (or Themes), deeply merged into the default
        // Ghostwire theme options; light and dark mode are independent.
        light: { palette: { primary: { main: '#6a1b9a' } } },
        dark: { palette: { primary: { main: '#ce93d8' } } },
    },
    branding: {
        name: 'Foo',               // app bar, title, about, help
        icon: FooIcon,             // React component taking MUI SvgIconProps
        favicon: fooFaviconUrl,    // replaces the favicon at runtime
        about: FooAbout,           // MDX content replacing the about content
        links: [                   // legal links at the end of the about view
            { label: 'Imprint', href: 'https://example.org/imprint' },
            { label: 'Open Source Software Notices', href: 'https://example.org/oss' },
        ],
        version: '1.2.3',          // defaults to VITE_REACT_APP_GIT_VERSION
    },
}

export default config
```

The Ghostwire-specific palette elements (such as `palette.wire`, `palette.containee`,
et cetera) are always present, so theme overrides only need to specify what they
want to change, but they can also change the Ghostwire-specific elements.

About content can use the `<Brand/>`, `<BrandIcon/>`, `<AppVersion/>`,
`<DiscoveryMetadata/>`, and `<CaptureMode/>` shortcodes; please see
`packages/ghostwire-ui/src/views/about/About.mdx` for the default about
content.

> [!NOTE] The brand name and icon passed in by the Gostwire service via its
> `--brand` and `--brandicon` CLI flags take precedence over the build-time
> branding. As `--brand` defaults to "Ghostwire", downstream main packages with
> a differently branded web UI should set the same default brand name using
> `app.WithBrand("Foo")`. The `--brandicon` flag defaults to empty, so the
> build-time brand icon is used unless explicitly overridden.

There are two ways for downstream builds to theme and brand the web UI:

1. a configuration overlay for the OSS web app, using only the dependencies of
   the OSS web app,
2. their own web app with its own `package.json`, using the `ghostwire-ui`
   library; this allows adding further dependencies, such as theme packages.

### Configuration Overlay

The OSS web app entry point `packages/app/src/index.tsx` imports its
configuration from the `ghostwire-app-config` module, which the Vite
configuration resolves to:

- `packages/app/src/defaultconfig.ts` (empty configuration) by default,
- the module specified in the `VITE_APP_CONFIG` environment variable (relative
  to `webui/`), taken from the environment or the `packages/app/.env` files.

Similarly, `VITE_APP_PUBLIC_DIR` optionally specifies a different directory of
static assets instead of `packages/app/public/` (favicon, app icons,
`manifest.json`); it must contain all files referenced from `index.html`.

The overlay module can import from `ghostwire-ui` and all of the web UI's
dependencies, as well as MDX, SVG and other assets next to it; it must be placed
inside `webui/` (such as in `webui/overlay/`) so that it can resolve the web
UI's dependencies. For instance:

```bash
cd webui
cp -r ../../foo/webui-overlay overlay
VITE_APP_CONFIG=overlay/config.tsx yarn build
```

The e2e test overlay in `webui/e2e/branded/overlay/` is a complete example; the
`branded` Playwright project builds and tests it (`make e2e`), and the
"Ghostwire/App Theming and Branding" Storybook story shows it.

In container image builds, the `Dockerfile` copies an optional `webappoverlay`
build context into `overlay/` inside the web UI sources, and passes the
`VITE_APP_CONFIG` and `VITE_APP_PUBLIC_DIR` build args to the web UI build:

```bash
scripts/docker-build.sh deployments/gostwire/Dockerfile \
    --build-context webappsrc=./webui \
    --build-context webappoverlay=../foo/webui-overlay \
    --build-arg VITE_APP_CONFIG=overlay/config.tsx \
    ...
```

Without these, the default Ghostwire web UI gets built.

### Own Web App Using the Library

A downstream web app with its own `package.json` depends on the `ghostwire-ui`
library at a pinned upstream version. The library gets consumed as TypeScript,
TSX, and MDX sources, so there is no build step and no registry involved. The
app then needs:

- the peer dependencies `react`, `react-dom`, `@mui/material`,
  `@mui/icons-material`, `@mui/system`, `@emotion/react`, `@emotion/styled`, and
  `vite` (please see the library's `package.json` for the version ranges),
- the `ghostwireUI()` Vite plugin preset from `ghostwire-ui/vite` (compiling
  the help and about MDX contents, SVG React components, and deduplicating
  singletons such as React and MUI), together with `@vitejs/plugin-react`,
- fonts and global CSS as needed; the OSS web app uses the Roboto and Roboto
  Mono fonts from `@fontsource`.

```ts
// vite.config.ts
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import { ghostwireUI } from 'ghostwire-ui/vite'

export default defineConfig({
    base: './',
    build: { outDir: 'build' },
    plugins: [react(), ghostwireUI()],
})
```

```tsx
// src/main.tsx
import { createRoot } from 'react-dom/client'
import { createGhostwireApp } from 'ghostwire-ui'

const App = createGhostwireApp({ themes: { /* ... */ }, branding: { /* ... */ } })
createRoot(document.getElementById('root')!).render(<App />)
```

Its `index.html` must contain `<base href="/" />` and
`<script>window.dynvars = {}</script>`, which the service rewrites with the
base path and the dynamic variables.

The example web app in `examples/downstream/webui/` is built by CI.

#### Git Dependency (Pinned Tag)

Yarn 4 can install a workspace from a subdirectory of a git repository, pinned
to a tag (or commit):

```json
"dependencies": {
    "ghostwire-ui": "git+https://github.com/siemens/ghostwire.git#cwd=webui&workspace=ghostwire-ui&tag=v2.9.0"
}
```

Yarn clones the repository, installs the upstream web UI workspace, and then
packs the `ghostwire-ui` workspace, so this takes some time on the first
install. Git dependencies need to be explicitly approved in `.yarnrc.yml`:

```yaml
nodeLinker: node-modules
approvedGitRepositories:
  - "https://github.com/siemens/ghostwire.git"
```

#### `portal:` to a Checkout

With the Ghostwire repository checked out at the pinned tag, such as a git
submodule in `ghostwire/`:

```json
"dependencies": {
    "ghostwire-ui": "portal:./ghostwire/webui/packages/ghostwire-ui"
}
```

As the library's dependencies get installed into the app's `node_modules`, the
app must resolve them from the symlinked library location: set
`resolve.preserveSymlinks: true` in the Vite configuration, and
`"preserveSymlinks": true` in the TypeScript configuration. The
`ghostwire-ui/vite` preset loads its own build-time dependencies from the app's
directory (the current working directory) when they aren't installed next to
the library, so the checkout doesn't need its own `yarn install`.

#### Package Tarball

`yarn workspace ghostwire-ui pack --out ghostwire-ui.tgz` inside `webui/`
creates a package tarball of the library, which the app can depend on using
`"ghostwire-ui": "file:./vendor/ghostwire-ui.tgz"`.

### Web UI in Container Images

The `Dockerfile` builds the web UI from the `webappsrc` build context, which
must be a yarn project (`package.json` with `"packageManager": "yarn@4…"`,
`yarn.lock`) whose `yarn build` writes the static web UI files into `build/` in
its root. This is the case for the upstream `webui/` workspace as well as a
downstream web app with `vite build --outDir build`:

```bash
scripts/docker-build.sh deployments/gostwire/Dockerfile \
    --build-context webappsrc=../downstream-webui \
    ...
```

The build uses yarn 4 with the `node-modules` linker, regardless of the
project's yarn configuration. Dependency installation is cached as long as the
`package.json`, `yarn.lock`, `.yarnrc.yml`, and package tarball (`*.tgz`)
files don't change; git dependencies are supported. Dependencies outside the
`webappsrc` directory are not available, so `portal:` dependencies need to point
inside it.

> [!NOTE] The service binary built by the upstream `cmd/gostwire` passes the
> brand name "Ghostwire" to the web UI unless started with `--brand`; downstream
> main packages should use `app.WithBrand`.
