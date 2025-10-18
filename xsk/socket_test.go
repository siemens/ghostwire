// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"errors"
	"fmt"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/kr/pretty"
	"github.com/thediveo/caps"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/macvlan"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"github.com/siemens/ghostwire/v2/xsk/umem"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

// EventuallyNew repeatedly attempts to create (and bind) an XDP socket until
// the operation succeeds or any error other than EBUSY occurs (including a
// timeout).
//
// The rationale is that after closing an XDP socket that was bound to a
// specific TX/RX netdev queue it takes "some" time for background cleaning
// processes to be carried out. During this timing window any attempt to bind
// another new XDP socket to the same netdev's queue will fail with EBUSY. And
// this is what happens in unit tests, where individual tests each open an XDP
// socket, do something on it, then close it, and quickly move on to the next
// unit test: boom!
func EventuallyNew(ifindex int, queueid int, opts ...Option) (xsk *Socket) {
	GinkgoHelper()
	Eventually(func() error {
		var err error
		xsk, err = New(ifindex, queueid, opts...)
		if err != nil && !errors.Is(err, unix.EBUSY) {
			StopTrying(err.Error()).Now()
		}
		return err
	}).Within(2*time.Second).ProbeEvery(20*time.Millisecond).
		Should(Succeed(),
			"repeated failure to create XSK for netdev %d, queue id %d", ifindex, queueid)
	Expect(xsk).NotTo(BeNil(), "New XSK succeeded, but returned nothing")
	return
}

var _ = Describe("Creating XDP sockets", Ordered, func() {

	const macvlanQueueID = 0 // MACVLAN netdevs only have a single "hardware" (hah!) queue

	//var testlink netlink.Link

	BeforeAll(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}
	})

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	It("fails when lacking CAP_NET_RAW", func() {
		runtime.LockOSThread() // ensure throw-away OS thread
		netnsfd := netns.NewTransient()
		Expect(unix.Setns(netnsfd, unix.CLONE_NEWNET)).To(Succeed())
		dmy := dummy.NewTransient()

		taskcaps := Successful(caps.OfThisTask())
		taskcaps.Effective.Clear()
		caps.SetForThisTask(taskcaps)
		Expect(New(dmy.Attrs().Index, 0)).Error().To(HaveOccurred())
	})

	It("rejects failing options when creating an XDP socket", func() {
		Expect(New(0, 0, ForceCopy(), ForceZeroCopy())).Error().To(HaveOccurred())
		Expect(New(0, 0, WithUmemFd(0))).Error().To(HaveOccurred())
		Expect(New(0, 0, WithChunkAmount(^uint32(0)))).Error().To(HaveOccurred())
	})

	It("rejects an released shared umem", func() {
		umemp := umem.NewPartner(make([]byte, 42))
		umemp.Cease()
		Expect(New(0, 0, WithSharedUmem(&Socket{
			umem: umemp,
		}))).Error().To(MatchError(ContainSubstring("sharing XDP socket has already released")))
	})

	It("rejects forced zero copy on dummy netdev", func() {
		defer netns.EnterTransient()()
		dmy := dummy.NewTransient()
		Expect(New(dmy.Attrs().Index, 0, ForceZeroCopy())).Error().To(HaveOccurred())
	})

	It("releases resources correctly", func() {
		const (
			chunkAmount = 256
			chunkSize   = 2048
			headroom    = 32
			crSize      = 64
			frSize      = 128
			rxSize      = 16
			txSize      = 32
		)
		umemfd := Successful(umem.New(int64(chunkAmount) * int64(chunkSize)))
		umemClose := sync.OnceFunc(func() { unix.Close(umemfd) })
		defer umemClose()

		defer netns.EnterTransient()()
		testlink := macvlan.NewTransient(dummy.NewTransient())

		xsk := EventuallyNew(testlink.Attrs().Index, macvlanQueueID,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithHeadroom(headroom),
			WithUmemFd(umemfd),
			WithCompletionRingSize(crSize),
			WithFillRingSize(frSize),
			WithRxRingSize(rxSize),
			WithTxRingSize(txSize))
		xskClose := sync.OnceFunc(func() { xsk.Close() })
		defer xskClose()

		cookie := Successful(unix.GetsockoptUint64(xsk.fd, unix.SOL_SOCKET, unix.SO_COOKIE))
		xskClose()
		umemClose()

		Expect(xsk.fd).To(BeNumerically("<", 0))

		xsks := Successful(netlink.SocketDiagXDP())
		Expect(xsks).NotTo(ContainElement(
			HaveField("XDPDiagMsg.Cookie", [2]uint32{
				uint32(cookie), uint32(cookie >> 32),
			})))
	})

	It("creates an XDP socket", func() {
		const (
			chunkAmount = 256
			chunkSize   = 2048
			crSize      = 64
			frSize      = 128
			txSize      = 16
			rxSize      = 16
		)

		umemfd := Successful(umem.New(int64(chunkAmount) * int64(chunkSize)))
		umemClose := sync.OnceFunc(func() { unix.Close(umemfd) })
		defer umemClose()

		defer netns.EnterTransient()()
		testlink := macvlan.NewTransient(dummy.NewTransient())

		xsk := EventuallyNew(testlink.Attrs().Index, macvlanQueueID,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithHeadroom(32),
			WithUmemFd(umemfd),
			WithCompletionRingSize(crSize),
			WithFillRingSize(frSize),
			WithRxRingSize(rxSize),
			WithTxRingSize(txSize))
		xskClose := sync.OnceFunc(func() { xsk.Close() })
		defer xskClose()

		procinfo := Successful(os.Stat(fmt.Sprintf("/proc/self/fd/%d", xsk.fd)))

		xsks := Successful(netlink.SocketDiagXDP())
		var xskinfo *netlink.XDPDiagInfoResp
		Expect(xsks).To(ContainElement(
			HaveField("XDPDiagMsg.Ino", uint32(procinfo.Sys().(*syscall.Stat_t).Ino)), &xskinfo))

		Expect(xskinfo.XDPInfo.RxRingEntries).To(Equal(uint32(rxSize)))
		Expect(xskinfo.XDPInfo.TxRingEntries).To(Equal(uint32(txSize)))
		Expect(xskinfo.XDPInfo.UmemFillRingEntries).To(Equal(uint32(frSize)))
		Expect(xskinfo.XDPInfo.UmemCompletionRingEntries).To(Equal(uint32(crSize)))
	})

	// TODO: retire in favor of the wrapping test using XSK from transient network namespace?
	/*
		It("creates an XDP socket object from a plain XSK file descriptor", func() {
			umemfd := Successful(umem.New(int64(DefaultSocketConfiguration.ChunkAmount) * int64(DefaultSocketConfiguration.ChunkSize)))
			defer unix.Close(umemfd)

			defer netns.EnterTransient()()
			testlink := macvlan.NewTransient(dummy.NewTransient())

			xsk := EventuallyNew(testlink.Attrs().Index, macvlanQueueID, umemfd, &DefaultSocketConfiguration)
			defer xsk.Close()

			xsk2 := Successful(NewFromFD(xsk.fd, umemfd, &DefaultSocketConfiguration))
			Expect(xsk2).NotTo(BeNil())
			Expect(xsk.ifindex).To(Equal(testlink.Attrs().Index))
			Expect(xsk.queueid).To(Equal(macvlanQueueID))
		})
	*/

	It("creates two XDP sockets and discovers them", func() {
		const (
			chunkAmount = 256
			chunkSize   = 2048
			crSize      = 64
			frSize      = 128
			txSize      = 16
			rxSize      = 16
		)

		By("creating first XSK")
		umemfd1 := Successful(umem.New(int64(chunkAmount) * int64(chunkSize)))
		defer unix.Close(umemfd1)

		defer netns.EnterTransient()()
		testlink1 := macvlan.NewTransient(dummy.NewTransient())

		xsk1 := EventuallyNew(testlink1.Attrs().Index, macvlanQueueID,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithHeadroom(0),
			WithUmemFd(umemfd1),
			WithCompletionRingSize(crSize),
			WithFillRingSize(frSize),
			WithRxRingSize(rxSize),
			WithTxRingSize(txSize))
		defer xsk1.Close()

		By("creating second XSK")
		testlink2 := macvlan.NewTransient(dummy.NewTransient())

		umemfd2 := Successful(umem.New(int64(chunkAmount) * int64(chunkSize)))
		defer unix.Close(umemfd2)

		xsk2 := EventuallyNew(testlink2.Attrs().Index, macvlanQueueID,
			WithChunkAmount(chunkAmount),
			WithChunkSize(chunkSize),
			WithHeadroom(0),
			WithUmemFd(umemfd2),
			WithCompletionRingSize(crSize),
			WithFillRingSize(frSize),
			WithRxRingSize(rxSize),
			WithTxRingSize(txSize))

		defer xsk2.Close()

		By("querying NETLINK for all XSKs")
		xsks := Successful(netlink.SocketDiagXDP())
		Expect(len(xsks)).To(BeNumerically(">=", 2))
		Expect(xsks).To(ContainElements(
			HaveField("XDPInfo.Ifindex", uint32(testlink1.Attrs().Index)),
			HaveField("XDPInfo.Ifindex", uint32(testlink2.Attrs().Index)),
		))

		By("querying NETLINK for a specific XSK")
		cookie := Successful(unix.GetsockoptUint64(xsk1.fd, unix.SOL_SOCKET, unix.SO_COOKIE))
		var stat unix.Stat_t
		Expect(unix.Fstat(xsk1.fd, &stat)).To(Succeed())
		xskinfo := Successful(netlink.SocketXDPGetInfo(uint32(stat.Ino), 0))
		Expect(xskinfo).To(HaveField("XDPInfo.Ifindex", uint32(testlink1.Attrs().Index)),
			"wrong XSK %s for ino 0x%x cookie [0x%x, 0x%x]", pretty.Sprint(xskinfo), stat.Ino, uint32(cookie), uint32(cookie>>32))

		cookie = Successful(unix.GetsockoptUint64(xsk2.fd, unix.SOL_SOCKET, unix.SO_COOKIE))
		xskinfo = Successful(netlink.SocketXDPGetInfo(0, cookie))
		Expect(xskinfo.XDPInfo.Ifindex).To(Equal(uint32(testlink2.Attrs().Index)))
	})

	It("wraps a Socket object around an XSK file descriptor connected to another network namespace", func() {
		By("creating and entering a transient network namespace")
		leaveNetns := sync.OnceFunc(netns.EnterTransient())
		defer leaveNetns()

		By("creating a bound XSK")
		unboundxsk := Successful(unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0))
		defer unix.Close(unboundxsk)

		By("leaving the transient network namespace")
		leaveNetns()

		// TODO:
		/*
			By("wrapping the XSK fd and getting its diagnosis data")
			socket := Successful(NewFromFD(unboundxsk, 0, nil))
			Expect(socket).NotTo(BeNil())
			Expect(socket.fd).To(Equal(unboundxsk))
		*/
	})

})
