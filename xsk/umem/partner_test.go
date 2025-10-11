// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package umem

import (
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("umem Partners", func() {

	Context("zero Partners", func() {

		It("cannot be cloned", func() {
			Expect(func() {
				_ = Partner{}.New()
			}).To(Panic())
		})

		It("cannot cease", func() {
			Expect(func() {
				Partner{}.Cease()
			}).To(Panic())
		})

	})

	Context("new Partner", func() {

		var umem []byte
		var p Partner

		BeforeEach(func() {
			umem = Successful(unix.Mmap(
				-1,
				0, 4096,
				unix.PROT_READ|unix.PROT_WRITE,
				unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_POPULATE))
			p = NewPartner(umem)
			Expect(p.Slice).NotTo(BeNil())
			DeferCleanup(func() {
				if !*p.ceased {
					p.Cease()
				}
				_ = Unmap(umem)
			})
		})

		It("creates a Partner of an umem slice and then ceases Partnership", func() {
			Expect(&p.Slice[0]).To(BeIdenticalTo(&umem[0]))
			p.Slice[0] = 0x42
			Expect(p.Cease()).To(BeTrue())
			Expect(unix.Munmap(umem)).NotTo(Succeed())
		})

		It("adds another Partner", func() {
			p2 := p.New()
			Expect(p2.Slice).NotTo(BeNil())
			Expect(p2.refs).To(BeIdenticalTo(p.refs))
			Expect(*p2.refs).To(Equal(uint(2)))
			p.Cease()
			Expect(*p2.refs).To(Equal(uint(1)))
			p2.Cease()
			Expect(*p2.refs).To(BeZero())
		})

		It("cannot cease to be a Partner multiple times", func() {
			p.Cease()
			Expect(func() {
				p.Cease()
			}).To(Panic())
		})

		It("cannot add a new Partner when all Partners have ceased", func() {
			p.Cease()
			Expect(p.New().Slice).To(BeNil())
		})

	})

})
