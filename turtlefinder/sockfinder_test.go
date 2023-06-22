// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"net"
	"os"
	"syscall"
	"time"

	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

const testListeningUnixSocketPath = "/tmp/gw-turtlefinder-test.sock"

var _ = Describe("socket finder", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("finds Docker API unix socket", func() {
		sox := discoverListeningSox(model.PIDType(os.Getpid()))
		Expect(sox).To(ContainElement("/run/docker.sock"))
	})

	It("doesn't find non-existing canary listening unix socket", func() {
		soxpaths := matchProcSox(
			model.PIDType(os.Getpid()),
			discoverListeningSox(model.PIDType(os.Getpid())))
		Expect(soxpaths).NotTo(ContainElement(testListeningUnixSocketPath))
	})

	It("finds listening canary unix socket", func() {
		_ = syscall.Unlink(testListeningUnixSocketPath)

		lsock, err := net.Listen("unix", testListeningUnixSocketPath)
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = lsock.Close()
			_ = syscall.Unlink(testListeningUnixSocketPath)
		}()

		soxpaths := matchProcSox(
			model.PIDType(os.Getpid()),
			discoverListeningSox(model.PIDType(os.Getpid())))
		Expect(soxpaths).To(ContainElement(testListeningUnixSocketPath))
	})

})
