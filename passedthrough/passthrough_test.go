// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package passedthrough

import (
	"os"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("passthrough original ifnames", func() {

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	DescribeTable("altname prefix value extraction",
		func(prefix string, expected string, expectedok bool) {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransient()
			Expect(netlink.LinkAddAltName(dmy, "colorofmagic=twoflower")).To(Succeed())
			dmy = Successful(netlink.LinkByName(dmy.Attrs().Name))
			actual, ok := AltnameValueOk(dmy, prefix)
			Expect(ok).To(Equal(expectedok))
			Expect(actual).To(Equal(expected))
		},
		Entry(nil, "foobar", "", false),
		Entry(nil, "colorofmagic=", "twoflower", true),
	)

	Context("original ifnames", func() {

		It("returns an empty original ifname when not known", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransient()
			dmy = Successful(netlink.LinkByName(dmy.Attrs().Name))
			Expect(OriginalIfname(dmy)).To(BeEmpty())
		})

		It("returns the original ifname", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransient()
			Expect(netlink.LinkAddAltName(dmy, "foobar=barz")).To(Succeed())
			Expect(netlink.LinkAddAltName(dmy, AltNameOriginalIfnamePrefix+"twoflower")).To(Succeed())

			dmy = Successful(netlink.LinkByName(dmy.Attrs().Name))
			Expect(OriginalIfname(dmy)).To(Equal("twoflower"))
		})

	})

	Context("docker network of passed-through nif", func() {

		It("returns an empty ID when not known", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransient()
			dmy = Successful(netlink.LinkByName(dmy.Attrs().Name))
			Expect(DockernetID(dmy)).To(BeEmpty())
		})

		It("returns the Docker network ID", func() {
			defer netns.EnterTransient()()
			dmy := dummy.NewTransient()
			Expect(netlink.LinkAddAltName(dmy, "foobar=barz")).To(Succeed())
			Expect(netlink.LinkAddAltName(dmy, AltNameDockerNetworkPrefix+"b0cce55ed.cafe")).To(Succeed())

			dmy = Successful(netlink.LinkByName(dmy.Attrs().Name))
			Expect(DockernetID(dmy)).To(Equal("b0cce55ed"))
		})

	})

})
