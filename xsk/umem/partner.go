// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package umem

import "sync"

// Partner with “interest” in a umem slice, obtained from [Map] or [unix.Mmap].
// When the last Partner ceases interest in the umem slice by calling
// [Partner.Cease], the umem slice is automatically [Unmap]'ed.
type Partner struct {
	Slice []byte // umem slice, cached, obtained by Map or unix.Mmap.

	mu   *sync.Mutex // shared mutex between Partners
	refs *uint       // shared “interest” (reference) counter between Partners

	ceased *bool // (unshared) double cease detection
}

// NewPartner returns the (first) Partner object interested in the passed umem
// slice. Partner objects can be passed around as values, but this does not
// change the “interested” Partner count; in order to add another Partner
// interested in the umem slice, use [Partner.New] (which actually can fail,
// returning a zero Partner).
//
// In order to cease “interest” in the umem, call [Partner.Cease].
func NewPartner(umem []byte) Partner {
	var refs uint = 1
	return Partner{
		Slice:  umem,
		mu:     new(sync.Mutex),
		refs:   &refs,
		ceased: new(bool),
	}
}

// New returns a new additional Partner also interested in the same umem slice,
// increasing the “interested” Partner count, if successful. In case the
// “interested” Partner count has already dropped to zero when trying to add
// another Partner, Another returns a zero Partner that can be detected by
// comparing [Partner.Slice] to nil.
func (p Partner) New() Partner {
	if p.mu == nil {
		panic("invalid Another on zero Partner, use NewPartner or Partner.Clone to obtain a Partner")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if *p.refs == 0 {
		return Partner{}
	}
	*p.refs++
	return Partner{
		Slice:  p.Slice,
		mu:     p.mu,
		refs:   p.refs,
		ceased: new(bool),
	}
}

// Cease being a umem partner, automatically [Unmap]'ing the umem slice true if
// this was the last umem partner interested. In this case, it returns true,
// otherwise false.
//
// Panics when called on either a zero Partner or when the Partner has already
// been stopped.
func (p Partner) Cease() (unmapped bool) {
	if p.mu == nil {
		panic("invalid Cease on zero Partner, use NewPartner or Partner.Clone to obtain a Partner")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if *p.ceased {
		panic("invalid Cease on already ceased Partner")
	}
	if *p.refs == 1 {
		// we're still under lock, so we can unmap before we decrement the
		// interest/reference counter.
		Unmap(p.Slice)
		unmapped = true
	}
	*p.refs--
	*p.ceased = true
	return
}
