// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rings

import (
	"fmt"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

// Descriptor specifies the offset of a chunk inside a umem. In case of RX and
// TX ring descriptors, it additional describes the length of a packet inside
// the chunk, as well as flags, such as [unix.XDP_PKT_CONTD] for [multi-buffer
// support].
//
// [multi-buffer support]: https://www.kernel.org/doc/html/next/networking/af_xdp.html#multi-buffer-support
type Descriptor interface {
	uint64 | unix.XDPDesc
}

// ring represents a descriptor ring of a “single producer and single consumer”
// design, where the descriptors “point” to chunks in the associated umem in
// form of offsets. Rings are used for fill/completion and RX/TX.
//
// The size of a ring must always be a power of two.
//
// Please note that this implementation is not optimized in the way [libxdp]'s is
// (the latter using clever caching and memory barriers).
//
// Ring objects can be copied after their creation and configuration, because
// the actual ring memory as well as the producer and consumer indices are in
// memory provided by the kernel and a Go ring object points to these instead of
// storing them itself.
//
// As kind of some consolation this implementation uses Go Generics. Take that,
// boilerplate!
//
// As for the ring inner workings, please head over to Juho Snellman's highly
// useful blog post “[I've been writing ring buffers wrong all these years]”. We
// have to strictly follow the “array with two unmasked indices” as this is what
// the kernel and libxdp do and what eventually the kernel-user space API
// contract is.
//
// [libxdp]: https://github.com/xdp-project/xdp-tools/tree/master/lib/libxdp
// [I've been writing ring buffers wrong all these years]: https://www.snellman.net/blog/archive/2016-12-13-ring-buffers/
type ring[D Descriptor] struct {
	// Index of the descriptor that is the current head of the queue (ring). To
	// be more precise, this is the “unmasked” index that only wraps around the
	// 32bit boundary, but never wraps at the end of the ring (size).
	//
	// The storage for the producer index is always provided by the kernel, so
	// we're pointing to that storage. This also takes care of aligning the
	// producer index on a cache line as to avoid cash line trashing, depending
	// on the particular CPU architecture.
	producer *uint32
	// Index of the descriptor that is the current tail of the queue (ring).
	// Again, this is the “unmasked” index that only wraps around the 32bit
	// boundary, but never wraps at the end of the ring (size).
	//
	// The storage for the consumer index is always provided by the kernel, so
	// we're pointing to that storage. This also takes care of aligning the
	// producer index on a cache line as to avoid cash line trashing, depending
	// on the particular CPU architecture.
	consumer *uint32
	// The descriptors for this ring.
	descriptors []D

	// Ring-specific flags, such as unix.XDP_RING_NEED_WAKEUP
	flags *uint32

	// The size of this ring, which must always be a power of two. Didn't
	// Tollkühn know this?!
	size uint32
	// Mask to map the free running producer and consumer indices into the
	// actual range of descriptor elements.
	mask uint32

	// Base address where the ring is somewhere located inside; needed for later
	// unmapping the ring memory region during clean-up.
	ringmem []byte
}

// producerRing implements adding (“producing”) descriptors to it, but not
// removing any (as the latter is the job of the kernel).
type producerRing[D Descriptor] struct{ ring[D] }

// consumerRing implements removing (“consuming”) descriptors from it, but not
// adding any (as the latter is the job of the kernel).
type consumerRing[D Descriptor] struct{ ring[D] }

// Rx implements an XDP socket ring for received packets. This is a
// [consumerRing] as it gets filled from kernel space and then the packets must
// be consumed in user space.
type Rx struct{ consumerRing[unix.XDPDesc] }

// Tx implements an XDP socket ring for packet transmission. This is a
// [producerRing] as it gets filled from user space with descriptors for packets
// to be sent.
type Tx struct{ producerRing[unix.XDPDesc] }

// Fill implements a umem ring for chunks to be filled with received packets.
// This is a [producerRing] as it gets filled from user space with descriptors
// for packets to be received.
type Fill struct{ producerRing[uint64] }

// Completion implements a umem ring for packets that have been sent, so that
// their chunks can now be reused for something else. This is a [consumerRing]
// as it gets filled from kernel space and then the packets must be consumed in
// user space.
type Completion struct{ consumerRing[uint64] }

// NewFill returns a new Fill ring object for the umem attached to the XDP
// socket referenced by xskfd. Otherwise, it returns an error. Make sure to call
// [ring.Close] to properly release the mmap'ed ring memory.
func NewFill(xskfd int, offsets unix.XDPRingOffset, size uint32) (Fill, error) {
	r := Fill{}
	if err := r.setup(xskfd,
		unix.XDP_UMEM_PGOFF_FILL_RING, offsets, size); err != nil {
		return Fill{}, err
	}
	return r, nil
}

// NewCompletion returns a new Completion ring object for the umem attached to
// the XDP socket referenced by xskfd. Otherwise, it returns an error. Make sure
// to call [ring.Close] to properly release the mmap'ed ring memory.
func NewCompletion(xskfd int, offsets unix.XDPRingOffset, size uint32) (Completion, error) {
	r := Completion{}
	if err := r.setup(xskfd,
		unix.XDP_UMEM_PGOFF_COMPLETION_RING, offsets, size); err != nil {
		return Completion{}, err
	}
	return r, nil
}

// NewRx returns a new RX ring object for the XDP socket referenced by xskfd.
// Otherwise, it returns an error. Make sure to call [ring.Close] to properly
// release the mmap'ed ring memory.
func NewRx(xskfd int, offsets unix.XDPRingOffset, size uint32) (Rx, error) {
	r := Rx{}
	if err := r.setup(xskfd,
		unix.XDP_PGOFF_RX_RING, offsets, size); err != nil {
		return Rx{}, err
	}
	return r, nil
}

// NewTx returns a new TX ring object for the XDP socket referenced by xskfd.
// Otherwise, it returns an error. Make sure to call [ring.Close] to properly
// release the mmap'ed ring memory.
func NewTx(xskfd int, offsets unix.XDPRingOffset, size uint32) (Tx, error) {
	r := Tx{}
	if err := r.setup(xskfd,
		unix.XDP_PGOFF_TX_RING, offsets, size); err != nil {
		return Tx{}, err
	}
	return r, nil
}

// String returns a textual representation of a ring, giving the current usage
// gauge.
func (r ring[D]) String() string {
	return fmt.Sprintf("%T in use: %d", r, r.Used())
}

// Flags return this Ring's flags as of this very moment.
//
// For instance, unix.
func (r ring[D]) Flags() uint32 {
	return atomic.LoadUint32(r.flags)
}

// Used returns the number of chunks that could be dequeued right at this moment
// by calling [ring.Next]. There might also be more just right now.
func (r ring[D]) Used() int {
	// Oh, the magic of unsigned integer subtraction with wrap-arounds.
	return int(atomic.LoadUint32(r.producer) - atomic.LoadUint32(r.consumer))
}

// Free returns the number of currently unused descriptors in this ring.
func (r *ring[D]) Free() int {
	return int(r.size) - r.Used()
}

// Full becomes true when the ring is filled to its brim, so that no new
// descriptors can be added.
func (r ring[D]) Full() bool {
	return r.Used() == int(r.size)
}

// Empty is true if the ring doesn't contain any descriptors.
func (r ring[D]) Empty() bool {
	return atomic.LoadUint32(r.producer) == atomic.LoadUint32(r.consumer)
}

// setup maps the specified ring (unix.XDP_UMEM_PGOFF_FILL_RING, ...) into this
// program's virtual memory space and initializes this ring to point to the ring
// memory (descriptors and head/tail indices).
//
//   - xskfd: an open file descriptor (number) referencing an XDP socket.
//   - ring: identifies the ring to set this object up for, this must be one of
//     unix.XDP_UMEM_PGOFF_FILL_RING, et cetera.
//   - offsets: of the ring elements as told us by the kernel via an getsockopt
//     with SOL_XDP and XDP_MMAP_OFFSETS.
//   - size: of the ring, must be a power of two.
func (r *ring[D]) setup(xskfd int, ring int64, offsets unix.XDPRingOffset, size uint32) (err error) {
	var zeroDescriptor D
	r.size = size
	r.mask = size - 1
	// First step: map the ring somewhere into our process's virtual memory
	// space. This ring memory consists of not only the ring entries themselves,
	// but also the producer and consumer indices, and some more.
	r.ringmem, err = unix.Mmap(
		xskfd,
		ring,
		int(offsets.Desc+uint64(size)*uint64(unsafe.Sizeof(zeroDescriptor))),
		unix.PROT_READ|unix.PROT_WRITE,
		unix.MAP_SHARED|unix.MAP_POPULATE,
	)
	if err != nil {
		return fmt.Errorf("cannot map ring at offset 0x%x, reason: %w", ring, err)
	}
	// Second step: locate where the head/tail (producer/consumer) indices of
	// this ring are to be found in shared memory.
	r.producer = (*uint32)(unsafe.Add(unsafe.Pointer(&r.ringmem[0]), offsets.Producer))
	r.consumer = (*uint32)(unsafe.Add(unsafe.Pointer(&r.ringmem[0]), offsets.Consumer))
	// Two and a half: locate the ring flags, which are also highly dynamic.
	r.flags = (*uint32)(unsafe.Add(unsafe.Pointer(&r.ringmem[0]), offsets.Flags))
	// Third step: let the ring memory appear to the Go runtime (gc) as a
	// correctly typed and sized slice.
	//
	// Now, reflect.SliceHeader is deprecated (good riddance) and we welcome
	// unsafe.Slice instead. The documentation talks about unsafe.Slice being
	// equivalent to
	//
	//   (*[len]ArbitraryType)(unsafe.Pointer(ptr))[:]
	//
	// which is a bit of a stretch considering that [len]ArbitraryType must be a
	// compile-time array type. However, we don't know the array length at
	// compile-time. Yet unsafe.Slice expects a *ArbitraryType as its first
	// parameter. The only way to successfully pass the compiler is to use
	// unsafe.Pointer acrobatics, because:
	//
	//   1. A pointer value of any type can be converted to an unsafe.Pointer.
	//   2. An unsafe.Pointer can be converted to a pointer value of any type.
	r.descriptors = unsafe.Slice((*D)(unsafe.Pointer(&r.ringmem[offsets.Desc])), size)

	// Done and dusted.
	return nil
}

// Close releases the virtual memory space once mapped into our user space when
// setting up this ring.
func (r *ring[D]) Close() {
	// We ignore any errors silently.
	_ = unix.Munmap(r.ringmem)
}

// Add adds a descriptor to a (producer) ring, returning true, if the ring had
// room for the descriptor. Otherwise, it returns false when the ring is full.
func (p producerRing[D]) Add(d D) bool {
	if p.Full() {
		return false
	}
	idx := atomic.LoadUint32(p.producer)
	p.descriptors[idx&p.mask] = d
	idx++
	atomic.StoreUint32(p.producer, idx)
	return true
}

// NeedsWakeup indicates whether this Tx needs an explicit poke after adding one
// or more descriptors. Please note that wakeup might be signalled without the
// [unix.XDP_USE_NEED_WAKEUP] option having been specified on the XDP socket at
// bind time.
//
// See also:
//
//   - Linux kernel documentation on [XDP_RING_NEED_WAKEUP] (but notice
//     how the documentation only mentions the XDP_USE_NEED_WAKEUP bind(!)
//     flag, but not XDP_RING_NEED_WAKEUP ring flag).
//   - [Tx needs to be explicitly woken up the first time] in the kernel
//     sources.
//
// [XDP_RING_NEED_WAKEUP]: https://www.kernel.org/networking/af_xdp.html#xdp-use-need-wakeup-bind-flag
// [Tx needs to be explicitly woken up the first time]: https://elixir.bootlin.com/linux/v6.9.4/source/net/xdp/xsk.c#L1376
func (t Tx) NeedsWakeup() bool {
	return t.Flags()&unix.XDP_RING_NEED_WAKEUP != 0
}

// Next removes and returns the next descriptor from the ring, if any, together
// with a true value. If the ring is empty, then a zero descriptor together with
// a false value is returned instead.
func (p consumerRing[D]) Next() (d D, ok bool) {
	if p.Empty() {
		return d, false
	}
	idx := atomic.LoadUint32(p.consumer)
	d = p.descriptors[idx&p.mask]
	idx++
	atomic.StoreUint32(p.consumer, idx)
	return d, true
}
