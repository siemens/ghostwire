// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package engines

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/whalewatcher/v2/engineclient/moby"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("container engines metadata", func() {

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

	It("finds OS release information in edge core container", func(ctx context.Context) {
		By("spinning up a Docker container")
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.metadata.engines=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		dummy := Successful(sess.Run(ctx, "busybox:latest",
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithCombinedOutput(GinkgoWriter),
		))
		_ = Successful(dummy.PID(ctx))

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			ShouldNot(BeEmpty())

		Expect(Metadata(r)).To(HaveKeyWithValue(
			"container-engines", ContainElement(And(
				HaveField("ID", Not(BeEmpty())),
				HaveField("Type", moby.Type),
				HaveField("Version", Not(BeEmpty())),
			))))
	})

})
