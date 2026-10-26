// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"os"

	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("network interface", func() {

	It("has all nif makers correctly registered", func() {
		collectNifMakers()
		for _, kind := range []string{
			"bridge",
			"macvlan",
			"veth",
			"vxlan",
		} {
			Expect(nifMakers).To(HaveKey(kind), "missing nif maker for kind %s", kind)
			Expect(nifMakers[kind]()).NotTo(BeNil())
		}
	})

	It("returns interfaces of a kind", func() {
		nifs := Interfaces{
			&NifAttrs{Name: "lo", Kind: ""},
			&VethAttrs{NifAttrs: NifAttrs{Name: "vethdead", Kind: "veth"}},
			&VethAttrs{NifAttrs: NifAttrs{Name: "vethbeef", Kind: "veth"}},
			&VxlanAttrs{NifAttrs: NifAttrs{Name: "vxlan666", Kind: "vxlan"}},
		}
		Expect(nifs.OfKind("veth")).To(ContainElements(
			HaveInterfaceKindAndName("veth", "vethdead"),
			HaveInterfaceKindAndName("veth", "vethbeef"),
		))
		Expect(nifs.OfKind("foobar")).To(BeEmpty())
		Expect(Interfaces{}.OfKind("lo")).To(BeEmpty())
	})

	It("picks up alternate names", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		defer netns.EnterTransient()()
		dmy := dummy.NewTransient()
		Expect(netlink.LinkAddAltName(dmy, "twoflower")).To(Succeed())
		Expect(netlink.LinkAddAltName(dmy, "rincewind")).To(Succeed())

		nlh := Successful(netlink.NewHandle())
		defer nlh.Close()
		nif := NewInterface(nlh, nil, Successful(nlh.LinkByName(dmy.Attrs().Name)))
		Expect(nif.Nif().AltNames).To(ConsistOf("twoflower", "rincewind"))
	})

	It("flags loopback interfaces", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		defer netns.EnterTransient()()
		dmy := dummy.NewTransient()

		nlh := Successful(netlink.NewHandle())
		defer nlh.Close()
		lo := NewInterface(nlh, nil, Successful(nlh.LinkByName("lo")))
		Expect(lo.Nif().Loopback).To(BeTrue())
		Expect(lo.Nif().Physical).To(BeFalse())
		nif := NewInterface(nlh, nil, Successful(nlh.LinkByName(dmy.Attrs().Name)))
		Expect(nif.Nif().Loopback).To(BeFalse())
	})

	When("sorting interfaces", func() {

		It("sorts lo first", func() {
			nifs := Interfaces{
				&NifAttrs{Name: "abc", Kind: ""},
				&NifAttrs{Name: "lo", Kind: ""},
			}
			nifs.Sort()
			Expect(nifs).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName("abc")))

			nifs.Sort()
			Expect(nifs).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName("abc")))

			nifs = Interfaces{
				&NifAttrs{Name: "lo", Kind: ""},
				&NifAttrs{Name: "zzz", Kind: ""},
			}
			nifs.Sort()
			Expect(nifs).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName("zzz")))
		})

		It("sorts by interface name", func() {
			nifs := Interfaces{
				&NifAttrs{Name: "zzz", Kind: ""},
				&NifAttrs{Name: "lo", Kind: ""},
				&NifAttrs{Name: "abc", Kind: ""},
			}
			nifs.Sort()
			Expect(nifs).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName("abc"),
				HaveInterfaceName("zzz")))

		})

	})

})
