/*
Package podmannet implements a Gostwire decorator that discovers podman (v4+)
managed networks and then decorates their corresponding Linux-kernel network
interfaces. Supported types of podman networks are “bridge” and “macvlan”.

In case of “bridge” networks this decorator assigns network names as alias names
to the corresponding Linux-kernel bridges and also as a Gostwire-specific label.

For “MACVLAN” networks this decorator assigns the network names as alias names
to the “parent” network interface (or “master” in Linux parlance).

This decorator also copies any network labels it finds into the corresponding
network.Interface instances in a Gostwire discovery information model.

# Note

The Docker-compatible podman API is subtly incompatible: it uses a different
bridge name-allocating method, and it doesn't reveal the bridge and macvlan
master names.

In consequence, we need to resort to a self-rolled minimal HTTP-over-UDS client
that supports a minimal subset of the podman-proprietary libpod API. As of
podman v4 the libpod API endpoint returns network information. As a nice
benefit, the network information endpoint abstracts from the different podmen
networking mechanisms, that is, CNI-based and/or [netavark]-based.

[netavark]: https://github.com/containers/netavark
*/
package podmannet
