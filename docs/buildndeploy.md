# Build & Deploy

> [!ATTENTION] The `gostwire` (sic!) binary must be build with `cgo` **enabled**
> because it needs a C-based re-execution functionality in order to easily
> access other mount namespaces – please see also the
> [mountineers](https://pkg.go.dev/github.com/thediveo/lxkns/ops/mountineer)
> from the `lxkns` dependency for technical details.

## Build Tags

No build tags are required in order to successfully build the `ghostwire` module
as well as the `gostwire` (sic!) command binary.

Build tags are (only) required to:

- build and run tests,
- drop all C library dependencies,
- enable special `gostwire` service features, such as pprof support.

### `matchers`

The build tag `matchers` enables Gostwire-specific Gomega matchers (such as
`HaveInterfaceWithName`). These matchers are required by several unit tests in
different Gostwire packages.

Technical background: when enabled, these matchers can be used in test packages
other than where an individual matcher originates from. In order to not pollute
package exports, these matchers are disabled unless specifically enabled using
the `matchers` build tag. Unfortunately, it's not possible to put matchers into
`*_test.go` files as then they won't be reusable outside their own package.

### `kind`

The build tag `kind` enables
[KinD](https://github.com/kubernetes-sigs/kind)-based unit tests. These tests
start a Kubernetes-in-Docker test cluster during tests. There is no need to
install any `kind` binary, as Gostwire unit tests use Kind's Go API directly.
Since this pulls in the kind dependency, it must be enabled explicitly.

### `pprof`

The build tag `pprof` enables a pprof HTTP handler at route `/debug/pprof/` in
the `gostwire` service. As the pprof dumps contain sensitive data, never enable
this tag for productive builds.

### Go Standard Library

Not Gostwire build tags, but nevertheless important to us here, especially when
building a standalone binary (see also [Statically compiling Go
programs](https://www.arp242.net/static-go.html) for more background
information):

- `netgo`: activates a pure Go implementation for resolving DNS names at build
  time. Normally, Go only decides at runtime whether to use the C library
  functions or its own Go versions. Using the `netgo` build tag ensures that
  there isn't any C library reference anymore left in the binary.

- `osusergo`: activates a pure Go implementation for parsing `/etc/passwd` and
  `/etc/group` and skips using the C library functions for this. This switches
  off LDAP and/or NIS integration – which doesn't seem to be of any real loss on
  today's container hosts and workload nodes.

## Production

The "standard" service container deployment builds a static and stripped
`gostwire` binary:

```dockerfile
ARG LDFLAGS="-s -w -extldflags=-static"
ARG TAGS="osusergo,netgo"
RUN go build -tags="${TAGS}" -ldflags="${LDFLAGS}" ./cmd/gostwire
```

The build tags `netgo` and `osusergo` are required in order to build with only
the "pure Go" versions of the `net` and `os` without any libc dependencies.

## Performance Profiling

### Containerized

Using `make pprofdeploy` you can quickly switch to a pprof-enabled service,
exposing pprof information at `:5000/debug/pprof/` (please note the trailing
slash).

In this case, the deployed binary contains pprof support and the debugging
information won't be stripped from the (still static) binary. The only downside
here is that a `list` command in `pprof` cannot show source code, because that's
not available.

### Bare

If you need to show source code, then run instead:

```bash
# run gostwire service without container, at port :5000 and with pprof enabled.
make pprof
```
