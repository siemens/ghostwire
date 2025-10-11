/*
Package xsk provides XDP sockets (“xsks”) for unit testing XDP socket diagnosis
and characteristics.

Please note that the XDP [Socket] type provided in this package is not meant for
hard realtime and/or high-end performance productive use, but instead for
testing and experimenting. Your mileage may vary.

# Introduction and Recap of Linux XDP Socket Concepts

The following is a condensed recap of the core concepts introduced in the [Linux
kernel networking documentation on AF_XDP].

# Umem

[umem] is any region of virtual contiguous memory (registered with the kernel)
divided into equally sized “chunks”. Chunks inside an umem are identified by
their offsets into the umem.

Umem can be shared between multiple XDP [Socket] objects, as long as these
sockets bind to the same network interface and queue. The same umem cannot be
shared between XDP sockets for different queues or network interfaces.

# Lords of the Rings

XDP sockets and [umem] make use of four so-called “rings”. These four rings are
all single producer and single consumer rings.

  - [rings.Fill] (umem): transfers ownership of umem chunks/frames from user
    space to kernel space so that the kernel can fill in received frames; these
    will then turn up in an RX ring.
  - [rings.Completion] (umem): transfers ownership of umem transmitted
    chunks/frames from kernel space back to user space.
  - [rings.Rx]: produces received frames, from the fill ring.
  - [rings.Tx]: consumes frames to be sent; after sending, their chunks will
    return to the completion ring.

These rings contain [rings.Descriptor] elements, these can be either just simple
uint64 offsets of chunks inside an umem, or an offset together with packet
length and packet flags information ([unix.XDPDesc]).

The ring types offer receiver methods as necessary in user space, with the fill
and TX rings embedding the [rings.ProducerRing] type. The completion and RX rings
embedd the [rings.ConsumerRing] type correspondingly, providing only producer
methods instead.

Please note that rings only come in sizes that are powers of two.

[Linux kernel networking documentation on AF_XDP]: https://www.kernel.org/doc/html/v6.0/networking/af_xdp.html
[umem]: https://www.kernel.org/doc/html/next/networking/af_xdp.html#umem
*/
package xsk
