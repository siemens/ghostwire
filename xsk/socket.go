// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"unsafe"

	"github.com/thediveo/caps/v2/errno"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/xsk/rings"
	"github.com/siemens/ghostwire/v2/xsk/umem"
)

// Socket represents an XDP socket – often just termed “XSK” due to tight
// memory restrictions, lack of time and an aim at determinism.
type Socket struct {
	mu sync.Mutex

	// XDP socket file descriptor.
	fd int
	// index of netdev this XSK is bound to.
	ifindex int
	// “hardware” queue ID this XSK is bound to.
	queueid int

	// local copy of the overall XDP socket configuration settings.
	opts Options

	// umem slice partner interest management.
	umem umem.Partner

	rx         rings.Rx
	tx         rings.Tx
	fill       rings.Fill
	completion rings.Completion
}

// New returns a new XDP socket that is attached to the specified network
// interface and “hardware” queue of that network interface. It registers the
// umem with the XDP socket and configures the various ring sizes. This umem can
// be either explicitly passed via configuration options [WithUmemFd] or
// [WithSharedUmem], or alternatively is created automatically based on the
// configured chunk amount and chunck size. Please see [Option] for the
// available configuration options for XDP sockets.
//
// Creating an XDP socket is a rather long-winded road and basically consists of
// the following steps:
//   - create the XDP socket itself, but that's just the beginning.
//   - automatically allocate umem if otherwise not explicitly configured.
//   - memory map the umem where necessary, and then map the umem in any case.
//   - configure the umem-related rings (sizes) for RX, TX, fill and completion.
//     that the caller or final consumer of the XDP socket needs to do. See the [Map] function.
//   - finally, finally, bind the XDP socket to the specified network interface and
//     queue of that network interface.
func New(ifindex int, queueid int, opts ...Option) (xsk *Socket, err error) {
	// Set the default option values and then process the options, bailing out
	// if there is an error setting options.
	options := DefaultOptions()
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, err
		}
	}
	xsk = &Socket{
		fd:      -1, // remember, remember, that 0 is a valid fd.
		ifindex: ifindex,
		queueid: queueid,
		opts:    options,
	}
	// Automatically close the already allocated XDP socket in case we leave
	// socket creation with any error. This avoids having to repeat the same
	// close-the-socket sequence in any error bail out over and over again. It
	// also ensures that we properly unmap memory regions that we already mapped
	// while setting up the socket.
	defer func() {
		if err != nil {
			if xsk == nil {
				return
			}
			if xsk.umem.Slice != nil {
				xsk.umem.Cease()
			}
			xsk.umem.Slice = nil // aid the GC
			if xsk.fd >= 0 {
				_ = unix.Close(xsk.fd)
			}
			xsk = nil
		}
	}()

	// First step: create the XDP socket ... that yet is more or less completely
	// useless at this stage and needs further configuration. This is just to
	// have the XSK fd as we need it later over and over again, as well as to
	// bail out early in case we're not allowed to create XDP sockets at all.
	var fd int
	fd, err = unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0)
	if err != nil {
		err = fmt.Errorf("cannot create XDP socket, reason: %w", err)
		return xsk, err // will nil in defer error handler
	}
	xsk.fd = fd

	// Next up: get the umem slice, taking care of shared umem management, et
	// cetera.
	if xsk.opts.sharedUmemSocket != nil {
		// Count us in with a shared interest in the umem that already has been
		// allocated.
		xsk.umem = xsk.opts.sharedUmemSocket.umem.New()
		if xsk.umem.Slice == nil {
			err = errors.New(
				"sharing XDP socket has already released its umem")
			return xsk, err
		}
		// nota bene: we don't share the umem fd... oh, well.
	} else if xsk.opts.umemFd < 0 {
		// If there hasn't been an already allocated umem specified (shared or not),
		// create an umem on the fly, but without any umem fd. This allows implicit
		// umem creation, but doesn't allow umem sharing; umem sharing always
		// requires explicit umem fd creation.
		umemsl, err := unix.Mmap(
			-1,
			0,
			int(xsk.opts.ChunkAmount)*int(xsk.opts.ChunkSize),
			unix.PROT_READ|unix.PROT_WRITE,
			unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_POPULATE)
		if err != nil {
			err = fmt.Errorf(
				"cannot implicitly create umem, reason: %w", err)
			return xsk, err
		}
		xsk.umem = umem.NewPartner(umemsl)
	} else /* xsk.opts.umemFd >= 0 */ {
		// With an explicitly configured umem fd (yet not shared with another
		// XSK), map the umem into memory.
		umemsl, err := umem.Map(xsk.opts.umemFd)
		if err != nil {
			err = fmt.Errorf(
				"cannot mmap configured umem fd, reason: %w", err)
			return xsk, err
		}
		xsk.umem = umem.NewPartner(umemsl)
	}

	// Do we need to register the umem and configure its fill and completion
	// rings? This can be done only if the umem isn't shared or for the first
	// XDP socket using a shared umem.
	if xsk.opts.FillRingSize > 0 && xsk.opts.CompletionRingSize > 0 {
		umemRegistration := unix.XDPUmemReg{
			Addr:     uint64(uintptr(unsafe.Pointer(&xsk.umem.Slice[0]))),
			Len:      uint64(len(xsk.umem.Slice)),
			Size:     xsk.opts.ChunkSize,
			Headroom: xsk.opts.Headroom,
		}
		if _, _, eno := unix.Syscall6(
			unix.SYS_SETSOCKOPT, uintptr(xsk.fd),
			unix.SOL_XDP, unix.XDP_UMEM_REG,
			uintptr(unsafe.Pointer(&umemRegistration)), unsafe.Sizeof(umemRegistration),
			0); eno != 0 {
			err = fmt.Errorf("cannot register umem, reason: %w", errno.Error(eno))
			return xsk, err
		}
		// Configure the size of the fill and completion rings.
		err = unix.SetsockoptInt(xsk.fd, unix.SOL_XDP, unix.XDP_UMEM_FILL_RING, int(xsk.opts.FillRingSize))
		if err != nil {
			err = fmt.Errorf(
				"cannot create fill ring of size %d, reason: %w", xsk.opts.FillRingSize, err)
			return xsk, err
		}
		err = unix.SetsockoptInt(xsk.fd, unix.SOL_XDP, unix.XDP_UMEM_COMPLETION_RING, int(xsk.opts.CompletionRingSize))
		if err != nil {
			err = fmt.Errorf(
				"cannot create completion ring of size %d, reason: %w", xsk.opts.CompletionRingSize, err)
			return xsk, err
		}
	}

	// Configure the size of the TX ring, unless not needed and specified as zero.
	if xsk.opts.TxRingSize > 0 {
		err = unix.SetsockoptInt(xsk.fd, unix.SOL_XDP, unix.XDP_TX_RING, int(xsk.opts.TxRingSize))
		if err != nil {
			err = fmt.Errorf(
				"cannot create TX ring of size %d, reason: %w", xsk.opts.TxRingSize, err)
			return xsk, err
		}
	}

	// Configure the size of the RX ring, unless not needed and specified as zero.
	if xsk.opts.RxRingSize > 0 {
		err = unix.SetsockoptInt(xsk.fd, unix.SOL_XDP, unix.XDP_RX_RING, int(xsk.opts.RxRingSize))
		if err != nil {
			err = fmt.Errorf(
				"cannot create RX ring of size %d, reason: %w", xsk.opts.RxRingSize, err)
			return xsk, err
		}
	}

	// Finally, finally, bind the XDP socket to its network interface and a
	// specific one of the network interface's queues.
	xskaddr := unix.SockaddrXDP{
		Flags:   xsk.opts.BindFlags,
		Ifindex: uint32(ifindex),
		QueueID: uint32(queueid),
	}
	if xsk.opts.sharedUmemSocket != nil {
		// The SharedUmemFD is not an umem fd -- and admittedly cannot be
		// because it might be non-mmapped umem -- but instead the XSK fd
		// what will share its registered umem with us. Ouch.
		xskaddr.SharedUmemFD = uint32(xsk.opts.sharedUmemSocket.fd)
	}
	err = unix.Bind(xsk.fd, &xskaddr)
	if err != nil {
		ifname := ""
		nif, err2 := net.InterfaceByIndex(ifindex)
		if err2 == nil {
			ifname = nif.Name
		}
		err = fmt.Errorf("cannot bind XDP socket to network interface %q (index %d) and queue %d, reason: %w",
			ifname, ifindex, queueid, err)
		return xsk, err
	}

	// Oh, one last thing: create and map the rings, as necessary.
	roffs, err := xsk.ringOffsets()
	if err != nil {
		return xsk, err
	}
	if xsk.opts.RxRingSize > 0 {
		xsk.rx, err = rings.NewRx(xsk.fd, roffs.Rx, xsk.opts.RxRingSize)
		if err != nil {
			return xsk, err
		}
	}
	if xsk.opts.TxRingSize > 0 {
		xsk.tx, err = rings.NewTx(xsk.fd, roffs.Tx, xsk.opts.TxRingSize)
		if err != nil {
			return xsk, err
		}
	}
	if xsk.opts.FillRingSize > 0 {
		xsk.fill, err = rings.NewFill(xsk.fd, roffs.Fr, xsk.opts.FillRingSize)
		if err != nil {
			return xsk, err
		}
	}
	if xsk.opts.CompletionRingSize > 0 {
		xsk.completion, err = rings.NewCompletion(xsk.fd, roffs.Cr, xsk.opts.CompletionRingSize)
		if err != nil {
			return xsk, err
		}
	}

	return xsk, nil
}

// Close the XDP socket. It is no error to close an already closed XDP socket.
// This also closes the RX and TX rings. It only closes the umem-related fill
// and completion rings if this is the last XDP socket closing sharing the same
// umem.
func (xsk *Socket) Close() error {
	xsk.mu.Lock()
	defer xsk.mu.Unlock()

	xsk.rx.Close()
	xsk.tx.Close()

	if xsk.umem.Slice != nil {
		if xsk.umem.Cease() {
			// We were the last partner that lost interest, so the umem has
			// already been unmapped and we also now need to unmap the
			// umem-related rings.
			xsk.fill.Close()
			xsk.completion.Close()
		}
		xsk.umem.Slice = nil // aid the GC
	}
	if xsk.fd < 0 {
		return nil
	}
	err := unix.Close(xsk.fd)
	xsk.fd = -1
	return err
}

// Rx returns the receiving packets ring object, or a zero ring object in case
// this XDP socket has no RX ring allocated. Please note that ring objects can
// be safely passed by value.
func (xsk *Socket) Rx() rings.Rx { return xsk.rx }

// Tx returns the sending packets ring object, or a zero ring object in case
// this XDP socket has no TX ring allocated. Please note that ring objects can
// be safely passed by value.
//
// Without the need for busy polling, XDP socket users can determine whether the
// Tx ring is willing to accept more chunks by using [unix.Poll] and looking for
// the [unix.EPOLLOUT] event. Please note that newer Linux kernels report
// [unix.EPOLLOUT] only as long as the Tx ring is less than half full
// (please refer to [xsk_poll], [xsk_tx_writeable] for implementation details).
//
// [xsk_poll]: https://elixir.bootlin.com/linux/v6.9.4/source/net/xdp/xsk.c#L1016
// [xsk_tx_writeable]: https://elixir.bootlin.com/linux/v6.9.4/source/net/xdp/xsk.c#L296
func (xsk *Socket) Tx() rings.Tx { return xsk.tx }

// Fill returns the fill ring object for providing umem chunks to be filled with
// received packets. If multiple XDP sockets share the same umem, only the first
// XDP socket returns a non-zero fill ring object.
//
// Please note that ring objects can be safely passed by value.
func (xsk *Socket) Fill() rings.Fill { return xsk.fill }

// Completion returns the completion ring object for picking up umem chunks
// after they have been transmitted over the network interface. If multiple XDP
// sockets share the same umem, only the first XDP socket returns a non-zero
// completion ring object.
//
// Please note that ring objects can be safely passed by value.
//
// Contrary to some documentation about XDP, it is actually not possible to
// [unix.Poll] for sent packet chunks to appear in the completion ring, the
// kernel [xsk_poll] code does implement such behavior.
//
// [xsk_poll]:
// https://elixir.bootlin.com/linux/v6.9.4/source/net/xdp/xsk.c#L1016
func (xsk *Socket) Completion() rings.Completion { return xsk.completion }

// Umem returns the umem slice holding chunks (“packet buffers”). If multiple
// XDP sockets share the same umem, the same umem slice is returned from the
// sharing XDP sockets.
func (xsk *Socket) Umem() []byte { return xsk.umem.Slice }

// Fd returns the file descriptor (number) of this XDP socket. It returns -1
// after [Socket.Close] has been called.
func (xsk *Socket) Fd() int { return xsk.fd }

// Options returns the configuration options used when creating this XDP socket.
func (xsk *Socket) Options() Options {
	return xsk.opts
}

// ringOffset returns the ring offsets into the ring memory the kernel allocated
// for this XDP socket. We handle both the old v1 syscall data structure, as
// well as the newer "v2" structure transparently, returning always the current
// structure, with the Flags field zero for v1.
func (xsk *Socket) ringOffsets() (unix.XDPMmapOffsets, error) {
	// To complicate matters, kernels up to and including 5.3 use an older "v1"
	// version that isn't present in the Go unix package. We simply supply
	// enough space for the most recent version and then check the outcome of
	// the XDP_MMAP_OFFSETS query to see if it's really the recent one, or the
	// older v1 version instead.
	var offsets unix.XDPMmapOffsets
	var offsetsV1 *XDPMmapOffsetsV1
	offsetsLen := int(unsafe.Sizeof(offsets))
	offsetsRaw := make([]byte, offsetsLen)
	_, _, e1 := unix.Syscall6(
		unix.SYS_GETSOCKOPT,
		uintptr(xsk.fd),
		unix.SOL_XDP,
		unix.XDP_MMAP_OFFSETS,
		uintptr(unsafe.Pointer(&offsetsRaw[0])),
		uintptr(unsafe.Pointer(&offsetsLen)),
		0)
	if e1 != 0 {
		return unix.XDPMmapOffsets{}, fmt.Errorf(
			"cannot retrieve XDP socket ring mmap offsets, reason: %d", e1)
	}
	// Now postprocess the ring offsets as necessary, translating v1 ring
	// offsets into the "v2"(?) offset data structures. V1 lacks the flags that
	// have been included as of v2.
	switch offsetsLen {
	case int(unsafe.Sizeof(offsets)):
		offsets = *(*unix.XDPMmapOffsets)(unsafe.Pointer(&offsetsRaw[0]))
	case int(unsafe.Sizeof(offsetsV1)):
		offsetsV1 = (*XDPMmapOffsetsV1)(unsafe.Pointer(&offsetsRaw[0]))
		offsets.Rx = unix.XDPRingOffset{
			Producer: offsetsV1.Rx.Producer,
			Consumer: offsetsV1.Rx.Consumer,
			Desc:     offsetsV1.Rx.Desc,
		}
		offsets.Tx = unix.XDPRingOffset{
			Producer: offsetsV1.Tx.Producer,
			Consumer: offsetsV1.Tx.Consumer,
			Desc:     offsetsV1.Tx.Desc,
		}
		offsets.Fr = unix.XDPRingOffset{
			Producer: offsetsV1.Fr.Producer,
			Consumer: offsetsV1.Fr.Consumer,
			Desc:     offsetsV1.Fr.Desc,
		}
		offsets.Cr = unix.XDPRingOffset{
			Producer: offsetsV1.Cr.Producer,
			Consumer: offsetsV1.Cr.Consumer,
			Desc:     offsetsV1.Cr.Desc,
		}
	default:
		return unix.XDPMmapOffsets{}, fmt.Errorf("cannot retrieve XDP socket ring mmap offsets, reason: unknown response size")
	}
	return offsets, nil
}

// XDPMmapOffsetsV1 is like [unix.XDPMmapOffsets], but without the flag fields
// for the individual rings.
type XDPMmapOffsetsV1 struct {
	Rx XDPRingOffsetV1
	Tx XDPRingOffsetV1
	Fr XDPRingOffsetV1
	Cr XDPRingOffsetV1
}

// XDPRingOffsetV1 is like [unix.XDPRingOffset], but without the flag field that
// was only later added.
type XDPRingOffsetV1 struct {
	Producer uint64
	Consumer uint64
	Desc     uint64
}
