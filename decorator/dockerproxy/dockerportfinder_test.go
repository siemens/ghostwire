// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package dockerproxy

import (
	"context"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/whalewatcher/v2/engineclient/moby"

	"github.com/siemens/ghostwire/v2/decorator"
	"github.com/siemens/ghostwire/v2/internal/discover"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	testWorkloadName = "gostwire-decorator-dockerproxy-test-workload"
)

var _ = Describe("dockernet decorator", func() {

	It("registers correctly", func() {
		Expect(plugger.Group[decorator.Decorate]().Plugins()).To(
			ContainElement("dockerportfinder"))
	})

	Context("when looking for Docker-managed networks", func() {

		BeforeEach(func() {
			goodfds := Filedescriptors()
			goodgos := Goroutines()
			DeferCleanup(func() {
				Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})

			DeferCleanup(slog.SetDefault, slog.Default())
			slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			})))
		})

		It("discovers port forwardings from dockerproxy processes", func(ctx context.Context) {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			sess := Successful(morbyd.NewSession(ctx,
				session.WithAutoCleaning("test.decorator.dockerproxy=")))
			DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

			By("creating a test workload and connecting it to the test network, publishing a random(xkcd) port on loopback")
			testwl := Successful(sess.Run(ctx, "busybox:latest",
				run.WithName(testWorkloadName),
				run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
				run.WithPublishedPort("127.0.0.1:60789:12345/tcp"),
			))
			_ = model.PIDType(Successful(testwl.PID(ctx)))

			By("running a discovery")
			ctx, cancel := context.WithCancel(ctx)
			cizer := turtlefinder.New(func() context.Context { return ctx })
			defer cancel()
			defer cizer.Close()
			allnetns, lxknsdisco := discover.Discover(ctx, cizer, nil)
			Expect(allnetns).NotTo(BeEmpty())

			By("decorating with proxy ports")
			Decorate(ctx, allnetns, lxknsdisco.Processes, cizer.Engines(ctx))

			wl := lxknsdisco.Containers.FirstWithNameType(testWorkloadName, moby.Type)
			Expect(wl).NotTo(BeNil())
			netns := allnetns.ByProcess(lxknsdisco.Processes[wl.Engine.PID])
			Expect(netns).NotTo(BeNil())
			Expect(netns.ForwardedPortsv4).To(ContainElement(
				HaveField("ForwardedPortRange", And(
					HaveField("Protocol", "tcp"),
					HaveField("IP", net.ParseIP("127.0.0.1").To4()),
					HaveField("PortMin", uint16(60789)),
					HaveField("ForwardIP", net.IP(testwl.IP(ctx).AsSlice())),
					HaveField("ForwardPortMin", uint16(12345)),
				))))
		})

	})

})
