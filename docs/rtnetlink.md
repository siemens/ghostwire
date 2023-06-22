# rtnetlink

[rtnetlink](https://man7.org/linux/man-pages/man7/rtnetlink.7.html) allows
user-space programs to communicate with network-related subsystem in the Linux
kernel. It allows to query and configure network routes, IP addresses, link
parameters, neighbor setups, queueing disciplines, traffic classes and packet
classifiers.

- `IFLA_MASTER`
- `IFLA_SLAVE`

- `NETNSA_FD`: network namespace IDs

> [!ATTENTION] In case a MACVLAN master interface or VETH peer interface has –
> by sheer coincidence – the same index number as the index of the MACVLAN or
> VETH interface itself, then rtnetlink **will drop the `IFLA_LINK` attribute**.
> Gostwire thus needs to handle this special case and regenerate the index
> number for the MACVLAN or VETH interface itself. _Bummer_.

## Network Indices

Network interface indices cannot be zero. This can be deduced, for instance,
from
[if_nametoindex(3)](https://man7.org/linux/man-pages/man3/if_nametoindex.3.html).
