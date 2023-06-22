# Repository

Following is a high-level overview about the G(h)ostwire repository, but we're
not going to list every nook and cranny.

`/`: the Gostwire repository root basically only contains the official
`Discovery` function, as well as the auto-generated semantic version of the
Gostwire module. The interesting stuff mostly lives in the sub directories.

- `api/`
  - `openapi-spec`: the OpenAPI3-based specification of the Ghostwire (sic!) v1
    REST API in `ghostwire-v1.yml`.
  - `v1/`: implements the JSON marshalling of Gostwire discovery results for the
    v1 REST API endpoints `/json` and `/mobyshark`. Please note that
    unmarshalling isn't supported.

- `cmd/`
  - `gostwire`: the real meat: this package implements the `gostwire` service. Use
    `make deploy` to build a containerized version and deploy it. Alternatively,
    use `make build` to build a standalone, stripped binary.
  - `lsallnifs`: "_list all network interfaces_" is more of a showcase CLI command
    that pretty-prints the discovery results to the console, clearly sorted by
    network namespace. It is kind of the spiritual successor to the v1 `gw show`
    command.
  - `gostdump`: simply dumps the discovery results as JSON to standard output, in
    REST API v1 discovery result format. Neither needs a running service nor
    `wget`.

- `docs/`: the Gostwire documentation; use `docsify serve docs` in the repo's
  root directory to serve and view this documentation. This documentation is
  self-contained and can be served from any ordinary HTTP server (the `docsify`
  helper is just a simple plain HTTP server).

- `deployments/`: contains the Gostwire service deployment information. See
  [Build & Deploy](buildndeploy) for more details.

- `scripts/`: some helper scripts.

- `decorator/`: implements so-called "[decorators](terminology#decorator)" that
  add useful (usually user-space) information to the discovered networks and
  containers.
  - `all/`: pulls in all defined Gostwire decorator modules, so they register
    themselves.
  - `dockernet/`: discovers Docker-managed networks and adds their names as
    alias names to the corresponding Linux network interfaces.
  - `ieappicon/`: discovers the IE App-specific icons of IE Apps from an IED's
    appengine data base (via the Docker compose project name).
  - `nerdctlnet/`: discovers nerdctl-managed CNI networks and adds their names
    as alias names to the corresponding Linux network interfaces.

- `metadata/`: implements the plugin-based discovery metadata mechanism. Plugins
  can discover and retrieve discovery meta-information, such as the host OS name
  and version, Industrial Edge core/runtime sem version, et cetera.
  - `all/`: pulls in all defined Gostwire metadata modules, so they register
    themselves.
  - `engines/`: gathers metadata about the container engines for which
    containers were discovered. Container engines without any workload won't be
    listed.
  - `host/`: gathers metadata about the host Gostwire operates on, such as:
    hostname, OS name and version.
  - `iecore/`: gathers metadata about an Industrial Edge core/runtime, such as:
    semantic version.

- `mobydig/`: integrates our [mobydig
  module](https://github.com/siemens/mobydig) into G(h)ostwire.

- `network/`: implements the network topology and configuration discovery, that
  is: network interfaces, their relationships, address configuration, route
  configuration, open transport ports, et cetera. Please note that the generic
  Linux-kernel process, namespace and container discovery is off-loaded to the
  [@thediveo/lxkns](https://github.com/thediveo/lxkns) and
  [@thediveo/whalewatcher](https://github.com/thediveo/whalewatcher). Thus, this
  package solely focuses on the Gostwire-specific network aspects.

- `turtlefinder/`: implements our container engines "anywhere" technology:
  Gostwire automatically detects container engines by their running processes,
  including the hierarchy, such as used by containers in containers
  (Kubernetes-in-Docker, et cetera). Currently auto-detects `dockerd` and
  `containerd`.
  - `detector/moby`: auto-detect Docker daemons.
  - `detector/containerd`: auto-detect containerd daemons.

- `util/`: internal utilities, such as Gomega matchers to simplify unit test
  `Expect`ations.

- `test/`:
  - `kind/`: [KinD](https://github.com/kubernetes-sigs/kind)-based tests. These
    _do not_ require the `kind` binary installed as they directly use KinD's Go
    API instead.

- `webui/`: opens up into the totally different world of G(h)ostwire's web user
  interface using all that crazy web-stuff like Javascript/Typescript, React, et
  cetera. The web user interface is a so-called "[single-page
  application](https://en.wikipedia.org/wiki/Single-page_application)" (SPA)
  that solely runs in the user's browser and accesses discovery information
  using the REST API.

- `media/`: mainly the mascot graphics.

- `internal/`: Gostwire-"internal" stuff not for general direct consumption.
