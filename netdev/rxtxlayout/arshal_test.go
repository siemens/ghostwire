// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"encoding/json"

	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("JSON un/marshalling netdev information", func() {

	Context("Queue", func() {

		It("marshals with nil NAPI and IRQ", func() {
			q := &Queue{
				ID:   42,
				Type: NETDEV_QUEUE_TYPE_RX,
				NAPI: nil,
				IRQ:  nil,
			}
			Expect(Successful(json.Marshal(q))).To(MatchJSON(`
{
	"id": 42,
	"type": 0
}`))
		})

		It("marshals with non-nil NAPI and IRQ", func() {
			q := &Queue{
				ID:   42,
				Type: NETDEV_QUEUE_TYPE_RX,
				NAPI: &NAPI{ID: 123},
				IRQ:  &IRQ{ID: 666},
			}
			Expect(Successful(json.Marshal(q))).To(MatchJSON(`
{
	"id": 42,
	"type": 0,
	"napi-id": 123,
	"irq": 666
}`))
		})

	})

	Context("NAPI", func() {

		It("marshals with nil IRQ", func() {
			n := &NAPI{
				ID:      42,
				Index:   666,
				PID:     123,
				Kthread: &model.Process{},
				IRQ:     nil,
			}
			Expect(Successful(json.Marshal(n))).To(MatchJSON(`
{
	"id": 42,
	"index": 666,
	"pid": 123
}`))
		})

		It("marshals with non-nil IRQ", func() {
			n := &NAPI{
				ID:      42,
				Index:   666,
				PID:     123,
				Kthread: &model.Process{},
				IRQ: &IRQ{
					ID:  888,
					PID: 124,
				},
			}
			Expect(Successful(json.Marshal(n))).To(MatchJSON(`
{
	"id": 42,
	"index": 666,
	"pid": 123,
	"irq": 888
}`))
		})

	})

	Context("Netdev", func() {

		It("marshals a Netdev, then unmarshals", func() {
			nd := &Netdev{
				Name:         "foobar",
				OriginalName: "twoflower",
				Index:        42,
				Queues:       []*Queue{},
				NAPIs:        map[uint]*NAPI{},
				IRQs: map[uint]*IRQ{
					12: {
						ID:  12,
						PID: 1200,
					},
					47: {
						ID:  47,
						PID: 4700,
					},
				},
			}
			nd.NAPIs[8192] = &NAPI{
				ID:    8192,
				Index: 42,
				PID:   81920,
				IRQ:   nd.IRQs[12],
			}
			nd.NAPIs[8193] = &NAPI{
				ID:    8193,
				Index: 42,
				PID:   81930,
				IRQ:   nd.IRQs[47],
			}
			nd.Queues = append(nd.Queues, &Queue{
				ID:   0,
				Type: NETDEV_QUEUE_TYPE_RX,
				NAPI: nd.NAPIs[8192],
				IRQ:  nd.NAPIs[8192].IRQ,
			})
			nd.Queues = append(nd.Queues, &Queue{
				ID:   1,
				Type: NETDEV_QUEUE_TYPE_RX,
				NAPI: nd.NAPIs[8193],
				IRQ:  nd.NAPIs[8193].IRQ,
			})

			jtext := Successful(json.Marshal(nd))
			Expect(jtext).To(MatchJSON(`{
	"source": 0,
    "name": "foobar",
	"orig-name": "twoflower",
    "index": 42,
    "queues": [
        {
            "id": 0,
            "type": 0,
            "napi-id": 8192,
            "irq": 12
        },
        {
            "id": 1,
            "type": 0,
            "napi-id": 8193,
            "irq": 47
        }
    ],
    "napis": {
        "8192": {
            "id": 8192,
            "index": 42,
            "pid": 81920,
			"irq": 12
        },
        "8193": {
            "id": 8193,
            "index": 42,
            "pid": 81930,
			"irq": 47
        }
    },
    "irqs": {
        "12": {
			"source": 0,
            "id": 12,
            "pid": 1200
        },
        "47": {
			"source": 0,
            "id": 47,
            "pid": 4700
        }
    }
}`))

			var nd2 Netdev
			Expect(json.Unmarshal(jtext, &nd2)).To(Succeed())
			Expect(nd2).To(And(
				HaveField("Name", "foobar"),
				HaveField("Index", int(42))))
			Expect(nd2.Queues).To(HaveLen(2))
			Expect(nd2.NAPIs).To(HaveLen(2))
			Expect(nd2.IRQs).To(HaveLen(2))

			Expect(nd2.Queues[0].NAPI).To(BeIdenticalTo(nd2.NAPIs[8192]))
			Expect(nd2.Queues[0].IRQ).To(BeIdenticalTo(nd2.IRQs[12]))
			Expect(nd2.Queues[1].NAPI).To(BeIdenticalTo(nd2.NAPIs[8193]))
			Expect(nd2.Queues[1].IRQ).To(BeIdenticalTo(nd2.IRQs[47]))

			Expect(nd2.NAPIs[8192].IRQ).To(BeIdenticalTo(nd2.IRQs[12]))
			Expect(nd2.NAPIs[8193].IRQ).To(BeIdenticalTo(nd2.IRQs[47]))
		})

	})

})
