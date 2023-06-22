// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT
package mobydig

import (
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var a = Containers{
	{ID: "1", Name: "Foo", Type: "Delta"},
	{ID: "2", Name: "Bar", Type: "Omicron"},
}

var b = Containers{
	{ID: "1", Name: "Foo", Type: "Delta"},
	{ID: "3", Name: "Zoo", Type: "Gomega"},
}

var _ = Describe("container list operations", func() {

	It("finds a container in a list", func() {
		Expect(Containers(nil).Contains(a[0])).To(BeFalse())
		Expect(a.Contains(&model.Container{ID: "1", Name: "Foo", Type: "Omicron"})).To(BeFalse())
		Expect(a.Contains(a[1])).To(BeTrue())
	})

	It("shared containers", func() {
		Expect(Containers(nil).Shares(nil)).To(BeFalse())
		Expect(a.Shares(nil)).To(BeFalse())
		Expect(a.Shares(b)).To(BeTrue())
		Expect(a.Shares(Containers{b[1]})).To(BeFalse())
	})

	It("merges container lists", func() {
		Expect(Containers(nil).Merge(nil)).To(BeEmpty())
		m := a.Merge(b)
		Expect(a).To(HaveLen(2))
		Expect(b).To(HaveLen(2))
		Expect(m).To(HaveLen(3))
		Expect(m).To(ContainElements(
			HaveField("ID", a[0].ID),
			HaveField("ID", a[1].ID),
			HaveField("ID", b[1].ID),
		))
	})

})
