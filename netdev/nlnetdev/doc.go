/*
Package nlnetdev provides access to netdev-related queue, IRQ, and NAPI
information via the Linux kernel's NETLINK API. See also the [netdev NETLINK
kernel specification].

# NETLINK Netdev Model

In the following sections, let's look at the main elements we're dealing with
int this package. Please note that the netdev NETLINK API covers more ground
that is out of scope for this package (at least for now).

This package provides two discovery entries, depending on the scope of discovery
needed:

  - [Discover] takes an [lxkns] (network) namespace and process discovery and
    then discovers the netdevs with their queue, IRQ, and NAPI configurations.
  - [Conn.Netdevs] covers only a the single network namespace the connection
    is leading to. Please note that this discovery function does not return
    interface naming information (whereas [Discover] does).

The main discovery method after establishing an API connection using [Dial] is
[Conn.Netdevs]: this returns [rxtxlayout.Netdev] objects for which queues, IRQ,
and NAPI information is available.

# Queues

Netdevs can be queried using [Conn.Queues] for information about their RX and TX
queues, for details please also see [netdev NETLINK queue specification]:

  - type of queue: RX or TX,
  - queue ID, as need when creating XDP sockets and binding them to queues.

The NETLINK-based API is better suited in many more complex use cases, where
access to the correct [sysfs(5)] instance can be challenging. In particular,
this API returns the correct information based on the currently active network
namespace of the caller's OS-level thread.

When querying a network namespace different from the caller's process network
namespace, make sure to lock the OS-level thread to the calling go routine when
creating the API connection using [Dial]. The go routine can be unlocked and the
original network namespace restored after Dial returned, as the NETLINK
connection is then fixed to the network namespace active when creating the
connection.

# NAPIs

Netdevs can be queried using [Conn.NAPIs] and [Conn.NifsNAPIs] for information
about their NAPI(s), for details please see [netdev NETLINK NAPI specification]:

  - NAPI ID
  - IRQ, where available – please note that unfortunately “some” drivers do not
    fill in all the API information, so make sure to check [rxtxlayout.Queue.HasIRQ]
    first.
  - the PID of an optional NAPI kthread, when enabled for the NAPI/netdev.

The NAPI ID and kthread PID cannot be retrieved via the [sysfs(5)]. However, the
sysfs might provide IRQ information even when the NETLINK netdev NAPI API
doesn't return IRQ information.

The same rules for network namespace awareness of this API apply, as with the
netdev queues API.

# NAPI Kthreads

NAPI kthreads can be enabled and disabled via the “threaded” pseudo files of
network interfaces in the sysfs. However, while this creates the kthread(s) upon
enabling, disabling does not stop/kill the NAPI kthreads. See also [NAPI thread
creation].

And before you ask: the NAPI kthreads don't respond to SIGTERM nor SIGKILL.

# IRQs

Please note that IRQ kernel thread information is filled in from the supplied
process table [model.ProcessTable] when calling [Conn.Netdevs].

# Further References

  - [LWN: Introduce queue and NAPI support in netdev-genl]

[netdev NETLINK kernel specification]: https://www.kernel.org/doc/html/latest/networking/netlink_spec/netdev.html
[netdev NETLINK queue specification]: https://www.kernel.org/doc/html/latest/networking/netlink_spec/netdev.html#queue
[netdev NETLINK NAPI specification]: https://www.kernel.org/doc/html/latest/networking/netlink_spec/netdev.html#napi
[sysfs(5)]: https://man7.org/linux/man-pages/man5/sysfs.5.html
[NAPI thread creation]: https://elixir.bootlin.com/linux/v6.9.5/source/net/core/dev.c#L6430
[LWN: Introduce queue and NAPI support in netdev-genl]: https://lwn.net/Articles/948273/
[lxkns]: https://github.com/thediveo/lxkns
*/
package nlnetdev
