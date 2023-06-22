// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/decorator/industrialedge"
	"github.com/thediveo/lxkns/model"
)

var _ = Describe("IE app icon cache", func() {

	It("compares old and new lists of container IDs", func() {
		Expect(isNewProject(nil, nil)).To(BeTrue())
		Expect(isNewProject([]string{"foo"}, nil)).To(BeTrue())
		Expect(isNewProject(nil, []string{"foo"})).To(BeTrue())
		Expect(isNewProject([]string{"foo"}, []string{"bar"})).To(BeTrue())

		Expect(isNewProject([]string{"foo"}, []string{"foo"})).To(BeFalse())
		Expect(isNewProject([]string{"foo", "bar"}, []string{"foo", "baz"})).To(BeFalse())
	})

	It("prunes and updates the cache", func() {
		cache := ieAppProjects{
			"foo": ieAppProject{
				Name:         "foo",
				ContainerIDs: []string{"foo-l1"},
				Title:        "foo App",
				IconData:     "null:",
			},
			"bar": ieAppProject{
				Name:         "bar",
				ContainerIDs: []string{"bar-42"},
				Title:        "bar App",
				IconData:     "null:",
			},
			"no10": ieAppProject{
				Name:         "no10",
				Title:        "no10 App",
				IconData:     "null:",
				ContainerIDs: []string{"bj", "dc"},
			},
		}
		engines := []*model.ContainerEngine{
			{
				Containers: []*model.Container{
					{
						ID:     "foo-l2",
						Flavor: industrialedge.IndustrialEdgeAppFlavor,
						Groups: []*model.Group{
							{
								Name: "foo",
								Type: composer.ComposerGroupType,
							},
						},
					},
					{
						ID:     "baz-666",
						Flavor: industrialedge.IndustrialEdgeAppFlavor,
						Groups: []*model.Group{
							{
								Name: "baz",
								Type: composer.ComposerGroupType,
							},
						},
					},
					{
						ID:     "bj",
						Flavor: industrialedge.IndustrialEdgeAppFlavor,
						Groups: []*model.Group{
							{
								Name: "no10",
								Type: composer.ComposerGroupType,
							},
						},
					},
				},
			},
		}
		for _, c := range engines[0].Containers {
			if g := c.Group(composer.ComposerGroupType); g != nil {
				g.Containers = []*model.Container{c}
			}
		}

		By("pruning and updating")
		news := cache.pruneAndUpdate(engines)
		Expect(cache).To(HaveKey("no10"), "endless project")
		Expect(cache).NotTo(HaveKey("bar"), "remove dead project")
		Expect(cache).NotTo(HaveKey("foo"), "remove replaced project")
		Expect(news).To(ConsistOf(
			HaveField("Name", "baz"),
			And(HaveField("Name", "foo"), HaveField("ContainerIDs", ConsistOf("foo-l2"))),
		), "newly found/replaced projects")

		By("adding new")
		for idx := range news {
			news[idx].Title = "Dummy"
			news[idx].IconData = "null:"
		}
		cache.add(news)
		Expect(cache).To(HaveKey("foo"))
		Expect(cache).To(HaveKey("baz"))

		By("again pruning and updating")
		news = cache.pruneAndUpdate(engines)
		Expect(news).To(BeEmpty())
		Expect(cache).To(HaveLen(3))
	})

})
