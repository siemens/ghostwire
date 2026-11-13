// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package umem

import (
	"os"
	"os/exec"
	"sync"
	"time"

	"golang.org/x/sys/unix"

	"github.com/onsi/gomega/gexec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

// just to test createTemp on all its unhappy paths...
type mockfs struct {
	failingOpenCount   int
	failingUnlinkCount int
	wasClosed          bool
}

func (t *mockfs) Open(path string, mode int, perm uint32) (int, error) {
	if t.failingOpenCount > 0 {
		t.failingOpenCount--
		return 0, os.ErrExist
	}
	return 42, nil
}

func (t *mockfs) Close(fd int) error {
	GinkgoHelper()
	Expect(fd).To(Equal(int(42)))
	t.wasClosed = true
	return nil
}

func (t *mockfs) Unlink(path string) error {
	if t.failingUnlinkCount > 0 {
		t.failingUnlinkCount--
		return os.ErrNotExist
	}
	return nil
}

var _ = Describe("fd-referenced umem", Ordered, func() {

	var pongPath string

	BeforeAll(func() {
		pongPath = Successful(gexec.Build("github.com/siemens/ghostwire/v2/xsk/umem/test/pong"))
	})

	AfterAll(func() {
		gexec.CleanupBuildArtifacts()
	})

	BeforeEach(func() {
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Filedescriptors).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
				ShouldNot(HaveLeakedFds(goodfds))
		})
	})

	Context("unhappy creating temporary shared memory file paths", func() {

		It("bails out after too many attempts", func() {
			mfs := mockfs{
				failingOpenCount: 10,
			}
			Expect(createTemp(&mfs, "/foobar-")).Error().To(MatchError(
				ContainSubstring("creating a temporary file")))
		})

		It("bails out if tmp file cannot be unlinked", func() {
			mfs := mockfs{
				failingOpenCount:   5,
				failingUnlinkCount: 1,
			}
			Expect(createTemp(&mfs, "/foobar-")).Error().To(MatchError(
				ContainSubstring("cannot unlink temporary file")))
			Expect(mfs.wasClosed).To(BeTrue())
		})

	})

	When("creating a temporary shared memory file", func() {

		It("reports invalid shmem path", func() {
			Expect(newShumem(42, "/this-does-NOT-exist/")).Error().To(HaveOccurred())
		})

		It("reports invalid size", func() {
			Expect(New(-1)).Error().To(HaveOccurred())
		})

		It("fails on an invalid shmem path", func() {
			cwd := Successful(os.Getwd())
			prefix := cwd + "/this-does-NOT-exist/"
			_, err := createTemp(&unixfs, prefix)
			Expect(err).To(HaveOccurred())
		})

		It("creates a temporary file inside shmem path and unlinks it", func() {
			fakeShmemDir := Successful(os.MkdirTemp("", "fake-shmem-dir-*"))
			defer func() {
				Expect(os.RemoveAll(fakeShmemDir)).To(Succeed())
			}()

			umemfd := Successful(createTemp(&unixfs, fakeShmemDir))
			defer func() {
				Expect(unix.Close(umemfd)).To(Succeed())
			}()
			Expect(os.ReadDir(fakeShmemDir)).To(BeEmpty())
		})

	})

	It("reports an error when attempting to an invalid fd", func() {
		Expect(Map(-1)).Error().To(HaveOccurred())
	})

	It("allocates a umem and maps it", func() {
		const length = 2048 * 64
		umemfd := Successful(New(length))
		DeferCleanup(unix.Close, umemfd)

		umem := Successful(Map(umemfd))
		DeferCleanup(Unmap, umem)
		Expect(umem).NotTo(BeNil())
		Expect(len(umem)).To(BeNumerically(">=", length))
	})

	It("plays umem ping-pong with another process", func() {
		const length = 128
		umemfd := Successful(New(length))
		closeUmemFd := sync.OnceFunc(func() {
			Expect(unix.Close(umemfd)).To(Succeed())
		})
		defer closeUmemFd()

		umem := Successful(Map(umemfd))
		Expect(umem).NotTo(BeNil())
		DeferCleanup(Unmap, umem)

		pong := exec.Command(pongPath)
		pong.ExtraFiles = []*os.File{os.NewFile(uintptr(umemfd), "umem")}
		session := Successful(gexec.Start(pong, GinkgoWriter, GinkgoWriter))
		defer session.Terminate()

		// wait for "PLAYER 2 READY" :p
		Eventually(func() byte { return umem[0] }).
			Within(5*time.Second).ProbeEvery(20*time.Millisecond).
			Should(Equal(byte(0x42)), "player 2 not ready")

		// https://stackoverflow.com/questions/17490033/do-i-need-to-keep-a-file-open-after-calling-mmap-on-it
		closeUmemFd()

		for _, pattern := range []byte{0x00, 0xff, 0x42, 0xaa, 0x55} {
			// ping...
			umem[0] = pattern
			// ...pong
			Eventually(func() byte { return umem[1] }).
				Within(2*time.Second).ProbeEvery(1*time.Microsecond).
				Should(Equal(byte(pattern+1)), "wanted %x", byte(pattern+1))
		}
	})

})
