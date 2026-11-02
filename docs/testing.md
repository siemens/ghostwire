# Testing

Ghostwire's unit tests are written using [Ginkgo
v2](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/).
Many tests work with real network namespaces, network interfaces, and container
workloads, so they need a properly equipped test environment.

## Dev Container

The designated way to develop and test Ghostwire is the **fully privileged dev
container** in
[`.devcontainer/fully-privileged/`](https://github.com/siemens/ghostwire/tree/main/.devcontainer/fully-privileged).
It comes with everything needed for testing:

- Go and Node.js,
- its own Docker-in-Docker engine, so test containers never touch your host's
  Docker engine,
- `containerd`, `nerdctl`, and CNI plugins,
- host PID and cgroup namespaces, as well as an unconfined AppArmor profile, so
  that the discovery sees all processes and is allowed to switch namespaces,
- passwordless `sudo`, with `go` available to root, too.

## Running Tests

Always run the test suites **twice**: once as root and once as an ordinary user.
Many specs `Skip("needs root")` when run as an ordinary user, while other specs
check the non-root code paths. Neither run is a substitute for the other.

Also, never run test suites in parallel (`-p 1`), as suites share the same host
resources, such as the container engines.

Inside the dev container, simply run:

```bash
gocover
```

`gocover` runs all tests with the race detector and the
[`matchers`](buildndeploy#matchers) build tag, first as root and then as the
ordinary user, never in parallel. It then aggregates the coverage of both runs
into `coverage.html` and updates the coverage badge. You can limit the run to
specific packages, for example `gocover ./network`, or skip the root run using
`gocover --no-root`.

To additionally run the KinD-based tests that create a two-node
Kubernetes-in-Docker cluster (please note that this takes some time), enable the
[`kind`](buildndeploy#kind) build tag:

```bash
gocover --tags matchers,kind
```

For focused runs without coverage, use `go test` directly, keeping to the same
rules:

```bash
go test -v -race -p 1 -tags=matchers ./network
go test -v -race -p 1 -tags=matchers -exec 'sudo -n' ./network
```

Some specs additionally require a kernel with netkit support, the `netdevsim`
kernel module, or an SR-IOV capable NIC with at least one VF. These specs are
skipped when their requirements are not met. Similarly, the `podmannet`
decorator tests build a "podman-in-Docker" image; the first build of this image
takes some time.

## Continuous Integration

The GitHub ["build and
test"](https://github.com/siemens/ghostwire/blob/main/.github/workflows/buildandtest.yaml)
workflow doesn't use the dev container. Instead, it installs nerdctl and the CNI
plugins using `scripts/setup-nerdctl-and-friends.sh`, and then runs `go test -v
-p 1 -tags=matchers ./...`, once as root and once as an ordinary user. It does
not run the KinD-based tests.

## Writing Tests

- each package with tests has a `package_test.go` containing only the Ginkgo
  test suite plumbing, as well as test helpers shared by the package's specs.
- each `*_test.go` file with specs has exactly one top-level `Describe`.
- use [success](https://github.com/thediveo/success) helpers, such as
  `Successful(...)`, instead of separate error assertions.
- create test containers in `morbyd` sessions with
  `session.WithAutoCleaning(...)`, so that stale test containers are
  automatically removed; never remove them explicitly.
- skip specs needing root with `Skip("needs root")` when not running as root.

For the complete set of conventions, please see
[`.github/copilot-instructions.md`](https://github.com/siemens/ghostwire/blob/main/.github/copilot-instructions.md).
