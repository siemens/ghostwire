// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package xsk

import (
	"os"
	"runtime"
	"time"

	"github.com/thediveo/caps/v2"
	"golang.org/x/sys/unix"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("XDP sockets needing capabilities", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(250 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	It("cannot create an AF_XDP socket without CAP_NET_RAW", func() {
		if os.Getuid() != 0 {
			Skip("needs root") // sic! ...we want the power and drop it
		}

		done := make(chan struct{})
		go func() {
			// run the real test on a separate go routine that we lock to a
			// throw-away OS-level thread (task in Linux user space parlance).
			defer GinkgoRecover()
			defer close(done)

			runtime.LockOSThread() // never unlock, so afterwards the thread gets thrown away

			before := Successful(caps.OfCurrentTask())
			Expect(before.Effective().Drop(caps.CAP_NET_RAW).ApplyToCurrentTask()).
				Error().NotTo(HaveOccurred(), "could not drop CAP_NET_RAW")

			// Please note that unix.Socket returns -1 for the fd instead of 0 in
			// case of an error. This would trip Gomega's error return pattern
			// checker so we need to write this spec the old way.
			_, err := unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0)
			Expect(err).To(HaveOccurred())

			Expect(before.ApplyToCurrentTask()).
				Error().NotTo(HaveOccurred(), "cannot regain CAP_NET_RAW")

			xskfd := Successful(unix.Socket(unix.AF_XDP, unix.SOCK_RAW, 0))
			Expect(unix.Close(xskfd)).To(Succeed())
		}()

		Eventually(done).Should(BeClosed())
	})

})
