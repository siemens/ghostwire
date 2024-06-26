# Gostwire

![G(h)ostwire mascot](media/gostwire-gr.png ':class=gwmascot')

**G(h)ostwire** is (what we assume) a unique virtual network topology and
configuration discovery engine for Linux hosts. It discovers the virtual layer 2
topology, as well as network address+route configuration, open transport-layer
ports, port forwarding, and even DNS (stub) resolver configurations. It then
correlates the networking with the containers found in a Linux system.

For the Linux geeks: **G(h)ostwire** = _namespaces_ + _RT netlink_ +
_containers_.

Our Go-based discovery engine is written in Go and offers good performance,
widely use of Go routines for better parallelism, good long-term
maintainability, and good extensibility.

## Special Highlights

- **"turtles anywhere" container engine discovery** and **multi-engine
  support**: automatically detects container engines, not only in the host, but
  also inside containers. This makes Gostwire a good analysis and diagnosis tool
  not just for [Kubernetes-in-Docker
  (KinD)](https://github.com/kubernetes-sigs/kind) configurations, but also for
  engine side-by-side deployments, et cetera. Supported engines:
  - Docker
  - "plain" containerd (such as with
    [`nerdctl`](https://github.com/containerd/nerdctl))

- **fully parallel service request handling**, including discovery. This speeds
  up especially the initial web user experience as the web ui-related HTTP
  requests finally can be handled in parallel to an in-flight discovery request.
  While this would have been possible with the Python-based engine, this would
  have required a separate HTTP proxy instance for handling the static web asset
  request, much increasing memory footprint as well as code base maintenance
  efforts. The integrated Go HTTP server just runs like Hades.

- **background tracking of the container workload**, reducing discovery request
  load and instead spreading the container-related discovery over a long time as
  it happens.

- the discovery service is just a **single static binary**.

## Minimal Kernel Requirement

Make sure you have a Linux kernel of at least version 4.11 installed, however we
highly recommend at least kernel version 5.6 or later.

## Name and Mascot

The name "G(h)ostwire" sprang from the view of virtual (VETH) wires somehow
belonging to the (&#42;_cough_&#42;,
[ethereal](https://www.wireshark.org/news/20060607.html)) world of ghosts. As a
nod to the implementation language our mascot is a Go Gopher under a fake Ghost
(Specte) blanket. [Spectre
v1](https://en.wikipedia.org/wiki/Spectre_(security_vulnerability)) might have
been influential, too.
