// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	lxkns "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/spacetest/netns"

	"github.com/siemens/ghostwire/v2/xsk"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

var _ = Describe("discovering XSKs", func() {

	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		Eventually(Goroutines).Within(2 * time.Second).ProbeEvery(10 * time.Millisecond).
			ShouldNot(HaveLeaked(goodgos))
		Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
	})

	It("discovers our canary XSK", func(ctx context.Context) {
		if os.Geteuid() != 0 {
			Skip("needs root")
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		By("creating a new network namespace and entering it for the rest of this test")
		// nota bene: this also helps making systemd-networkd keeping its
		// sticky fingers off our virtual network interfaces.
		defer netns.EnterTransient()()

		By("creating a dummy network interface")
		dummyNif := dummy.NewTransientUp()

		By("creating an XSK and umem")
		xsk := Successful(xsk.New(
			dummyNif.Attrs().Index,
			0,
			xsk.WithChunkAmount(256),
			xsk.WithHeadroom(32)))
		xskClose := sync.OnceFunc(func() { _ = xsk.Close() })
		defer xskClose()

		By("discovering our XSK related to our process")
		disco := lxkns.Namespaces(
			lxkns.WithStandardDiscovery(),
			lxkns.FromTasks(),
			lxkns.FromFds(), // also discovers socket ino-to-PID mapping
		)
		Expect(len(disco.Namespaces[model.NetNS])).To(BeNumerically(">=", 2))
		xsks := AllXSKs(disco)
		Expect(xsks).To(ContainElement(
			ContainElement(And(
				HaveField("Processes", ContainElement(
					HaveField("PID", model.PIDType(os.Getpid())))),
				HaveField("Diag.XDPInfo", And(
					HaveField("Ifindex", uint32(dummyNif.Attrs().Index)),
					HaveField("Umem", And(
						HaveField("Size", Not(BeZero())),
						HaveField("NumPages", Not(BeZero())),
						HaveField("ChunkSize", xsk.Options().ChunkSize),
						HaveField("Ifindex", uint32(dummyNif.Attrs().Index)),
					)),
				)),
				HaveField("Nifname", dummyNif.Attrs().Name),
			))))
	})

})
