/*
Package network defines Gostwire's virtual network configuration and topology
model and implements the discovery.

Network configuration covers (not only) IP address and route configuration, but
also “netstat on drugs”: discovery of open ports and then relating them not only
to processes (as “netstat” and “showsocket” do; the latter has a condemnable
short name without any understanding of history, really) but also to containers.

The topology model bases on conveniently post-processed RT netlink data, taking
care of netlink glitches especially with interface indices in other network
namespaces. The in-memory model allows for quick topology traversal using simple
pointer dereferencing.

And finally, Gostwire is fully dual-stack IPv4/IPv6 aware.

# Basics

In any case, any Gostwire discovery turns up a set of network namespaces,
represented by type NetworkNamespace. Please note that network namespaces are
flat and thus do not form any kind of hierarchy.

Each NetworkNamespace lists the network interfaces assigned to it, where the
interfaces are represented by the aptly-named type Interface. NetworkNamespace
also holds the IP routes (sic!) as well as open transport ports. Finally, a
NetworkNamespace lists the “tenants” attached to this network stack. Here, a
“tenant” is either a process at the top of a process sub-tree, still attached to
this network stack, or the initial process of a container.

Depending on system configuration and process state, multiple tenants might be
present. A most prominent example are the containers of a single Kubernetes pod
sharing the same single network stack between themselves.

Only Interface provides information about the configured (or “assigned”) IP
addresses. In contrast, routes, belong to a network stack and thus a
NetworkNamespace, but not to any Interface.
*/
package network
