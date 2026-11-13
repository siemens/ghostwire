/*
Package discover provides discovering XDP sockets and their relationships to
processes and containers, and their umems. See the [AllXSKs] function in this
package.

While the NETLINK_INET_DIAG family dumps diagnosis information about all XDP
sockets in a particular network namespace (please note that the kernel doesn't
implement filtering) it doesn't return any information about how these XDP
sockets are relating to processes. This is where package discover connects the
dots.

Please note that the XSK discovery needs to scan the process file system, and
the open file descriptors of all processes: it gathers socket inode numbers to
match them against the XSK inode numbers we found in the dump.

Additionally, [ExtractUmems] recreates the unique umems potentially shared by
multiple XDP sockets.

Please note that it is not possible to discover the umem memory itself; that is,
it is not possible to map the registered umem memory into user space memory,
given only an XDP socket file descriptor and/or the NETLINK XDP dump API.

# Usage

In its simplest form, XDP socket and umem diagnosis information can be gathered
(without container discovery) as follows:

	import (
	    "github.com/siemens/ghostwire/v2/xsk/discover"
	    "github.com/thediveo/lxkns"
	)

	disco := lxkns.Namespaces(
	    lxkns.WithStandardDiscovery(),
	    lxkns.FromTasks(),
	)
	xsks := discover.AllXSKs(disco)
	umems := discover.ExtractUmems(xsks)
*/
package discover
