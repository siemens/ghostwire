// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"
	"syscall"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	"golang.org/x/sys/unix"
)

var _ = Describe("handles network addresses", func() {

	It("explodes addresses", func() {
		Expect(Address{Family: unix.AF_INET6, Address: net.ParseIP("fe80::dead:beef")}.Exploded()).
			To(Equal("fe80:0000:0000:0000:0000:0000:dead:beef"))

		Expect(Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.0.1").To4()}.Exploded()).
			To(Equal("127.000.000.001"))
		Expect(Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.0.1")}.Exploded()).
			To(Equal("127.000.000.001"))

		Expect(func() {
			Address{Family: unix.AF_INET, Address: net.ParseIP("fe80::1")}.Exploded()
		}).To(Panic())

		Expect(Exploded(net.IP([]byte{127, 0, 0, 1}))).To(Equal("127.000.000.001"))
		Expect(Exploded(net.ParseIP("fe80::dead:beef"))).To(Equal("fe80:0000:0000:0000:0000:0000:dead:beef"))

		Expect(exploded(net.IP([]byte{1, 2, 3}), 0)).To(Equal("?010203"))
	})

	Context("sorts addresses", func() {

		It("sorts IPv4 before IPv6", func() {
			addrs := Addresses{
				Address{Family: unix.AF_INET6, Address: net.ParseIP("fe80::dead:beef")},
				Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.0.1")},
			}
			addrs.Sort()
			Expect(addrs).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Family": Equal(unix.AF_INET),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Family": Equal(unix.AF_INET6),
				}),
			))
		})

		It("sorts IPv4 correctly", func() {
			addrs := Addresses{
				Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.100.1")},
				Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.20.1")},
			}
			addrs.Sort()
			Expect(addrs).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Address": Equal(net.ParseIP("127.0.20.1")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Address": Equal(net.ParseIP("127.0.100.1")),
				}),
			))
		})

		It("sorts IPv6 correctly", func() {
			addrs := Addresses{
				Address{Family: unix.AF_INET6, Address: net.ParseIP("fe80::1")},
				Address{Family: unix.AF_INET6, Address: net.ParseIP("fe8::1")}, // bad, really bad ;)
			}
			addrs.Sort()
			Expect(addrs).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"Address": Equal(net.ParseIP("fe8::1")),
				}),
				MatchFields(IgnoreExtras, Fields{
					"Address": Equal(net.ParseIP("fe80::1")),
				}),
			))
		})

		It("sorts IP with prefix lengths", func() {
			addrs := Addresses{
				Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.0.1"), PrefixLength: 24},
				Address{Family: unix.AF_INET, Address: net.ParseIP("127.0.0.1"), PrefixLength: 20},
			}
			addrs.Sort()
			Expect(addrs).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"PrefixLength": Equal(uint(20)),
				}),
				MatchFields(IgnoreExtras, Fields{
					"PrefixLength": Equal(uint(24)),
				}),
			))

		})

	})

	It("stringifies UDP/TCP protocol numbers", func() {
		Expect(Protocol(syscall.IPPROTO_TCP).String()).To(Equal("TCP"))
		Expect(Protocol(syscall.IPPROTO_UDP).String()).To(Equal("UDP"))
		Expect(Protocol(0).String()).To(Equal("Protocol(0)"))
	})

	It("stringifies IPv4/v6 address families", func() {
		Expect(AddressFamily(unix.AF_INET).String()).To(Equal("IPv4"))
		Expect(AddressFamily(unix.AF_INET6).String()).To(Equal("IPv6"))
		Expect(AddressFamily(0).String()).To(Equal("AddressFamily(0)"))
	})

	It("correctly stringifies IP addresses in port contexts", func() {
		Expect(IP(net.ParseIP("127.0.0.1")).String()).To(Equal("127.0.0.1"))
		Expect(IP([]byte{127, 0, 0, 1}).String()).To(Equal("127.0.0.1"))
		Expect(IP(net.ParseIP("fe80::1")).String()).To(Equal("[fe80::1]"))
	})

})
