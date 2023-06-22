// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"
	"os"
	"syscall"
	"time"

	"github.com/thediveo/lxkns/model"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/onsi/gomega/gstruct"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

var _ = Describe("discovers transport ports", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	It("returns state descriptions", func() {
		Expect(TCP_LISTEN.String()).To(Equal("listening"))
		Expect(SocketState(0).String()).To(Equal("SocketState(0)"))

		Expect(Listening.String()).To(Equal("listening"))
		Expect(Connected.String()).To(Equal("connected"))
		Expect(SocketSimplifiedState(123).String()).To(Equal("SocketSimplifiedState(123)"))
	})

	It("handles invalid IPv4 and IPv6 addresses:ports", func() {
		ip, port := decodeSockAddrPort("abc")
		Expect(ip).To(BeNil())
		Expect(port).To(BeZero())

		ip, port = decodeSockAddrPort("abc:123z")
		Expect(ip).To(BeNil())
		Expect(port).To(BeZero())

		ip, port = decodeSockAddrPort("abcdefgh:1234")
		Expect(ip).To(BeNil())
		Expect(port).To(BeZero())

		ip, port = decodeSockAddrPort("abcz:1234")
		Expect(ip).To(BeNil())
		Expect(port).To(BeZero())
	})

	Context("Little Endian", func() {

		var pE bool

		BeforeEach(func() {
			pE = isLE
			isLE = true
		})

		AfterEach(func() {
			isLE = pE
		})

		It("decodes IPv4 and IPv6 socket addresses:ports", func() {
			ip, port := decodeSockAddrPort("3500007F:0035")
			Expect(ip.String()).To(Equal("127.0.0.53"))
			Expect(port).To(Equal(uint16(53)))

			ip, port = decodeSockAddrPort("0000000000000000FFFF00000100007F:1234")
			Expect(ip.String()).To(Equal("127.0.0.1")) // sic!
			Expect(port).To(Equal(uint16(0x1234)))
		})

		It("decodes socket state", func() {
			sm := socketToProcessMap{
				71934: []model.PIDType{666},
			}
			for _, testdata := range []struct {
				pass       bool
				proto      Protocol
				hexstate   string
				state      SocketState
				simplified SocketSimplifiedState
			}{
				{
					pass:       true,
					proto:      syscall.IPPROTO_TCP,
					hexstate:   "0A",
					state:      TCP_LISTEN,
					simplified: Listening,
				},
				{
					pass:       true,
					proto:      syscall.IPPROTO_TCP,
					hexstate:   "01",
					state:      TCP_ESTABLISHED,
					simplified: Connected,
				},
				{
					pass:       true,
					proto:      syscall.IPPROTO_TCP,
					hexstate:   "07",
					state:      TCP_CLOSE,
					simplified: Unconnected,
				},
				{
					pass:       false,
					proto:      syscall.IPPROTO_TCP,
					hexstate:   "FF",
					state:      0xFF,
					simplified: 0,
				},
				{
					pass:       true,
					proto:      syscall.IPPROTO_UDP,
					hexstate:   "07",
					state:      UDP_LISTEN,
					simplified: Listening,
				},
				{
					pass:       true,
					proto:      syscall.IPPROTO_UDP,
					hexstate:   "01",
					state:      UDP_ESTABLISHED,
					simplified: Connected,
				},
				{
					pass:       false,
					proto:      syscall.IPPROTO_UDP,
					hexstate:   "02",
					state:      0x02,
					simplified: Unconnected,
				},
			} {
				ps := newProcessSocket(
					"0: 0100007F:2414 01000000:0001 "+testdata.hexstate+" 00000000:00000000 00:00000000 00000000   996        0 71934 1 0000000000000000 100 0 0 10 0",
					unix.AF_INET,
					int(testdata.proto),
					sm)
				if testdata.pass {
					Expect(ps).To(MatchFields(IgnoreExtras, Fields{
						"Family":          Equal(AddressFamily(unix.AF_INET)),
						"Protocol":        Equal(testdata.proto),
						"LocalIP":         Equal(net.ParseIP("127.0.0.1").To4()),
						"LocalPort":       Equal(uint16(0x2414)),
						"RemoteIP":        Equal(net.ParseIP("0.0.0.1").To4()),
						"RemotePort":      Equal(uint16(1)),
						"State":           Equal(testdata.state),
						"SimplifiedState": Equal(testdata.simplified),
						"PIDs":            ConsistOf(model.PIDType(666)),
					}), "testdata %+v", testdata)
				} else {
					Expect(ps).NotTo(MatchFields(IgnoreExtras, Fields{
						"Family":          Equal(AddressFamily(unix.AF_INET)),
						"Protocol":        Equal(testdata.proto),
						"LocalIP":         Equal(net.ParseIP("127.0.0.1").To4()),
						"LocalPort":       Equal(uint16(0x2414)),
						"RemoteIP":        Equal(net.ParseIP("0.0.0.1").To4()),
						"RemotePort":      Equal(uint16(1)),
						"State":           Equal(testdata.state),
						"SimplifiedState": Equal(testdata.simplified),
						"PIDs":            ConsistOf(model.PIDType(666)),
					}), "testdata %+v", testdata)
				}
			}
		})

		It("handles empty socket information", func() {
			Expect(discoverSockets("./test/proc", 0, unix.AF_INET6, syscall.IPPROTO_TCP, nil)).To(BeEmpty())
		})

		It("discoverSockets() panics on nonsense address families and transport protocols", func() {
			Expect(func() {
				_ = discoverSockets("./test/proc", 0, unix.AF_APPLETALK, syscall.IPPROTO_TCP, nil)
			}).To(Panic())
			Expect(func() {
				_ = discoverSockets("./test/proc", 0, unix.AF_INET, syscall.IPPROTO_ICMP, nil)
			}).To(Panic())
		})

		It("discovers process sockets", func() {
			sm := socketToProcessMap{
				71934: []model.PIDType{666},
			}
			sox := discoverSockets("./test/proc", 0, unix.AF_INET, syscall.IPPROTO_TCP, sm)
			Expect(sox).To(BeEmpty())

			sox = discoverSockets("./test/proc", 666, unix.AF_INET, syscall.IPPROTO_TCP, sm)
			Expect(sox).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"LocalIP": Equal(net.ParseIP("127.0.0.1").To4()),
				}),
				MatchFields(IgnoreExtras, Fields{
					"LocalIP": Equal(net.ParseIP("127.0.0.2").To4()),
				}),
			))

			sox = discoverSockets("./test/proc", 666, unix.AF_INET6, syscall.IPPROTO_UDP, sm)
			Expect(sox).To(ConsistOf(
				MatchFields(IgnoreExtras, Fields{
					"LocalIP": Equal(net.ParseIP("127.0.0.10").To4()),
				}),
			))
		})

	})

	Context("Big Endian", func() {

		var pE bool

		BeforeEach(func() {
			pE = isLE
			isLE = false
		})

		AfterEach(func() {
			isLE = pE
		})

		It("decodes IPv4 and IPv6 socket addresses:ports", func() {
			ip, port := decodeSockAddrPort("7F000035:0035")
			Expect(ip.String()).To(Equal("127.0.0.53"))
			Expect(port).To(Equal(uint16(53)))

			ip, port = decodeSockAddrPort("00000000000000000000FFFF7F000001:1234")
			Expect(ip.String()).To(Equal("127.0.0.1")) // sic!
			Expect(port).To(Equal(uint16(0x1234)))
		})

	})

	It("discovers its own socket's inode-to-PID mapping", func() {
		// let's create a dummy socket we then next try to find.
		sock, err := net.Listen("tcp", ":0")
		Expect(err).NotTo(HaveOccurred())
		defer sock.Close()
		sockport := sock.Addr().(*net.TCPAddr).Port
		Expect(sockport).NotTo(BeZero())

		sockmap := discoverAllSockInodes("/non-exisiting")
		Expect(sockmap).To(BeEmpty())

		sockmap = discoverAllSockInodes("/proc")
		Expect(sockmap).NotTo(BeEmpty())
		pid := os.Getpid()

		Expect(sockmap).To(ContainElement(
			ConsistOf(model.PIDType(pid)),
		))

		sock.Close()
		sockmap = discoverAllSockInodes("/proc")
		Expect(sockmap).NotTo(BeEmpty())
		Expect(sockmap).NotTo(ContainElement(
			ConsistOf(model.PIDType(pid)),
		))
	})

})
