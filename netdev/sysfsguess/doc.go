/*
Package sysfsguess provides guesswork information about netdev queues and NAPI
configuration, including NAPI threads.

While kernel 6.8 saw the introduction of the [netdev NETLINK specification] and
protocol, the situation still is rather lacking (to put it mildly). The netdev
NETLINK API basically exposes all the dirt swept under the NIC driver carpets:
at the time of this writing, there are only 2 (in words: “two”) NIC drivers that
correctly fill in the IRQ information after NAPI registration.

The netdev NETLINK API has been explicitly introduced to avoid the “guesswork”
necessary when trying to figure out queue IDs, IRQs, NAPI instances, and related
NAPI kernel threads. The sysfs was never designed to provide this information,
and not even in a consistent manner across different NIC device drivers.

To add insult to injury, the network namespace-related sysfs view isn't dynamic
at all; whoever looks at it will always see the same frozen network namespace
view that the mounting process had. A particular sysfs instance will always show
the same network namespace.

The limitation thus is that we won't be able to correctly gather information for
network namespaces that don't have a matching sysfs instance mounted somewhere.
However, a large number of use cases for network namespaces are (OCI) containers
and these have proper sysfs instances set up. So, as long as we know the
“ealdorman” (to use [lxkns] terminology) process for a network namespace we can
assume that it has a proper mount namespace set up with a proper sysfs instance,
and we can access it via the “/proc/$PID/root” links in the procfs.

[netdev NETLINK specification]: https://www.kernel.org/doc/html/latest/networking/netlink_spec/netdev.html
[lxkns]: https://github.com/thediveo/lxkns
*/
package sysfsguess
