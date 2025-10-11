// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rings

import "context"

// DescriptorPooler allows retrieval and return of umem chunk descriptor
// (addresses) when chunks are temporarily not in use in any of the RX, TX, fill
// and completion rings.
type DescriptorPooler interface {
	// Get a descriptor (address), or false if none are currently available.
	// This will never block.
	Get() (addr uint64, ok bool)
	// Wait for a descriptor (address) to become available.
	Wait(ctx context.Context) (addr uint64, err error)
	// Put back a descriptor (address) into the pool; this will panic in case
	// you are putting back more descriptors than the pool actually can hold.
	Put(addr uint64)
}

// Descriptor pool implements a concurrency-safe descriptor pool. DescriptorPool
// is safe to pass around by value. It not optimized in any way.
type DescriptorPool struct {
	pool chan uint64 // creativity alert: misusing a channel as a pool
}

var _ DescriptorPooler = (*DescriptorPool)(nil)

// NewDescriptorPool returns a DescriptorPool with descriptor (addresses)
// initialized in the range low*chunksize..(high-1)*chunksize. The concrete
// DescriptorPool type is returned in favor of the DescriptorPooler interface in
// order to allow the Go compiler to optimize pool function calls.
func NewDescriptorPool(chunksize uint32, low uint32, high uint32) DescriptorPool {
	p := DescriptorPool{
		pool: make(chan uint64, high-low),
	}
	for chunkno := low; chunkno < high; chunkno++ {
		p.pool <- uint64(chunkno) * uint64(chunksize)
	}
	return p
}

// Get a descriptor (address), or false if none are currently available.
// This will never block.
func (p DescriptorPool) Get() (addr uint64, ok bool) {
	select {
	case addr := <-p.pool:
		return addr, true
	default:
	}
	return 0, false
}

// Wait for a descriptor (address) to become available.
func (p DescriptorPool) Wait(ctx context.Context) (addr uint64, err error) {
	select {
	case d := <-p.pool:
		return d, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// Put back a descriptor (address) into the pool; this will panic in case
// you are putting back more descriptors than the pool actually can hold.
func (p DescriptorPool) Put(addr uint64) {
	select {
	case p.pool <- addr:
		return
	default:
		panic("DescriptorPool overflow")
	}
}
