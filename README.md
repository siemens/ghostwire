<img align="right" width="100" height="100" src="media/gostwire-mascot-darklight-150x150.png" style="padding: 0 0 1ex 0.8em">

[![Siemens](https://img.shields.io/badge/github-siemens-009999?logo=github)](https://github.com/siemens)
[![Industrial Edge](https://img.shields.io/badge/github-industrial%20edge-e39537?logo=github)](https://github.com/industrial-edge)
[![Edgeshark](https://img.shields.io/badge/github-Edgeshark-003751?logo=github)](https://github.com/siemens/edgeshark)

# Ghostwire Virtual Communication Discovery

[![PkgGoDev](https://pkg.go.dev/badge/github.com/siemens/ghostwire/v2)](https://pkg.go.dev/github.com/siemens/ghostwire/v2)
[![GitHub](https://img.shields.io/github/license/siemens/ghostwire)](https://img.shields.io/github/license/siemens/ghostwire)
![build and test](https://github.com/siemens/ghostwire/workflows/build%20and%20test/badge.svg?branch=main)
![goroutines](https://img.shields.io/badge/go%20routines-not%20leaking-success)
![file descriptors](https://img.shields.io/badge/file%20descriptors-not%20leaking-success)
[![Go Report Card](https://goreportcard.com/badge/github.com/siemens/ghostwire/v2)](https://goreportcard.com/report/github.com/siemens/ghostwire/v2)
![Coverage](https://img.shields.io/badge/Coverage-72.6%25-yellow)

**G(h)ostwire** discovers the virtual (or not) network configuration inside
_Linux_ hosts – and can be deployed as a REST service or consumed as a Go
module. It comes with its unique container engine discovery that auto-discovers
multiple container engines in hosts, including engines _inside containers_.
Ghostwire not only understands containers, but also Composer projects and
Kubernetes pods.

Currently supported container engines are:
- [containerd](https://containerd.io), 
- [Docker](https://docker.com),
- [CRI-O](https://cri-o.io),
- [podman](https://podman.io) – when set up to be socket-activated by `systemd`
  (see also [podman Quick Start: Starting the service with
  systemd](https://github.com/containers/podman/blob/main/pkg/bindings/README.md#quick-start)).
  Please note that we only support the Docker-compatible API, but not the
  podman-proprietary workload features, such as podman pods.

Ghostwire is also Kubernetes-aware and even understands that KinD simulates
Kubernetes nodes in Docker containers.

## Quick Start

We provide multi-architecture Docker images for `linux/amd64` and `linux/arm64`.
First, ensure that you have the Docker _compose_ plugin v2 installed. For Debian
users it is strongly recommended to install docker-ce instead of docker.io
packages, as these are updated on a regular basis.

Make sure you have a Linux kernel of at least version 4.11 installed, however we
highly recommend at least kernel version 5.6 or later.

```bash
wget -q --no-cache -O - \
  https://github.com/siemens/ghostwire/raw/main/deployments/wget/docker-compose.yaml \
  | docker compose -f - up
```

Finally, visit http://localhost:5000 and start looking around the virtual
container networking, IP and DNS configuration, open and forwarded ports, and
much more.

> ⚠ This quick start deployment will **expose TCP port 5000** also to clients
> external to your host. Make sure to have proper network protection in place.

## Eye Candy!

### Lots of Virtual "Wires"

A slightly busy Industrial Edge host:

![Edgeshark wiring screenshot](media/edgeshark%20screenshot.png)

### IP & DNS Configuration

...lots of gory configuration details:

![Edgeshark details screenshot](media/edgeshark%20screenshot%20details.png)

For instance: container/process capabilities, with Docker non-standard
capabilities being highlighted. Open and connected ports, and forwarded ports,
all with the containers and processes serving them. Other Docker containers
addressable from your container using DNS names for Docker services and
containers.

## See the Current System State

Information is gathered as much as possible from Linux APIs in order to show the
current _effective_ state, instead of potentially invalid or stale engineering
configuration. Container engines are only queried in order to understand which
processes and network elements have a user-space meaning beyond the kernel-space
view.

Teaming up Ghostwire with [Packetflix](https://github.com/siemens/packetflix)
makes [Wireshark](https://www.wireshark.org/) container-aware: a simple click on
one of the "shark fin" buttons starts a Wireshark session, directly capturing
network traffic inside a container (or even [KinD](https://kind.sigs.k8s.io/)
pod).

## Edgeshark Project Context

`ghostwire/v2` is part of the "Edgeshark" project that consist of several
repositories:
- [Edgeshark Hub repository](https://github.com/siemens/edgeshark)
- 🖝 **G(h)ostwire discovery service** 🖜
- [Packetflix packet streaming service](https://github.com/siemens/packetflix)
- [Containershark Extcap plugin for
  Wireshark](https://github.com/siemens/cshargextcap)
- support modules:
  - [turtlefinder](https://github.com/siemens/turtlefinder)
  - [csharg (CLI)](https://github.com/siemens/csharg)
  - [mobydig](https://github.com/siemens/mobydig)
  - [ieddata](https://github.com/siemens/ieddata)

## Documentation

For deployment, usage and development information, please see the accompanying
documentation in `docs/`. The most convenient way to view this documentation is
with the help of [docsify](https://docsify.js.org/):

```bash
make docsify
```

Then navigate to [http://localhost:3300](http://localhost:3300) to read on.

## Build & Deploy

Building the Packetflix service requires the Go toolchain, `make`, a C compiler
(used by cgeo), and finally Docker installed.

The preferred build method is as a containerized service. For development, a
standalone binary as well as the webui assets can be built outside of any
container.

> **IMPORTANT:** the Ghostwire service exposes (read-only) host information on
> port 5000. Never expose this services without taking additional security
> measures.

### Container Service

To build and deploy the Gostwire containerized service and exposing it at port
5000 of the host (requires the build(x) plugin to be installed, which is GA on
Linux for some time now):

```bash
make deploy
```

Alternatively, you can build and deploy a pprof-enabled service, which enables
pprof information at `:5000/debug/pprof`.

```bash
make pprofdeploy
```

For more information, please refer to the (docsified) documentation in `docs/`.

### Standalone Service Binary

- ensure that Go 1.20 (or later) is installed.
- ensure that [nodejs](https://nodejs.org/en/) 16+,
  [npm](https://docs.npmjs.com/downloading-and-installing-node-js-and-npm), and
  [yarn](https://classic.yarnpkg.com/en/docs/install/#debian-stable) (`npm
  install --global yarn`) are installed.

```bash
make yarnsetup # if not done already so
make build-embedded
./gostwire --debug
```

## VSCode Tasks

The included `gostwire.code-workspace` defines the following tasks:

- **View Go module documentation** task: installs `pkgsite`, if not done already
  so, then starts `pkgsite` and opens VSCode's integrated ("simple") browser to
  show the csharg documentation.

#### Aux Tasks

- _pksite service_: auxilliary task to run `pkgsite` as a background service
  using `scripts/pkgsite.sh`. The script leverages browser-sync and nodemon to
  hot reload the Go module documentation on changes; many thanks to @mdaverde's
  [_Build your Golang package docs
  locally_](https://mdaverde.com/posts/golang-local-docs) for paving the way.
  `scripts/pkgsite.sh` adds automatic installation of `pkgsite`, as well as the
  `browser-sync` and `nodemon` npm packages for the local user.
- _view pkgsite_: auxilliary task to open the VSCode-integrated "simple" browser
  and pass it the local URL to open in order to show the module documentation
  rendered by `pkgsite`. This requires a detour via a task input with ID
  "_pkgsite_".

## Make Targets

- `make`: lists all targets.
- `make dist`: builds snapshot packages and archives of the csharg CLI binary.
- `make pkgsite`: installs [`x/pkgsite`](golang.org/x/pkgsite/cmd/pkgsite), as
  well as the [`browser-sync`](https://www.npmjs.com/package/browser-sync) and
  [`nodemon`](https://www.npmjs.com/package/nodemon) npm packages first, if not
  already done so. Then runs the `pkgsite` and hot reloads it whenever the
  documentation changes.
- `make report`: installs
  [`@gojp/goreportcard`](https://github.com/gojp/goreportcard) if not yet done
  so and then runs it on the code base.
- `make vuln`: install (or updates) govuln and then checks the Go sources.

# Contributing

Please see [CONTRIBUTING.md](CONTRIBUTING.md).

## License and Copyright

(c) Siemens AG 2023

[SPDX-License-Identifier: MIT](LICENSE)
