# NIC Passthrough

In its original meaning, "_NIC passthrough_" is a virtualization feature that
allows a physical _Network Interface Card_ (NIC) to be assigned directly to a
virtual machine (VM). In the context of container engines, such as Docker, it
refers to moving a network interface out of the host network namespace into the
network namespace of a container.

NIC passthrough in Docker has its uses, such as in industrial automation on the
Siemens Industrial Edge runtime platform: it provides hard realtime control
applications an "isolated" network interface over which the containerized
control application has full control, without noisy neighbors such as the host
and other containers interfering with the control traffic.

## Tracing NIC Roaming

For Docker, any network interface that is passed through into a container has
its name changed, not least to conform to the rules imposed by Docker's
`libnetwork`: in order to prevent suffix naming conflicts, `libnetwork` adds
`0`, `1`, and so on, hence the infamous `eth0`.

To help humans and agents diagnose the communication topology, as well as for
its own housekeeping, the passthrough Docker network driver of the Siemens
Industrial Edge runtime attaches details to passed through network interfaces
using Linux' network interface alternate names – not to be confused with the
older mechanism of alias names.

Now, network interface alternate names aren't label, but are very lenient when
it comes to the ASCII characters allowed: almost all work, but for some reason
the famous iproute2 tool suite (`ip`) chokes on "/" slashes. Dashes and minus
signs are perfectly fine though.

So in order to trace NIC "roaming" the following alternate name prefixes are
used, with the concrete information attached:

- `siemens.iedge.orig-ifname=` with the name of the network interface in the
  host network namespace appended. More precisely, the "host network namespace"
  is the network namespace where a particular Docker daemon is attached to.
- `siemens.iedge.docker-nw=` with the ID of a Docker network appended, as well
  as sepearated by "." another piece of information that the iedge
  passthrough-driver of the Industrial Edge runtime uses for its internal
  housekeeping purposes.

## Surfacing Roaming Details in Ghostwire

Ghostwire detects and surfaces details about passed-through NICs as follows:

- the alias name of a Ghostwire network interface receives the name of the
  Docker network that drives the pass-through. This is in line with using the
  alias detail for surfacing the names of Docker networks.
- the labels of a Ghostwire network interface carry additional labels:
  - `gostwire/interface/orig-name`: the "original" name of a network interface
    when in the host network namespace.
  - `gostwire/network/name`: the name of the driving Docker network; please note
    that this label is also present for network interfaces in the host network
    namespace that belong to Docker networks, like the bridge network
    interfaces.
  - `gostwire/network/id`: the Docker network ID for the custom network the
    passed-through network interface belongs to; please note that this label is
    also present for network interfaces in the host network namespace that
    belong to Docker networks, like the bridge network interfaces.
