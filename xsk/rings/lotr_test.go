// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rings

import (
	"bytes"
	"os"
	"time"
	"unsafe"

	"github.com/mdlayher/ethernet"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/netns"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

func newRing[D Descriptor](
	size uint32,
	producer, consumer uint32,
) *Ring[D] {
	GinkgoHelper()
	Expect(size&(size-1)).To(BeZero(), "size must be a power of two")
	return &Ring[D]{
		producer:    &producer,
		consumer:    &consumer,
		descriptors: make([]D, size),

		size: size,
		mask: size - 1,
	}
}

var _ = Describe("Lord of the Rings", func() {

	Context("rings in general", func() {

		DescribeTable("fill level, free space, ... of a ring",
			func(size int, prod, cons uint32, used int) {
				r := newRing[uint64](uint32(size), uint32(prod), uint32(cons))
				Expect(r.Used()).To(Equal(used))
				Expect(r.Free()).To(Equal(int(r.size) - used))
				Expect(r.Empty()).To(Equal(used == 0))
				Expect(r.Full()).To(Equal(used == int(r.size)))
			},
			Entry("empty", 4, uint32(1), uint32(1), 0),
			Entry("full", 4, uint32(4), uint32(0), 4),
			Entry("non-wrapping", 4, uint32(3), uint32(1), 2),
			Entry("wrapping", 4, uint32(0), ^uint32(0), 1),
		)

		It("describes a ring", func() {
			r := &ProducerRing[uint64]{
				Ring: *newRing[uint64](16, 3, 0),
			}
			Expect(r.String()).To(Equal("rings.Ring[uint64] in use: 3"))
		})

		It("rejects unsettable rings", func() {
			Expect(NewFill(0, unix.XDPRingOffset{}, 32)).Error().NotTo(Succeed())
			Expect(NewCompletion(0, unix.XDPRingOffset{}, 32)).Error().NotTo(Succeed())
			Expect(NewTx(0, unix.XDPRingOffset{}, 32)).Error().NotTo(Succeed())
			Expect(NewRx(0, unix.XDPRingOffset{}, 32)).Error().NotTo(Succeed())
		})

	})

	Context("producer ring", func() {

		It("adds descriptor when there is room", func() {
			r := &ProducerRing[uint64]{
				Ring: *newRing[uint64](4, 5, 5),
			}
			Expect(r.Used()).To(Equal(0))
			Expect(r.Add(42)).To(BeTrue())
			Expect(r.Used()).To(Equal(1))
			Expect(r.descriptors[5&(4-1)]).To(Equal(uint64(42)))
		})

		It("rejects adding a descriptor when full", func() {
			r := &ProducerRing[uint64]{
				Ring: *newRing[uint64](4, 5+4, 5),
			}
			Expect(r.Add(666)).To(BeFalse())
		})

	})

	Context("consumer ring", func() {

		It("returns a descriptor when there are some", func() {
			c := &ConsumerRing[uint64]{
				Ring: *newRing[uint64](4, 5, 4),
			}
			c.descriptors[4&(4-1)] = 666
			Expect(c.Used()).To(Equal(1))
			d, ok := c.Next()
			Expect(ok).To(BeTrue())
			Expect(d).To(Equal(uint64(666)))
			Expect(c.Used()).To(Equal(0))
			_, ok = c.Next()
			Expect(ok).To(BeFalse())
		})

		It("rejects returning a descriptor when empty", func() {
			c := &ConsumerRing[uint64]{
				Ring: *newRing[uint64](4, 5, 5),
			}
			Expect(c.Used()).To(BeZero())
			d, ok := c.Next()
			Expect(ok).To(BeFalse())
			Expect(d).To(BeZero())
		})

	})

	// Testing TX and completion rings on the Real Thing™, but without
	// (circular) dependencies on the convenient umem and XSK stuff. So, back to
	// square one with bare bones testing, the hard way.
	It("creates rings for an XSK with umem and sends some frames into the bit bucket", func() {
		const chunkSize = 2048
		const chunkAmount = 64
		const FillRingSize = 64
		const CompletionRingSize = 64
		const RxRingSize = 64
		const TxRingSize = 64

		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a dummy device in an ephemeral network namespace")
		defer netns.EnterTransient()()
		dummy := dummy.NewTransientUp()

		By("allocating mmap'ed umem")
		umem := Successful(unix.Mmap(
			-1,
			0,
			chunkAmount*chunkSize,
			unix.PROT_READ|unix.PROT_WRITE,
			unix.MAP_PRIVATE|unix.MAP_ANONYMOUS|unix.MAP_POPULATE))
		defer func() { _ = unix.Munmap(umem) }()

		By("creating an XSK")
		xskfd := Successful(unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0))
		defer unix.Close(xskfd)

		By("registering the umem with the XSK")
		umemReg := unix.XDPUmemReg{
			Addr:     uint64(uintptr(unsafe.Pointer(&umem[0]))),
			Len:      uint64(len(umem)),
			Size:     chunkSize,
			Headroom: 0,
		}
		_, _, eno := unix.Syscall6(
			unix.SYS_SETSOCKOPT, uintptr(xskfd),
			unix.SOL_XDP, unix.XDP_UMEM_REG,
			uintptr(unsafe.Pointer(&umemReg)), unsafe.Sizeof(umemReg),
			0)
		Expect(eno).To(BeZero())

		By("configuring ring sizes")
		Expect(unix.SetsockoptInt(xskfd, unix.SOL_XDP, unix.XDP_UMEM_FILL_RING, FillRingSize)).
			To(Succeed())
		Expect(unix.SetsockoptInt(xskfd, unix.SOL_XDP, unix.XDP_UMEM_COMPLETION_RING, CompletionRingSize)).
			To(Succeed())
		Expect(unix.SetsockoptInt(xskfd, unix.SOL_XDP, unix.XDP_TX_RING, TxRingSize)).
			To(Succeed())
		Expect(unix.SetsockoptInt(xskfd, unix.SOL_XDP, unix.XDP_RX_RING, RxRingSize)).
			To(Succeed())

		By("binding the XSK to the dummy netdev")
		xskAddr := unix.SockaddrXDP{
			Flags:   0, // not using busypoll on purpose!
			Ifindex: uint32(dummy.Attrs().Index),
			QueueID: 0,
		}
		Expect(unix.Bind(xskfd, &xskAddr)).To(Succeed())

		By("retrieving ring offsets")
		var offsets unix.XDPMmapOffsets
		offsetsLen := int(unsafe.Sizeof(offsets))
		_, _, eno = unix.Syscall6(
			unix.SYS_GETSOCKOPT,
			uintptr(xskfd),
			unix.SOL_XDP,
			unix.XDP_MMAP_OFFSETS,
			uintptr(unsafe.Pointer(&offsets)),
			uintptr(unsafe.Pointer(&offsetsLen)),
			0)
		Expect(eno).To(BeZero())

		By("creating rings")
		fill := Successful(NewFill(xskfd, offsets.Fr, FillRingSize))
		defer fill.Close()
		completion := Successful(NewCompletion(xskfd, offsets.Cr, CompletionRingSize))
		defer completion.Close()
		tx := Successful(NewTx(xskfd, offsets.Tx, TxRingSize))
		defer tx.Close()
		rx := Successful(NewRx(xskfd, offsets.Rx, RxRingSize))
		defer rx.Close()

		// use an anonymous struct so that in case of a failed assertion
		// we see the flags for all four rings at a glance.
		ringflags := struct {
			Rx, Tx, Fill, Completion uint32
		}{
			Rx:         rx.Flags(),
			Tx:         tx.Flags(),
			Fill:       fill.Flags(),
			Completion: completion.Flags(),
		}
		Expect(ringflags).To(And(
			HaveField("Rx", BeZero()),
			HaveField("Tx", Equal(uint32(unix.XDP_RING_NEED_WAKEUP))),
			HaveField("Fill", BeZero()),
			HaveField("Completion", BeZero()),
		), "ring flags: %+v", ringflags)

		By("establishing a descriptor pool")
		descpool := NewDescriptorPool(chunkSize, 0, chunkAmount)

		By("generating a packet template")
		const experimentalEthType = 0xffee // https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml#ieee-802-numbers-1
		mac := dummy.Attrs().HardwareAddr
		f := ethernet.Frame{
			Destination: mac,
			Source:      mac,
			EtherType:   experimentalEthType,
			Payload:     bytes.Repeat([]byte("HELO"), 100),
		}
		frame := Successful(f.MarshalBinary())

		Consistently(completion.Used).Within(1 * time.Second).ProbeEvery(10 * time.Millisecond).
			Should(BeZero())

		By("scheduling packets for transmission and picking up completed descriptors")
		for i := 0; i < 32; i++ {
			txChunkAddr, ok := descpool.Get()
			Expect(ok).To(BeTrue())

			copy(umem[txChunkAddr:], frame)
			Expect(tx.Add(unix.XDPDesc{
				Addr: txChunkAddr,
				Len:  uint32(len(frame)),
			})).To(BeTrue())

			if tx.NeedsWakeup() {
				// poke the XSK, simply using poll with POLLOUT; see also
				// xsk_poll in xsk.c:
				// https://elixir.bootlin.com/linux/v6.9.4/source/net/xdp/xsk.c#L990
				fds := []unix.PollFd{
					{Fd: int32(xskfd), Events: unix.POLLOUT},
				}
				_, err := unix.Poll(fds, 0)
				Expect(err).NotTo(HaveOccurred())
			}

			Eventually(tx.Used).Within(100 * time.Millisecond).ProbeEvery(1 * time.Millisecond).
				Should(BeZero())
			Eventually(completion.Used).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
				ShouldNot(BeZero())
			complChunkAddr, ok := completion.Next()
			Expect(ok).To(BeTrue())
			Expect(complChunkAddr).To(Equal(txChunkAddr))
		}
	})

})
