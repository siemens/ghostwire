// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"slices"
	"time"

	"github.com/mdlayher/ethernet"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/link"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/xsk/rings"
	"github.com/siemens/ghostwire/v2/xsk/xdpnetdev"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("playing XDP socket ping-pong", func() {

	const (
		experimentalEthType = 0xffee // https://www.iana.org/assignments/ieee-802-numbers/ieee-802-numbers.xhtml#ieee-802-numbers-1

		ethTypeOffset = 12

		chunkAmount    = 32
		chunkSize      = 2048
		txSize         = chunkAmount / 2
		rxSize         = txSize
		fillSize       = rxSize
		completionSize = txSize

		packets = 2000

		throttle = 1 * time.Millisecond
		step     = 250
	)

	var payload = bytes.Repeat([]byte("HELO"), 100)

	// This is an end-to-end test using two XDP sockets, each bound to its own
	// MACVLAN virtual network interfaces, and the two MACVLAN network
	// interfaces both tied to a dummy virtual network interface.
	//
	// Now, as we're using XDP on MACVLANs, they usually tend to send to their
	// siblings as "run to completion". This means that there is no limited
	// "link bandwidth", only processing limits, memory bandwidth limits, and
	// scheduling (CPU time) limits. In consequence, we're throttling our
	// transmission in order to not overrun the receiver, losing frames.
	//
	// Moreover, if the receiving MACVLAN doesn't have any fill ring chunks
	// available, the frames will be lost. Thus, we need to wait until the
	// receiver part in this test signals that he've placed chunks into its fill
	// ring.
	It("ping-pongs", func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		defer netns.EnterTransient()()

		By("establishing an isolated virtual network infrastructure")
		dmy := dummy.NewTransientUp()
		macvlan1 := macvlan.NewTransient(dmy)
		link.EnsureUp(macvlan1)
		macvlan2 := macvlan.NewTransient(dmy)
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

		By("creating players")
		xskPlayer1 := Successful(New(macvlan1.Attrs().Index, 0,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithTxRingSize(txSize),
			WithoutRxRing(),
			WithFillRingSize(fillSize),
			WithCompletionRingSize(completionSize)))
		DeferCleanup(xskPlayer1.Close)

		xskPlayer2 := Successful(New(macvlan2.Attrs().Index, 0,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithoutTxRing(),
			WithRxRingSize(rxSize),
			WithFillRingSize(fillSize),
			WithCompletionRingSize(completionSize)))
		DeferCleanup(xskPlayer2.Close)

		By("installing an XSK director into the second MACVLAN and registering the RX XDP socket")
		netdev2 := Successful(xdpnetdev.NewByIndex(macvlan2.Attrs().Index))
		DeferCleanup(netdev2.Release)
		Expect(netdev2.AddXsk(0, xskPlayer2.Fd())).To(Succeed())
		DeferCleanup(netdev2.RemoveXsk, 0)

		By("creating a template frame")
		f := ethernet.Frame{
			Destination: mac2,
			Source:      mac1,
			EtherType:   experimentalEthType,
			Payload:     payload,
		}
		frame := Successful(f.MarshalBinary())
		frameLen := uint32(len(frame))

		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		// we need to ensure that the fill and RX rings are warmed up before we
		// start transmission: otherwise, the packets will be dropped as there
		// is no room for them and we're on a minimalist virtual MACVLAN netdev.
		prattle := make(chan struct{})

		// player1 TX frame generator
		txdone := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer func() {
				By("sending ended")
				close(txdone)
			}()

			By("sending packets")

			tx := xskPlayer1.Tx()
			completion := xskPlayer1.Completion()
			umem := xskPlayer1.Umem()
			dpool := rings.NewDescriptorPool(chunkSize, 0, chunkAmount/2)

			fds := []unix.PollFd{
				{Fd: int32(xskPlayer1.Fd()), Events: unix.POLLOUT},
			}

			// Don't start babbling until the receiver has put some chunks into
			// its fill ring.
			Eventually(prattle).WithContext(ctx).Within(1 * time.Second).ProbeEvery(1 * time.Millisecond).
				Should(BeClosed())

			// Ouch, we cannot prime the completion ring!

			for i := 1; i <= packets; i++ {
				var chunkAddr uint64
				for {
					// Abort if context is done/cancelled...
					if ctx.Err() != nil {
						return
					}
					// Sleep-wait for the TX ring to accept more frames/chunks.
					// This also pokes the kernel to notice any pending TX ring
					// elements, if any. A side effect of this happy snail path
					// is that we will basically always see an empty TX ring...
					n := Successful(unix.Poll(fds, 100 /* ms */))
					if n == 0 {
						continue
					}
					// TX is ready to accept a chunk; now get either one from
					// the pool to get the ball rolling, otherwise from the
					// completion ring. We prefer the completion ring, but it is
					// important that the pool is sized the same as the TX ring
					// as otherwise we could slowly get stuck over the course of
					// a few iterations.
					var ok bool
					if chunkAddr, ok = completion.Next(); !ok {
						if chunkAddr, ok = dpool.Get(); !ok {
							continue
						}
					}
					break
				}
				if i > 0 {
					time.Sleep(throttle)
				}
				// ...got a chunk, so fill it, and then ship it!
				copy(umem[chunkAddr:], frame)
				umem[chunkAddr+uint64(frameLen)-1] = byte(i & 0xff)
				tx.Add(unix.XDPDesc{
					Addr: chunkAddr,
					Len:  frameLen,
				})
				if i%step == 0 {
					By(fmt.Sprintf("%d packets sent", i))
				}
			}
			// A final kick on our way out...
			Expect(unix.Poll(fds, 0 /* ms */)).Error().NotTo(HaveOccurred())
		}()

		// Player2 RX frame generator
		rxdone := make(chan struct{})
		go func() {
			defer GinkgoRecover()
			defer func() {
				By("receiving ended")
				close(rxdone)
			}()

			By("receiving packets")

			rx := xskPlayer2.Rx()
			fill := xskPlayer2.Fill()
			umem := xskPlayer2.Umem()
			dpool := rings.NewDescriptorPool(chunkSize, chunkAmount/2, chunkAmount)

			fds := []unix.PollFd{
				{Fd: int32(xskPlayer2.Fd()), Events: unix.POLLIN},
			}

			By("priming the RX fill ring")
			// before we start churning, prime the fill ring with descriptors.
			// Since we don't do much received packet processing it's probably
			// fine to have half of the umem for fill/RX and make the
			// corresponding rings the same size.
			for fill.Free() > 0 {
				desc, ok := dpool.Get()
				if !ok {
					break
				}
				fill.Add(desc)
			}

			// Now let's receive frames; we keep juggling just with the chunks
			// that we've put into the fill ring, none more.
			close(prattle)
			for i := 1; i <= packets; i++ {
				if i%step == 0 {
					By(fmt.Sprintf("%d packets received", i))
				}
				for {
					// Abort if context is done/cancelled...
					if ctx.Err() != nil {
						return
					}
					// Sleep-wait for the RX ring to contain at least one chunk;
					// and yes, this is slow and inefficent, but this is just a
					// test after all!
					for {
						switch _, err := unix.Poll(fds, 100 /* ms */); err {
						case unix.EINTR, unix.EAGAIN:
							continue
						default:
							Expect(err).NotTo(HaveOccurred())
						}
						break
					}
					// Fetch the next received frame and check it...
					desc, ok := rx.Next()
					if !ok {
						continue
					}
					if desc.Len != frameLen ||
						!slices.Equal(umem[desc.Addr+ethTypeOffset:desc.Addr+ethTypeOffset+2], frame[12:14]) {
						continue
					}
					// Put the chunk into the fill ring to be filled anew.
					fill.Add(desc.Addr)
					break
				}
			}
		}()

		Eventually(rxdone).WithContext(ctx).Within(20 * time.Second).ProbeEvery(100 * time.Millisecond).Should(BeClosed())
		Eventually(txdone).WithContext(ctx).Within(1 * time.Second).ProbeEvery(100 * time.Millisecond).Should(BeClosed())

		By("cancelling context while shutting down")
		cancel()
	})

})
