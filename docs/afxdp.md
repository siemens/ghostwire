# XDP and XSK

Ghostwire's `xsk` package provides a pure-Go[^1] implementation to access and
diagnose [XDP](https://en.wikipedia.org/wiki/Express_Data_Path) sockets
(`xsk.Socket`). Here, "XDP" is the so-called "_express data path_" that bypasses
most of the Linux networking stack in order to allow user space to send and
receive packets at high rates or low latencies. In the best case, user space
processes can send and receive data-link layer packets with zero copy using
remote DMA, memory-mapped packet areas and ring buffers.

Because time is crucial, the term "_XDP socket_" usually gets shortened to just
"_XSK_".

Please note that our package is not meant for hard realtime or high-end
performance productive use, but instead for testing, diagnosing, and
experimenting.

The `xsk/discovery` package provides discovering XSKs of processes in any
network namespace that is discoverable by the underlying
[lxkns](https://github.com/thediveo/lxkns) namespace discovery engine. This
complements the existing IP transport-layer TCP and UDP sockets discovery.

The discovery reveals not only the network interface an XSK is bound to, but
also which (RX and/or TX) queue, as well as the size of the umem assigned and
the individual ring buffer sizes (for RX, TX, fill, and completion). Also packet
drops as well as ring under- and overflows are also reported.

Further core elements when dealing with XSKs are "umem" – maybe derived from
"user memory", but never explained in the kernel documentation – and "rings".
The `xsk/umem` package simplifies allocating suitable shared memory (backed by
shm) and mapping it into the Go application's address space. This umem is then
used as packets buffer to send and receive packets.

And this then brings us to the `xsk/rings` package which implements the
necessary single-producer/single-consumer RX, TX, fill and complete rings to
keep track of packet buffers. Most important, the `xsk/rings` package hides the
ugly `unsafe` dances in order to deal with the kernel-provided rings after
mapping them into Go user space.

As a fun fact, the (non-exported) generics `xsk/rings.ring[Descriptor]` type
helped alot to keep the code straightforward and good to maintain. The rest is
behavioral inheritance using composition (embedding). The user-code facing types
are `xsk/rings.Tx`, `xsk/rings.Rx`, `xsk/rings.Fill` and `xsk/rings.Completion`.

A highly recommended background read on how the Linux kernel's rings work is
Juho Snellman's blog post [I've been writing ring buffers wrong all these
years](https://www.snellman.net/blog/archive/2016-12-13-ring-buffers/).

#### Notes

[^1]: absolutely no `libbpf` and cgo here.
