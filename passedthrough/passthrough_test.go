// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package passedthrough

import (
	"os"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/netns"
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
