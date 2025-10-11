/*
Package rings provides XDP socket/umem single-producer/single-consumer RX, TX,
fill and complete rings. The concrete user application-facing [Ring]-derived
types are:

  - [Fill] ring associated with [umem].
  - [Completion] ring associated with umem.
  - [Rx] ring optionally associated with an [XDP socket].
  - [Tx] ring optionally associated with an [XDP socket].

Additionally, this package also offers a convenience descriptor chunk (address)
pool manager, see [NewDescriptorPool]: this pool can store descriptor chunk
(addresses) currently not in any of the four rings.

# Usage

User applications usually won't use the constructor functions for specific
[Ring]-derived types directly, but instead retrieve ring objects from XDP Socket
objects instead. Thus, the most prevalent API elements are probably:

  - [ProducerRing.Add] adds a descriptor to the ring.
  - [ConsumerRing.Next] removes the next available descriptor from the ring and
    returns it.

Please note that there are two different types of descriptors, namely:

  - uint64-typed descriptors are used with the [Fill] and [Completion] rings, as
    we just work with the whole chunk.
  - [unix.XDPDesc] is used instead with the [Rx] and [Tx] rings, as here we need
    not only the packet offset into the [umem], but also the actual length
    inside its chunk and maybe some option flags (such as [unix.XDP_PKT_CONTD]
    for [multi-buffer support]) on top of this.

# Passing Rings Around

As the actual ring state information is held in kernel-allocated memory mapped
into user space, [Ring] objects can easily passed around by value rather than
reference.

# Single Producer/Single Consumer Ring Architecture

Did we mention that [Ring] objects are strictly single-producer/single-consumer,
so while you cannot use the same Ring (or a copy thereof) from different go
routines concurrently without implementing your own synchronization?

But then the overhead of additional locking might defeat the performance aspect
of XDP rings ... erm, in Go ... “nevermind”.

[umem]: https://docs.kernel.org/networking/af_xdp.html#umem
[XDP socket]: https://docs.kernel.org/networking/af_xdp.html#overview
[multi-buffer support]: https://docs.kernel.org/networking/af_xdp.html#multi-buffer-support
*/
package rings
