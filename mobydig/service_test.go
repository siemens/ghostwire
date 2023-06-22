// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"github.com/jinzhu/copier"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var s0 = ServiceOnNetworks{
	Name: "",
	Servants: Containers{
		{Name: "c1"},
		{Name: "c2"},
	},
	NetworkLabels: []string{"net-a", "net-b"},
}

var s0a = ServiceOnNetworks{
	Name: "",
	Servants: Containers{
		{Name: "c-a"},
	},
	NetworkLabels: []string{"net-a"},
}

var s0b = ServiceOnNetworks{
	Name: "",
	Servants: Containers{
		{Name: "c-b"},
	},
	NetworkLabels: []string{"net-b"},
}

var s1 = ServiceOnNetworks{
	Name: "service",
	Servants: Containers{
		{Name: "c1"},
		{Name: "c2"},
	},
	NetworkLabels: []string{"net-a", "net-b"},
}

var sx = ServiceOnNetworks{
	Name: "service-x",
	Servants: Containers{
		{Name: "x1"},
	},
	NetworkLabels: []string{"net-a"},
}

var sx2 = ServiceOnNetworks{
	Name: "service-x",
	Servants: Containers{
		{Name: "x1"},
	},
	NetworkLabels: []string{"net-b"},
}

var _ = Describe("Services", func() {

	It("returns all DNS name combinations of a service", func() {
		Expect(s0.DnsNames()).To(ConsistOf(
			"c1", "c1.net-a", "c1.net-b",
			"c2", "c2.net-a", "c2.net-b",
		))

		Expect(s1.DnsNames()).To(ConsistOf(
			"service", "service.net-a", "service.net-b",
			"c1", "c1.net-a", "c1.net-b",
			"c2", "c2.net-a", "c2.net-b",
		))
	})

	It("filters service network labels", func() {
		s := make(Services, 1)
		copier.CopyWithOption(&s[0], s1, copier.Option{DeepCopy: true, IgnoreEmpty: false})
		s.FilterNetworkLabels([]string{"net-x"})
		Expect(s[0].NetworkLabels).To(BeEmpty())

		copier.CopyWithOption(&s[0], s1, copier.Option{DeepCopy: true, IgnoreEmpty: false})
		s.FilterNetworkLabels([]string{"net-b"})
		Expect(s[0].NetworkLabels).To(ConsistOf("net-b"))

		copier.CopyWithOption(&s[0], s1, copier.Option{DeepCopy: true, IgnoreEmpty: false})
		s.FilterNetworkLabels([]string{"net-b", "net-a", "net-x"})
		Expect(s[0].NetworkLabels).To(ConsistOf("net-a", "net-b"))
	})

	It("deduplicates services", func() {
		Expect(Services{s1, sx}.Deduplicate()).To(ConsistOf(
			HaveField("Name", "service"),
			HaveField("Name", "service-x"),
		))

		dd := Services{s1, s1}.Deduplicate()
		Expect(dd).To(HaveLen(1))
		Expect(dd[0]).To(And(
			HaveField("Name", "service"),
			HaveField("Servants", HaveLen(2)),
			HaveField("NetworkLabels", HaveLen(2)),
		))

		dd = Services{sx, sx2}.Deduplicate()
		Expect(dd).To(HaveLen(1))
		Expect(dd[0]).To(And(
			HaveField("Name", "service-x"),
			HaveField("Servants", HaveLen(1)),
			HaveField("NetworkLabels", ConsistOf("net-a", "net-b")),
		))

		dd = Services{s0a, s0b}.Deduplicate()
		Expect(dd).To(HaveLen(2))
		Expect(dd).To(ConsistOf(
			And(HaveField("Servants", ConsistOf(HaveField("Name", "c-a"))),
				HaveField("NetworkLabels", ConsistOf("net-a"))),
			And(HaveField("Servants", ConsistOf(HaveField("Name", "c-b"))),
				HaveField("NetworkLabels", ConsistOf("net-b"))),
		), "should keep nameless services/stand-alone containers separate")
	})

})
