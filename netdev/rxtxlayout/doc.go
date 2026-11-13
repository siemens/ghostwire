/*
Package rxtxlayout models the “layout” or “structure” of Linux netdevs in terms
of their RX/TX queues, NAPI instances, and IRQs. While this model on purpose has
cycles involving queues, NAPIs, and IRQs, it can still be properly un/marshalled
from/to JSON, thanks to internally using custom un/marshalling methods.

Our model bases on the [NETLINK netdev API] information model and is able to
fully represent it. Additionally, it provides relationships between the model
elements as pointers and thus is not “just” the NETLINK netdev API data
structures.

Please note that this package does not provide any discovery functionality, but
“only” access to the “netdev” family netlink protocol (API). Instead,
discovery beyond simple single known network namespace netdev queries is the
responsibility of other packages (which then build upon the primitives provided
by this package).

[NETLINK netdev API]: https://www.kernel.org/doc/html/latest/netlink/specs/netdev.html
*/
package rxtxlayout
