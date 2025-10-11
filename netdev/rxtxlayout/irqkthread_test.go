// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IRQ kthreads", func() {

	var kthreadprocs = model.ProcessTable{
		2: &model.Process{},
		8: &model.Process{
			PID: 8,
			ProTaskCommon: model.ProTaskCommon{
				Name: "napi/twoflower-8193",
			},
		},
		13: &model.Process{
			PID: 13,
			ProTaskCommon: model.ProTaskCommon{
				Name: "irq/42-twoflower",
			},
		},
		14: &model.Process{
			PID: 14,
			ProTaskCommon: model.ProTaskCommon{
				Name: "irq/666-twoflower",
			},
		},
	}

	kthreadprocs[2].Children = []*model.Process{
		kthreadprocs[8], kthreadprocs[13], kthreadprocs[14],
	}

	It("returns empty maps when there's simply nothing", func() {
		Expect(NewIRQKthreadsMap(nil)).To(BeEmpty())
		Expect(NewIRQKthreadsMap(model.ProcessTable{
			42: &model.Process{},
		})).To(BeEmpty())
	})

	It("maps IRQ kthreads", func() {
		Expect(NewIRQKthreadsMap(kthreadprocs)).To(And(
			HaveKeyWithValue(uint(42), HaveField("PID", model.PIDType(13))),
			HaveKeyWithValue(uint(666), HaveField("PID", model.PIDType(14)))))
	})

	It("resolves kthread Process links", func() {
		irqmap := NewIRQKthreadsMap(kthreadprocs)
		ndevs := []*Netdev{
			{
				IRQs: map[uint]*IRQ{
					42:  {ID: 42},
					666: {ID: 666},
				},
			},
		}
		FillInIRQKthreads(irqmap, ndevs)
		Expect(ndevs[0].IRQs).To(ConsistOf(
			And(HaveField("ID", uint(42)),
				HaveField("Kthread", BeIdenticalTo(kthreadprocs[13]))),
			And(HaveField("ID", uint(666)),
				HaveField("Kthread", BeIdenticalTo(kthreadprocs[14])))))
	})

	It("skips kthread fill-in", func() {
		ndevs := []*Netdev{
			{
				IRQs: map[uint]*IRQ{
					123: {ID: 123},
				},
			},
		}
		FillInIRQKthreads(nil, ndevs)
		Expect(ndevs[0].IRQs[123]).To(And(
			HaveField("PID", BeZero()),
			HaveField("Kthread", BeNil())))
	})

	It("accepts correctly formed IRQ kthread names", func() {
		irq, kt, ok := kvOfIRQKthread(&model.Process{
			ProTaskCommon: model.ProTaskCommon{Name: irqKthreadnamePrefix + "42-abc"},
		})
		Expect(irq).To(Equal(uint(42)))
		Expect(kt).NotTo(BeNil())
		Expect(ok).To(BeTrue())
	})

	It("skips malformed IRQ kthread names", func() {
		irq, kt, ok := kvOfIRQKthread(&model.Process{
			ProTaskCommon: model.ProTaskCommon{Name: irqKthreadnamePrefix},
		})
		Expect(irq).To(BeZero())
		Expect(kt).To(BeNil())
		Expect(ok).To(BeFalse())

		irq, kt, ok = kvOfIRQKthread(&model.Process{
			ProTaskCommon: model.ProTaskCommon{Name: irqKthreadnamePrefix + "42"},
		})
		Expect(irq).To(BeZero())
		Expect(kt).To(BeNil())
		Expect(ok).To(BeFalse())

		irq, kt, ok = kvOfIRQKthread(&model.Process{
			ProTaskCommon: model.ProTaskCommon{Name: irqKthreadnamePrefix + "abc-"},
		})
		Expect(irq).To(BeZero())
		Expect(kt).To(BeNil())
		Expect(ok).To(BeFalse())
	})

})
