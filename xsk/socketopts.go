// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"errors"
	"fmt"

	"golang.org/x/sys/unix"
)

// Option configures an XDP socket option for creation.
type Option func(*Options) error

// Options stores XDP socket configuration information.
type Options struct {
	// Amount (number) of chunks in umem.
	ChunkAmount uint32
	// Size of each chunk inside the umem. Needs to be at least 2048 octets, but
	// cannot be larger than the page size (where the page size is
	// architecture-dependent).
	ChunkSize uint32
	// Head room within each chunk. Oh Max, those 80's!
	Headroom uint32

	// The socket with which to share its umem, if any. This is un-exported on
	// purpose, as we will nil this reference immediately after XSK creation.
	sharedUmemSocket *Socket

	// Number of fill ring descriptors.
	FillRingSize uint32
	// Number of completion ring descriptors.
	CompletionRingSize uint32

	// number of RX ring descriptors.
	RxRingSize uint32
	// number of TX ring descriptors.
	TxRingSize uint32

	// XSK bind flags
	BindFlags uint16

	// The umem fd being used (if any), either exclusively or shared with
	// another XDP socket. If not -1, it is a caller supplied fd that is valid
	// only for the duration of the Socket creation call.
	umemFd int
}

// XDP socket configuration defaults
const (
	DefaultChunkAmount        = 128
	DefaultChunkSize          = 2048
	DefaultHeadroom           = 0
	DefaultFillRingSize       = 64
	DefaultCompletionRingSize = 64
	DefaultRxRingSize         = 64
	DefaultTxRingSize         = 64
)

// DefaultOptions returns the default settings for XDP socket configuration
// options.
func DefaultOptions() Options {
	return Options{
		ChunkAmount: DefaultChunkAmount,
		ChunkSize:   DefaultChunkSize,
		Headroom:    DefaultHeadroom,

		umemFd: -1,

		FillRingSize:       DefaultFillRingSize,
		CompletionRingSize: DefaultCompletionRingSize,
		RxRingSize:         DefaultRxRingSize,
		TxRingSize:         DefaultTxRingSize,
	}
}

// WithUmemFd configures the XDP socket to use the umem referenced by the
// specified file descriptor (number). This fd is used only during the creation
// of a new [Socket] object, but not kept any longer, so the caller is free to
// close this fd as soon as [New] returns.
func WithUmemFd(fd int) Option {
	return func(o *Options) error {
		if fd < 0 {
			return fmt.Errorf("invalid umem fd %d", fd)
		}
		if o.sharedUmemSocket != nil {
			return errors.New("WithUmemFd and WithSharedUmem are mutually exclusive")
		}
		o.umemFd = fd
		return nil
	}
}

// WithSharedUmem configures the XDP socket to be created to use an already
// registered umem, as well as to not (re)configure fill and completion rings.
//
// Please note that [Socket.Fill] and [Socket.Completion] will return zero value
// ring objects in this case for this XDP socket.
func WithSharedUmem(xsk *Socket) Option {
	return func(o *Options) error {
		if xsk == nil {
			return errors.New("WithSharedUmem: nil XDP Socket object specified")
		}
		if o.umemFd >= 0 {
			return errors.New("WithSharedUmem and WithUmemFd are mutually exclusive")
		}
		o.FillRingSize = 0
		o.CompletionRingSize = 0
		o.ChunkAmount = xsk.opts.ChunkAmount
		o.ChunkSize = xsk.opts.ChunkSize
		o.Headroom = xsk.opts.Headroom
		o.BindFlags |= unix.XDP_SHARED_UMEM
		o.sharedUmemSocket = xsk
		return nil
	}
}

// WithChunkAmount configures the amount of umem chunks to be allocated.
func WithChunkAmount(n uint32) Option {
	return func(o *Options) error {
		if n == 0 {
			return errors.New("WithChunkAmount: chunk amount must not be zero")
		}
		o.ChunkAmount = n
		return nil
	}
}

// WithChunkSize configures the size of the umem chunks.
func WithChunkSize(n uint32) Option {
	return func(o *Options) error {
		if n == 0 {
			return errors.New("WithChunkSize: chunk size must not be zero")
		}
		o.ChunkSize = n
		return nil
	}
}

// WithHeadroom configures the amount of space between the beginning of a chunk
// and the beginning of frame data.
//
// Oh Max, those 80's!
func WithHeadroom(h uint32) Option {
	return func(o *Options) error {
		o.Headroom = h
		return nil
	}
}

// WithRxRingSize configures the size of an XDP socket's RX ring size. The size
// must be larger than zero, otherwise the configuration will raise an error.
func WithRxRingSize(size uint32) Option {
	return func(o *Options) error {
		if size < 2 {
			return errors.New("WithRxRingSize: ring size must be two or larger")
		}
		if size&(size-1) != 0 {
			return errors.New("WithRxRingSize: ring size must be a power of two")
		}
		o.RxRingSize = size
		return nil
	}
}

// WithoutRxRing configures an XDP socket without any RX ring. Please note that
// [Socket.Rx] will then return a zero value ring object.
func WithoutRxRing() Option {
	return func(o *Options) error {
		o.RxRingSize = 0
		return nil
	}
}

// WithTxRingSize configures the size of an XDP socket's TX ring size. The size
// must be larger than zero, otherwise the configuration will raise an error.
func WithTxRingSize(size uint32) Option {
	return func(o *Options) error {
		if size < 2 {
			return errors.New("WithTxRingSize: ring size must be two or larger")
		}
		if size&(size-1) != 0 {
			return errors.New("WithTxRingSize: ring size must be a power of two")
		}
		o.TxRingSize = size
		return nil
	}
}

// WithoutTxRing configures an XDP socket without any TX ring. Please note that
// [Socket.Tx] will then return a zero value ring object.
func WithoutTxRing() Option {
	return func(o *Options) error {
		o.TxRingSize = 0
		return nil
	}
}

// WithFillRingSize configures the size of an XDP socket's umem fill ring size.
// The size must be larger than zero, otherwise the configuration will raise an
// error.
func WithFillRingSize(size uint32) Option {
	return func(o *Options) error {
		if size < 2 {
			return errors.New("WithFillRingSize: umem fill ring size must be two or larger")
		}
		if size&(size-1) != 0 {
			return errors.New("WithFillRingSize: ring size must be a power of two")
		}
		o.FillRingSize = size
		return nil
	}
}

// WithCompletionRingSize configures the size of an XDP socket's umem fill ring size.
// The size must be larger than zero, otherwise the configuration will raise an
// error.
func WithCompletionRingSize(size uint32) Option {
	return func(o *Options) error {
		if size < 2 {
			return errors.New("WithCompletionRingSize: umem completion ring size must be two or larger")
		}
		if size&(size-1) != 0 {
			return errors.New("WithCompletionRingSize: ring size must be a power of two")
		}
		o.CompletionRingSize = size
		return nil
	}
}

// ForceZeroCopy configures the XDP socket to always use zero copy if supported,
// or otherwise fail XDP socket creation.
func ForceZeroCopy() Option {
	return func(o *Options) error {
		if o.BindFlags&unix.XDP_COPY != 0 {
			return errors.New("WithZeroCopy and WithCopy are mutually exclusive")
		}
		o.BindFlags |= unix.XDP_ZEROCOPY
		return nil
	}
}

// ForceCopy configures the XDP socket to always use non-zero copy if supported,
// or otherwise fail XDP socket creation.
func ForceCopy() Option {
	return func(o *Options) error {
		if o.BindFlags&unix.XDP_ZEROCOPY != 0 {
			return errors.New("WithCopy and WithZeroCopy are mutually exclusive")
		}
		o.BindFlags |= unix.XDP_COPY
		return nil
	}
}
