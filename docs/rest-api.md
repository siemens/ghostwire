# REST API

> Are you looking for Gostwire's [Go API](api) instead?

G(h)ostwire's <a href="api/index.html" target="_blank">REST API endpoints</a>
(opens documentation in new tab) are defined in the OpenAPI v3 specification in
`api/openapi-spec/ghostwire-v1.yaml`. API unit tests check against this API
specification.

- `/json`: complete discovery information, as used, for instance, by the web UI.
  
  At this time, there is only a single optional query parameter `?ieappicons`
  (implemented by cmd/gostwire/v1endpoints.go). If this query parameter is
  present, then the parameter's value is passed via the "discovery labels" as
  part of the Ghostwire discovery options. In particular, the
  `ieappicon.IEAppDiscoveryLabel` is set to the specified parameter value (even
  if just empty `""`). This discovery option (label) instructs the
  `decorator/ieappicon` decorator as follows:
  
  - `?ieappicons` not present: skip the `decorator/ieappicon` decorator for the
    particular discovery operation.
  - `?ieappicons` present with a value other than `off`: return IE App icons for
    the discovered containers. The App icon information is in form of PNG data
    URLs stored in a label `gostwire/icon` (constant named
    `ieappicon.IconLabel`) of a particular container.
  - `?ieappicons=off`: skip the `decorator/ieappicon` decorator for the
    particular discovery operation.

- `/mobyshark`: discovery information only about "capture targets", that is, the
  pod, containers, processes, et cetera, with network interfaces for which
  network captures could be taken. For instance, this excludes network topology
  information and configuration details of network interfaces (no need for IP
  addresses, ...).
