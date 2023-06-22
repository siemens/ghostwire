# netlink Go Module

This part of the documentation sheds some more information on using
[@vishvananda/netlink](https://github.com/vishvananda/netlink) in Gostwire,
especially when it comes to idiosyncrasies rooted either in the Go `netlink`
module or Linux' netlink implementation.

On a side note, @vishvananda/netlink inherits quite some ideas from the famous
Python [pyroute2](https://github.com/svinota/pyroute2) package (albeit it does
not attempt to be or become the Go twin to pyroute2).

## Link Type

Catch: unknown "virtual" link kinds result in `link.Type() == "device"`. This
includes the loopback interface `lo`.

## L2 Relation Attributes

The following network interface-related netlink attributes are of special
interest to Gostwire, because they describe the topology of the (virtual) data
link layer 2.

| netlink Attribute… | …ends up in `netlink.Link` Attribute |
| --- | --- |
| `IFLA_LINK` | `ParentIndex` (but see note below) |
| `IFLA_LINK_NETNSID` | `NetNsID` |
| `IFLA_MASTER` |  `MasterIndex` |
| `IFLA_INFO_SLAVE_KIND` | needs to be derived from value of `Slave` |
| `IFLA_INFO_SLAVE_DATA` | `Slave` |
| `IFLA_VXLAN_LINK` |  `VtepDevIndex` |

> [!ATTENTION] In case the MACVLAN master interface or VETH peer interface has –
> by sheer coincidence – the same index number as the index of the MACVLAN or
> VETH interface itself, then rtnetlink **will drop the `IFLA_LINK` attribute**.
> In turn, `ParentIndex` will be the zero value and is thus detectable, because
> network interfaces indices cannot be zero (otherwise,
> [if_nametoindex(3)](https://man7.org/linux/man-pages/man3/if_nametoindex.3.html)
> would be toast.)

So when and where do these attributes appear on network interfaces?

| When/Where? | Relation | netlink Attribute(s) | `netlink.Link` Attribute(s) |
| --- | --- | --- | --- |
| `bridge` | ports | | needs to be derived from `ParentIndex` of enslaved port network interfaces |
| bridge "port" | bridge | `IFLA_MASTER` | `MasterIndex` |
| `macvlan` | master(!) | `IFLA_LINK`+`IFLA_LINK_NETNSID` | `ParentIndex`+`NetNsID` |
| macvlan "master" | slaves | | needs to be derived from `ParentIndex` of enslaved macvlans |
| `veth` | peer | `IFLA_LINK`+`IFLA_LINK_NETNSID` | `ParentIndex`+`NetNsID` |
| `vxlan` | peer | `IFLA_VXLAN_LINK`+`IFLA_LINK_NETNSID` | `VtepDevIndex`+`NetNsID` |
| `tap`/`tun` |  |  |

## L2 Master and Parent

First things first, `ParentIndex` is a misnomer (but established Linux
terminology). It is used in these diverse contexts:

- `macvlan`: `ParentIndex` and `NetNsID` refer to the _master_ network
  interface.
- `veth`: `ParentIndex` and `NetNsID` refer to the _peer_ network interface.

In consequence, `MasterIndex` isn't always used where Linux actually uses the
master/slave terminology, notably for bridge ports.

- bridge "port": `MasterIndex` refers to be _bridge_ network interface, and is
  always used **without NetNsID** as the port and bridge must be in the same
  network namespace.

And now please attempt to get Depeche Mode's lyrics out of your head again. Or,
even better, don't.
