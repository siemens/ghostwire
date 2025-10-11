// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rings

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("descriptor pools", func() {

	It("returns all correct descriptors", func() {
		p := NewDescriptorPool(10, 2, 5)
		ds := []uint64{}
		for i := 2; i < 5; i++ {
			d, ok := p.Get()
			Expect(ok).To(BeTrue())
			ds = append(ds, d)
		}
		Expect(ds).To(ConsistOf(
			uint64(20),
			uint64(30),
			uint64(40),
		))
		_, ok := p.Get()
		Expect(ok).To(BeFalse())
	})

	It("waits for a descriptor to become available", NodeTimeout(2*time.Second), func(ctx context.Context) {
		p := NewDescriptorPool(10, 2, 3)
		Expect(p.Wait(ctx)).Error().NotTo(HaveOccurred())

		cuttyctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
		Expect(p.Wait(cuttyctx)).Error().To(HaveOccurred())

		p.Put(42)
		cuttyctx, cancel = context.WithTimeout(ctx, 200*time.Millisecond)
		defer cancel()
		Expect(p.Wait(cuttyctx)).To(Equal(uint64(42)))

	})

	It("puts back descriptors", func() {
		p := NewDescriptorPool(10, 2, 5)
		for {
			if _, ok := p.Get(); !ok {
				break
			}
		}
		p.Put(42)
		d, ok := p.Get()
		Expect(ok).To(BeTrue())
		Expect(d).To(Equal(uint64(42)))
	})

	It("panics when overflowing", func() {
		p := NewDescriptorPool(10, 0, 4)
		Expect(func() { p.Put(42) }).To(Panic())
	})

})
