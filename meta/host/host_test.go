// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package host

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/lxkns/containerizer"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("host metadata", func() {

	var cizer containerizer.Containerizer

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		ctx, cancel := context.WithCancel(context.Background())
		cizer = turtlefinder.New(func() context.Context { return ctx })

		DeferCleanup(func() {
			cancel()
			cizer.Close()
			Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("returns host metadata", func(ctx context.Context) {
		r := gostwire.Discover(ctx, cizer, nil)
		m := Metadata(r)
		Expect(m).To(HaveKeyWithValue("hostname", Not(BeEmpty())))
		Expect(m).To(HaveKeyWithValue("osrel-name", Not(BeEmpty())))
		Expect(m).To(HaveKeyWithValue("osrel-version", Not(BeEmpty())))
		Expect(m).To(HaveKeyWithValue("kernel-version", ContainSubstring("Linux version")))
	})

})
