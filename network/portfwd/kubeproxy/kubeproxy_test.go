// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package kubeproxy

import (
	"net"

	"github.com/google/nftables"
	"github.com/google/nftables/expr"
	"github.com/google/nftables/xt"
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/portfinder"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("kube-proxy port forwarding", func() {

	Context("detecting forwarded kube-proxy ports", func() {

		It("doesn't crash", func() {
			Expect(PortForwardings(nufftables.TableMap{}, nufftables.TableFamilyINet))
			Expect(PortForwardings(nil, nufftables.TableFamilyIPv4))
			Expect(PortForwardings(nufftables.TableMap{
				nufftables.TableKey{Name: "nat", Family: nufftables.TableFamilyIPv4}: &nufftables.Table{},
			}, nufftables.TableFamilyIPv4))
			Expect(PortForwardings(nufftables.TableMap{
				nufftables.TableKey{Name: "nat", Family: nufftables.TableFamilyIPv4}: &nufftables.Table{
					ChainsByName: map[string]*nufftables.Chain{
						kubeServicesChain: {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: nil,
									},
								},
							},
						},
					},
				},
			}, nufftables.TableFamilyIPv4))
		})

		It("detects forwarded ports", func() {
			comment := xt.Comment("foo")
			tables := nufftables.TableMap{
				{Name: "nat", Family: nufftables.TableFamilyIPv4}: &nufftables.Table{
					ChainsByName: map[string]*nufftables.Chain{
						kubeServicesChain: {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Cmp{
												Data: []byte{unix.IPPROTO_TCP},
											},
											&expr.Cmp{
												Data: net.ParseIP("1.2.3.4"),
											},
											&expr.Match{
												Name: "comment",
												Info: &comment,
											},
											&expr.Cmp{
												Data: []byte{1, 2},
											},
											&expr.Verdict{
												Kind:  expr.VerdictJump,
												Chain: kubeServiceChainPrefix + "HELLORLD",
											},
										},
									},
								},
							},
						},
						kubeServiceChainPrefix + "HELLORLD": {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Verdict{
												Kind:  expr.VerdictJump,
												Chain: kubeSeparationChainPrefix + "FOO",
											},
										},
									},
								},
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Verdict{
												Kind:  expr.VerdictJump,
												Chain: kubeSeparationChainPrefix + "BAR",
											},
										},
									},
								},
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Verdict{
												Kind:  expr.VerdictJump,
												Chain: kubeSeparationChainPrefix + "BAZ",
											},
										},
									},
								},
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Verdict{
												Kind:  expr.VerdictJump,
												Chain: kubeSeparationChainPrefix + "BAZZZZ",
											},
										},
									},
								},
							},
						},
						kubeSeparationChainPrefix + "FOO": {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Target{
												Name: "DNAT",
												Info: &xt.NatRange2{
													NatRange: xt.NatRange{
														MinIP:   net.ParseIP("10.20.30.40"),
														MinPort: 123,
													},
												},
											},
										},
									},
								},
							},
						},
						kubeSeparationChainPrefix + "BAR": {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Target{
												Name: "DNAT",
												Info: &xt.NatRange2{
													NatRange: xt.NatRange{
														MinIP:   net.ParseIP("10.20.30.44"),
														MinPort: 123,
													},
												},
											},
										},
									},
								},
							},
						},
						kubeSeparationChainPrefix + "BAZ": {
							Rules: []nufftables.Rule{
								{
									Rule: &nftables.Rule{
										Exprs: []expr.Any{
											&expr.Target{},
										},
									},
								},
							},
						},
					},
				},
			}
			portfwds := PortForwardings(tables, nufftables.TableFamilyIPv4)
			Expect(portfwds).To(ConsistOf(
				&portfinder.ForwardedPortRange{
					Protocol:       "tcp",
					IP:             net.ParseIP("1.2.3.4"),
					PortMin:        0x0102,
					PortMax:        0x0102,
					ForwardIP:      net.ParseIP("10.20.30.40"),
					ForwardPortMin: 123,
				},
				&portfinder.ForwardedPortRange{
					Protocol:       "tcp",
					IP:             net.ParseIP("1.2.3.4"),
					PortMin:        0x0102,
					PortMax:        0x0102,
					ForwardIP:      net.ParseIP("10.20.30.44"),
					ForwardPortMin: 123,
				},
			))
		})

	})

	It("matches DNAT target expressions", func() {
		dnat, ok := getDNAT(&expr.Target{Name: "hellorld"})
		Expect(ok).To(BeFalse())
		Expect(dnat).To(BeNil())

		dnat, ok = getDNAT(&expr.Target{Name: "DNAT"})
		Expect(ok).To(BeFalse())
		Expect(dnat).To(BeNil())

		dnat, ok = getDNAT(&expr.Target{
			Name: "DNAT",
			Info: &xt.NatRange2{
				BasePort: 42,
			},
		})
		Expect(ok).To(BeTrue())
		Expect(dnat).NotTo(BeNil())
		Expect(dnat.BasePort).To(Equal(uint16(42)))
	})

	Context("service provider chains", func() {

		It("returns nothing for non-existing chain", func() {
			Expect(serviceProviderChains(nil)).To(BeNil())
		})

		It("returns separated service provider chains", func() {
			chains := serviceProviderChains(&nufftables.Chain{
				Rules: []nufftables.Rule{
					{
						Rule: &nftables.Rule{
							Exprs: []expr.Any{
								&expr.Bitwise{},
							},
						},
					},
					{
						Rule: &nftables.Rule{
							Exprs: []expr.Any{
								&expr.Verdict{
									Kind:  expr.VerdictJump,
									Chain: kubeSeparationChainPrefix + "HELLORLD",
								},
							},
						},
					},
					{
						Rule: &nftables.Rule{
							Exprs: []expr.Any{
								&expr.Verdict{
									Kind:  expr.VerdictJump,
									Chain: kubeSeparationChainPrefix + "GOODBYE",
								},
							},
						},
					},
				},
			})
			Expect(chains).To(ConsistOf(
				kubeSeparationChainPrefix+"HELLORLD",
				kubeSeparationChainPrefix+"GOODBYE",
			))
		})

	})

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

	Context("service jump verdicts", func() {

		It("ignores non-service jump targets other verdicts", func() {
			chain, ok := getJumpVerdictServiceChain(&expr.Verdict{})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictServiceChain(&expr.Verdict{
				Kind:  expr.VerdictContinue,
				Chain: "foobar",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictServiceChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "hellorld",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())
		})

		It("returns the service target chain name", func() {
			chain, ok := getJumpVerdictServiceChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "KUBE-SVC-HELLORLD",
			})
			Expect(ok).To(BeTrue())
			Expect(chain).NotTo(BeEmpty())

		})

	})

	Context("separation jump verdicts", func() {

		It("ignores non-separation jump targets other verdicts", func() {
			chain, ok := getJumpVerdictSeparationChain(&expr.Verdict{})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictSeparationChain(&expr.Verdict{
				Kind:  expr.VerdictContinue,
				Chain: "foobar",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())

			chain, ok = getJumpVerdictSeparationChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "hellorld",
			})
			Expect(ok).To(BeFalse())
			Expect(chain).To(BeEmpty())
		})

		It("returns the service target chain name", func() {
			chain, ok := getJumpVerdictSeparationChain(&expr.Verdict{
				Kind:  expr.VerdictJump,
				Chain: "KUBE-SEP-HELLORLD",
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
