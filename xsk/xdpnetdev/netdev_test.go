// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xdpnetdev

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/mdlayher/ethernet"
	"github.com/mdlayher/packet"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

const (
	experimentalEthType = 0xffee // https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml#ieee-802-numbers-1
	pings               = 10
	pingInterval        = 100 * time.Millisecond
)

var payload = bytes.Repeat([]byte("HELO"), 100)

var _ = Describe("XDP processing netdevs", Ordered, func() {

	BeforeAll(func() {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}
	})

	It("rejects an unknown netdev name", func() {
		defer netns.EnterTransient()()
		Expect(NewByName("foobar")).Error().To(HaveOccurred())
	})

	It("rejects registering XSKs for non-existing netdev queue", func() {
		defer netns.EnterTransient()()

		dmy := dummy.NewTransient()
		netdev := Successful(NewByName(dmy.Attrs().Name))
		defer netdev.Release()
		Expect(netdev.AddXsk(42, 0)).To(MatchError(ContainSubstring("queue index 42")))
		Expect(netdev.RemoveXsk(42)).To(MatchError(ContainSubstring("queue index 42")))
	})

	DescribeTable("loads an XDP program and sends/receives traffic",
		func(ctx context.Context, dropall bool) {
			By("creating a new network namespace and entering it for the rest of this test")
			// nota bene: this also helps making systemd-networkd keeping its
			// sticky fingers off our virtual network interfaces, and especially
			// not interfering with our tests by changing MAC addresses while
			// we're running out test...
			defer netns.EnterTransient()()

			By("creating two MACVLANs connected via a dummy network interface")
			dummyNif := dummy.NewTransientUp()
			macvlan1 := macvlan.NewTransient(dummyNif)
			link.EnsureUp(macvlan1)
			macvlan2 := macvlan.NewTransient(dummyNif)
			link.EnsureUp(macvlan2)

			macvlan1 = Successful(netlink.LinkByIndex(macvlan1.Attrs().Index))
			mac1 := macvlan1.Attrs().HardwareAddr
			Expect(mac1).To(HaveLen(6))
			Expect(mac1).NotTo(Equal(net.HardwareAddr{0, 0, 0, 0, 0, 0}))

			macvlan2 = Successful(netlink.LinkByIndex(macvlan2.Attrs().Index))
			mac2 := macvlan2.Attrs().HardwareAddr
			Expect(mac2).To(HaveLen(6))
			Expect(mac2).NotTo(Equal(net.HardwareAddr{0, 0, 0, 0, 0, 0}))

			Expect(mac1).NotTo(Equal(mac2))

			By(fmt.Sprintf("waiting for MACVLANs (%s-%s, %s-%s) to become operationally UP",
				macvlan1.Attrs().Name, macvlan1.Attrs().HardwareAddr.String(),
				macvlan2.Attrs().Name, macvlan2.Attrs().HardwareAddr.String()))
			link.EnsureUp(macvlan1)
			link.EnsureUp(macvlan2)

			By("attaching a default XDP program to the second MACVLAN")
			rxnetdev := Successful(
				newByIndex(macvlan2.Attrs().Index, dropall))
			defer rxnetdev.Release()

			By("systemd-notworkd check")
			mac2now := Successful(
				netlink.LinkByIndex(macvlan2.Attrs().Index)).Attrs().HardwareAddr
			Expect(mac2now).To(Equal(mac2),
				"systemd-notworkd trashed the netdev: original MAC %s, new MAC %s",
				mac2.String(), mac2now.String())

			By("opening data-link layer sockets")
			txconn := Successful(packet.Listen(
				&net.Interface{Index: macvlan1.Attrs().Index}, packet.Raw, experimentalEthType, nil))
			DeferCleanup(txconn.Close)
			rxconn := Successful(packet.Listen(
				&net.Interface{Index: macvlan2.Attrs().Index}, packet.Raw, experimentalEthType, nil))
			DeferCleanup(rxconn.Close)

			ctx, cancel := context.WithCancel(ctx)
			defer cancel()

			By("sending data-link layer PDUs")
			go func() {
				defer cancel()
				defer GinkgoRecover()
				f := ethernet.Frame{
					Destination: mac2,
					Source:      mac1,
					EtherType:   experimentalEthType,
					Payload:     payload,
				}
				frame := Successful(f.MarshalBinary())
				toAddr := packet.Addr{HardwareAddr: mac2}
				for range pings {
					By("sending something...")
					_, err := txconn.WriteTo(frame, &toAddr)
					Expect(err).NotTo(HaveOccurred())
					select {
					case <-ctx.Done():
						return
					case <-time.After(pingInterval):
					}
				}
			}()

			By("receiving data-link layer PDUs (or not)")
			received := 0
		receive:
			for {
				buffer := make([]byte, 1500)
				Expect(rxconn.SetReadDeadline(time.Now().Add(1 * time.Second))).To(Succeed())
				n, fromAddr, err := rxconn.ReadFrom(buffer)
				select {
				case <-ctx.Done():
					break receive
				default:
				}
				if err != nil && dropall && strings.Contains(err.Error(), "i/o timeout") {
					continue
				}
				Expect(err).NotTo(HaveOccurred())
				By("...received something")
				f := ethernet.Frame{}
				Expect(f.UnmarshalBinary(buffer[:n])).To(Succeed())
				Expect(f.EtherType).To(Equal(ethernet.EtherType(experimentalEthType)))
				Expect(fromAddr.(*packet.Addr).HardwareAddr).To(Equal(mac1))
				Expect(f.EtherType).To(Equal(ethernet.EtherType(experimentalEthType)))
				Expect(len(f.Payload)).To(BeNumerically(">=", len(payload)))
				Expect(f.Payload[:len(payload)]).To(Equal(payload))
				received++
			}

			if !dropall {
				Expect(received).To(BeNumerically(">=", (2*pings)/3), "too much packet loss")
			} else {
				Expect(received).To(BeZero())
			}

		},
		Entry("receives passed-on packets", false),
		Entry("drops all packets", true),
	)

})
