// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"reflect"
	"runtime"
	"strings"

	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("XDP socket configuration options", func() {

	Context("mutually exclusive options", func() {

		It("Umem", func() {
			o := &Options{
				umemFd: -1,
			}
			Expect(WithSharedUmem(&Socket{})(o)).To(Succeed())
			Expect(WithUmemFd(0)(o)).NotTo(Succeed())

			o = &Options{}
			Expect(WithUmemFd(0)(o)).To(Succeed())
			defer unix.Close(o.umemFd)
			Expect(WithSharedUmem(&Socket{})(o)).NotTo(Succeed())
		})

		It("non/zero copy", func() {
			o := &Options{}
			Expect(ForceZeroCopy()(o)).To(Succeed())
			Expect(ForceCopy()(o)).NotTo(Succeed())

			o = &Options{}
			Expect(ForceCopy()(o)).To(Succeed())
			Expect(ForceZeroCopy()(o)).NotTo(Succeed())
		})

	})

	It("rejects invalid options", func() {
		opts := []Option{
			WithUmemFd(-1),
			WithSharedUmem(nil),
			WithChunkAmount(0),
			WithChunkSize(0),
			WithRxRingSize(0),
			WithRxRingSize(10),
			WithTxRingSize(0),
			WithTxRingSize(10),
			WithFillRingSize(0),
			WithFillRingSize(3),
			WithCompletionRingSize(0),
			WithCompletionRingSize(3),
		}
		for _, opt := range opts {
			fnpath := strings.Split(runtime.FuncForPC(reflect.ValueOf(opt).Pointer()).Name(), "/")
			fnels := strings.Split(fnpath[len(fnpath)-1], ".")
			optname := fnels[len(fnels)-2]
			o := &Options{}
			Expect(opt(o)).NotTo(Succeed(), "expected option %s to fail, but it succeeded", optname)
		}
	})

	It("accepts valid options", func() {
		o := &Options{}
		Expect(WithChunkAmount(16)(o)).To(Succeed())
		Expect(o.ChunkAmount).To(Equal(uint32(16)))
		Expect(WithChunkSize(4096)(o)).To(Succeed())
		Expect(o.ChunkSize).To(Equal(uint32(4096)))

		Expect(WithHeadroom(16)(o)).To(Succeed())
		Expect(o.Headroom).To(Equal(uint32(16)))

		Expect(WithRxRingSize(16)(o)).To(Succeed())
		Expect(o.RxRingSize).To(Equal(uint32(16)))
		Expect(WithoutRxRing()(o)).To(Succeed())
		Expect(o.RxRingSize).To(BeZero())

		Expect(WithTxRingSize(32)(o)).To(Succeed())
		Expect(o.TxRingSize).To(Equal(uint32(32)))
		Expect(WithoutTxRing()(o)).To(Succeed())
		Expect(o.TxRingSize).To(BeZero())

		Expect(WithFillRingSize(64)(o)).To(Succeed())
		Expect(o.FillRingSize).To(Equal(uint32(64)))
		Expect(WithCompletionRingSize(128)(o)).To(Succeed())
		Expect(o.CompletionRingSize).To(Equal(uint32(128)))

		o = &Options{}
		Expect(ForceCopy()(o)).To(Succeed())
		Expect(o.BindFlags).To(Equal(uint16(unix.XDP_COPY)))

		o = &Options{}
		Expect(ForceZeroCopy()(o)).To(Succeed())
		Expect(o.BindFlags).To(Equal(uint16(unix.XDP_ZEROCOPY)))
	})

	It("returns default options", func() {
		Expect(DefaultOptions()).NotTo(BeZero())
	})

})
