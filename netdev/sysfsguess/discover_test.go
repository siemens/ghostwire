// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
	"github.com/siemens/ghostwire/v2/passedthrough"
	"github.com/thediveo/lxkns/model"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("discover netdev configurations", func() {

	It("discovers twoflower's layout", func() {
		fakenlndevs := []netlink.Link{
			&netlink.Device{
				LinkAttrs: netlink.LinkAttrs{
					Name:  "cohen",
					Index: 42,
					AltNames: []string{
						"rumpelpumpel",
						passedthrough.AltNameOriginalIfnamePrefix + "twoflower",
					},
				},
			},
		}
		sysfspath := "./_test/cohen"
		fakeprocs := model.ProcessTable{
			2: { /* Children */ },
			100: {
				PID: 100,
				ProTaskCommon: model.ProTaskCommon{
					Name: "irq/42-twoflower",
				},
			},
			101: {
				PID: 101,
				ProTaskCommon: model.ProTaskCommon{
					Name: "irq/666-twoflower",
				},
			},
			200: {
				PID: 200,
				ProTaskCommon: model.ProTaskCommon{
					Name: "napi/twoflower-8193",
				},
			},
			201: {
				PID: 201,
				ProTaskCommon: model.ProTaskCommon{
					Name: "napi/twoflower-8194",
				},
			},
		}
		fakeprocs[2].Children = []*model.Process{
			fakeprocs[100], fakeprocs[101], fakeprocs[200], fakeprocs[201],
		}
		fakeirqkthreads := rxtxlayout.NewIRQKthreadsMap(fakeprocs)
		Expect(fakeirqkthreads).To(HaveLen(2))
		fakenapikthreads := rxtxlayout.NewNAPIKthreadsMap(fakeprocs)
		Expect(fakenapikthreads["twoflower"]).To(HaveLen(2))
		ndevs := discoverLayouts(fakenlndevs, sysfspath, fakeirqkthreads, fakenapikthreads)
		Expect(ndevs).To(HaveLen(1))
		ndev := ndevs[0]
		Expect(ndev).To(And(
			HaveField("Name", "cohen"),
			HaveField("Index", 42),
			HaveField("OriginalName", "twoflower")))
		Expect(ndev.Queues).To(And(
			HaveLen(4),
			HaveEach(And(
				HaveField("IRQ", Not(BeNil())),
				HaveField("NAPI", BeNil()), // because sysfs doesn't tell us this relation :(
			))))
		Expect(ndev.NAPIs).To(And(
			HaveLen(2),
			HaveEach(And(
				HaveField("ID", Not(BeZero())),
				HaveField("Index", 42),
				HaveField("PID", Not(BeZero())),
				HaveField("Kthread", Not(BeNil())),
				HaveField("IRQ", BeNil()), // because sysfs doesn't tell us this relation :(
			))))
		Expect(ndev.IRQs).To(And(
			HaveLen(2),
			HaveEach(And(
				HaveField("ID", Not(BeZero())),
				HaveField("PID", Not(BeZero())),
				HaveField("Kthread", Not(BeNil())),
			))))
	})

})
