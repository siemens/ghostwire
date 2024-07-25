// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package docker

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/google/nftables"
	"github.com/thediveo/morbyd"
	"github.com/thediveo/morbyd/run"
	"github.com/thediveo/morbyd/session"
	"github.com/thediveo/notwork/netns"
	"github.com/thediveo/nufftables"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("Docker port forwarding", Ordered, func() {

	var cntrPID int
	var hostPort uint16
	var cntrIP net.IP // 127.0.0.1 -> ?.?.?.?

	BeforeAll(func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test=ghostwire.network.portfwd.docker")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})
		By("creating a temporary Docker custom test network")
		netw := Successful(sess.CreateNetwork(ctx,
			"test-portfwd-docker"))
		By("spinning up a temporary test container, exposing a useless port on a random host port")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithNetwork(netw.ID),
			run.WithPublishedPort("127.0.0.1:1234/tcp"),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
		))
		cntrPID = Successful(cntr.PID(ctx))
		By("picking up the random exposed host port number")
		svcAddr := cntr.PublishedPort("1234/tcp").First().String() // there's only one
		Expect(svcAddr).To(MatchRegexp(`127\.0\.0\.1:\d+`))
		hostPort = uint16(Successful(strconv.ParseUint(strings.Split(svcAddr, ":")[1], 10, 16)))
		cntrIP = cntr.IP(ctx).To4()
		Expect(hostPort).To(BeNumerically(">", uint16(32767)))
	})

	It("discovers host's port forwarding to container", func() {
		conn := Successful(nftables.New(nftables.AsLasting()))
		DeferCleanup(func() {
			_ = conn.CloseLasting()
		})
		tables := Successful(nufftables.GetAllTables(conn))
		nattable := tables.Table("nat", nufftables.TableFamilyIPv4)
		Expect(nattable).NotTo(BeNil())
		Expect(nattable.ChainsByName).NotTo(BeEmpty())
		forwardedPorts := grabPortForwardings(nattable)
		Expect(forwardedPorts).To(ContainElement(And(
			HaveField("Protocol", "tcp"),
			HaveField("IP", net.ParseIP("127.0.0.1").To4()),
			HaveField("PortMin", hostPort),
			HaveField("PortMax", hostPort),
			HaveField("ForwardIP", cntrIP),
			HaveField("ForwardPortMin", uint16(1234)),
		)))
	})

	It("discovers container-local Docker embedded DNS port forwarding", func() {
		cntrnetnsf := Successful(os.Open(fmt.Sprintf("/proc/%d/ns/net", cntrPID)))
		DeferCleanup(func() { cntrnetnsf.Close() })
		var conn *nftables.Conn
		var err error
		netns.Execute(int(cntrnetnsf.Fd()), func() {
			conn, err = nftables.New(nftables.AsLasting())
		})
		Expect(err).NotTo(HaveOccurred())
		DeferCleanup(func() {
			_ = conn.CloseLasting()
		})
		tables := Successful(nufftables.GetAllTables(conn))
		nattable := tables.Table("nat", nufftables.TableFamilyIPv4)
		Expect(nattable).NotTo(BeNil())
		Expect(nattable.ChainsByName).NotTo(BeEmpty())
		forwardedPorts := grabPortForwardings(nattable)
		// Ensure to exactly match in order to catch any false positives; this
		// is possible in this case because we're looking at the nft inside the
		// container and thus know what should be there and what shouldn't.
		Expect(forwardedPorts).To(ConsistOf(
			And(
				HaveField("Protocol", "tcp"),
				HaveField("IP", net.ParseIP("127.0.0.11").To4()),
				HaveField("PortMin", uint16(53)),
				HaveField("PortMax", uint16(53)),
				HaveField("ForwardIP", net.ParseIP("127.0.0.11").To4()),
				HaveField("ForwardPortMin", Not(BeZero())),
			),
			And(
				HaveField("Protocol", "udp"),
				HaveField("IP", net.ParseIP("127.0.0.11").To4()),
				HaveField("PortMin", uint16(53)),
				HaveField("PortMax", uint16(53)),
				HaveField("ForwardIP", net.ParseIP("127.0.0.11").To4()),
				HaveField("ForwardPortMin", Not(BeZero())),
			),
		))
	})

})
