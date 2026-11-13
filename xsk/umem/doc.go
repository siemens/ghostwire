/*
Package umem implements allocating “umem” (backed by shm) and mapping it into
user space, where “umem” memory is associated with a network interface for
efficient packet transmission and reception. Under best conditions (such as when
the netdev driver supports it), “umem” allows zero-copy receiving and sending
network packets.

Additionally, this package provides maintaining groups of “interested” umem
“partners” (eventually XDP sockets) using the same umem (slice) between them,
and unmapping the umem after the last [Partner] has ceased interest.

In this module, a [umem] is a region of virtual contiguous memory that is
allocated from a “shm” shared memory filesystem, and namely a tmpfs instance
mounted at /dev/shm. For the purposes of efficient communication via XDP sockets
(“xsk's”) this virtual contiguous memory is then later divided into equal-sized
so-called “frames”, but this is done by the xsk package.

On purpose, the design of this umem package hands out the allocated umem when
using [New] only in form of a “fd” file descriptor that API clients then can
pass around even between separate processes. However, passing the fd between
processes means transfer of ownership. In a separate step at the receiver's
side, the memory associated with this file descriptor then can be mapped using
[Map].

The rationale for this at first seemingly cumbersome handling is that this makes
it possible to move allocation and setup of umem into a separate privileged
manager process if so desired. User applications then can simply request a
configured XDP socket and associated ready-made umem from such a manager without
needing any special privileges.

# Important Notes

While the same umem can be attached to multiple XDP sockets simultaneously, the
umem-related fill and completion rings must still adhere to the
single-producer/single-consumer design. (The fill ring contains the umem chunks
to be used for receiving packets, while the completion ring contains umem chunks
that have been transmitted by the kernel and now can be reused.)

In consequence, these rings cannot be used by multiple processes simultaneously.

In addition, the rings package implements producing or consuming ring
descriptors without any additional locking, so ring access is no safe for use
from parallel go routine access. This design follows the design of the [libbpf
code].

This still holds while using the [Partner] objects, because the “partners” are
XDP sockets sharing the same umem for sending and receiving, handling the
lifecycle of the umem memory mapping.

[umem]: https://kernel.org/doc/html/v6.0/networking/af_xdp.html#umem
[libbpf code]: https://www.kernel.org/doc/html/next/networking/af_xdp.html#xdp-shared-umem-bind-flag
*/
package umem
