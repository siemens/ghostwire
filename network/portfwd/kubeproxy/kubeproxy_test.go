// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package kubeproxy

import (
	"net"

	"github.com/google/nftables/expr"
	"github.com/google/nftables/xt"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("kube-proxy port forwarding", func() {

	Context("transport protocol", func() {

		It("ignores Cmp expressions with other data", func() {
			proto, ok := getTcpUdp(&expr.Cmp{})
			Expect(ok).To(BeFalse())
			Expect(proto).To(BeZero())

			proto, ok = getTcpUdp(&expr.Cmp{Data: []byte{1, 2, 3}})
			Expect(ok).To(BeFalse())
			Expect(proto).To(BeZero())

			proto, ok = getTcpUdp(&expr.Cmp{Data: []byte{0}})
			Expect(ok).To(BeFalse())
			Expect(proto).To(BeZero())
		})

		It("returns port", func() {
			port, ok := getTcpUdp(&expr.Cmp{Data: []byte{unix.IPPROTO_TCP}})
			Expect(ok).To(BeTrue())
			Expect(port).To(Equal("tcp"))

			port, ok = getTcpUdp(&expr.Cmp{Data: []byte{unix.IPPROTO_UDP}})
			Expect(ok).To(BeTrue())
			Expect(port).To(Equal("udp"))
		})

	})

	Context("IP addresses", func() {

		It("ignores Cmp expressions with other data", func() {
			ip, ok := getIPv46(&expr.Cmp{})
			Expect(ok).To(BeFalse())
			Expect(ip).To(BeZero())

			ip, ok = getIPv46(&expr.Cmp{Data: []byte{1, 2, 3}})
			Expect(ok).To(BeFalse())
			Expect(ip).To(BeZero())
		})

		It("returns port", func() {
			ip, ok := getIPv46(&expr.Cmp{Data: []byte(net.ParseIP("fe80::dead:beef"))})
			Expect(ok).To(BeTrue())
			Expect(ip).To(Equal(net.ParseIP("fe80::dead:beef")))
		})

	})

	Context("ports", func() {

		It("ignores Cmp expressions with other data", func() {
			port, ok := getPort(&expr.Cmp{})
			Expect(ok).To(BeFalse())
			Expect(port).To(BeZero())

			port, ok = getPort(&expr.Cmp{Data: []byte{1, 2, 3}})
			Expect(ok).To(BeFalse())
			Expect(port).To(BeZero())
		})

		It("returns port", func() {
			port, ok := getPort(&expr.Cmp{Data: []byte{1, 2}})
			Expect(ok).To(BeTrue())
			Expect(port).To(Equal(uint16(0x0102)))
		})

	})

	Context("jump verdicts", func() {

		It("ignores non-service jump targets other verdicts", func() {
			chain, ok := getJumpVerdictChain(&expr.Verdict{})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictChain(&expr.Verdict{
				Kind:  expr.VerdictContinue,
				Chain: "foobar",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "hellorld",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())
		})

		It("returns the service target chain name", func() {
			chain, ok := getJumpVerdictChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "KUBE-SVC-HELLORLD",
			})
			Expect(ok).To(BeTrue())
			Expect(chain).NotTo(BeEmpty())

		})

	})

	Context("comment expressions", func() {

		It("ignores non-comment Match expressions", func() {
			comment, ok := getComment(&expr.Match{})
			Expect(ok).To(BeFalse())
			Expect(comment).To(BeEmpty())

			comment, ok = getComment(&expr.Match{
				Name: "comment",
			})
			Expect(ok).To(BeFalse())
			Expect(comment).To(BeEmpty())
		})

		It("returns comments", func() {
			xtc := xt.Comment("Hellorld!")
			comment, ok := getComment(&expr.Match{
				Name: "comment",
				Info: &xtc,
			})
			Expect(ok).To(BeTrue())
			Expect(comment).To(Equal("Hellorld!"))
		})

	})

})
