// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package iecore

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/siemens/ieddata"
	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/build"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"

	gostwire "github.com/siemens/ghostwire/v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const fakeCoreWorkloadName = "edge-iot-core"

// For full coverage we use our broken mount namespace that returns a wholly
// unusable filesystem reference. The good people of Ankh Morpork would surely
// like it.
type brokenMntNs struct{}

func (m brokenMntNs) ID() species.NamespaceID         { return species.NoneID }
func (m brokenMntNs) Type() species.NamespaceType     { return species.CLONE_NEWNS }
func (m brokenMntNs) Owner() model.Ownership          { return nil }
func (m brokenMntNs) Ref() model.NamespaceRef         { return []string{"/proc/foobar"} }
func (m brokenMntNs) Leaders() []*model.Process       { return nil }
func (m brokenMntNs) LeaderPIDs() []model.PIDType     { return nil }
func (m brokenMntNs) Ealdorman() *model.Process       { return nil }
func (m brokenMntNs) String() string                  { return "brokenMntNs" }
func (m brokenMntNs) LooseThreadIDs() []model.PIDType { return nil }
func (m brokenMntNs) LooseThreads() []*model.Task     { return nil }

var _ = Describe("Industrial Edge core/runtime metadata", Ordered, func() {

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

		// Ensure that separate tests start a full metadata discovery over and
		// over again.
		once = &sync.Once{}

		DeferCleanup(func() {
			cancel()
			cizer.Close()
			Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("doesn't crash without edge iot core container", func(ctx context.Context) {
		dummy := &model.Container{}
		Expect(readEdgeCoreContainerOsrelease(dummy)).To(BeNil())

		r := gostwire.Discover(ctx, cizer, nil)

		dummy.Process = &model.Process{}
		dummy.Process.Namespaces[model.MountNS] = &brokenMntNs{}
		Expect(readEdgeCoreContainerOsrelease(dummy)).To(BeNil())

		Expect(findEdgeCoreContainer(r)).To(BeNil())
		Expect(gatherMetadata(r)).To(BeNil())
	})

	It("survives fake edge core container", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.metadata.engines=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		fakewl := Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(fakeCoreWorkloadName),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
			run.WithCombinedOutput(GinkgoWriter),
		))
		_ = Successful(fakewl.PID(ctx))

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			ShouldNot(BeEmpty())

		cc := findEdgeCoreContainer(r)
		Expect(cc).NotTo(BeNil())
		Expect(readEdgeCoreContainerOsrelease(cc)).To(BeNil())
	})

	It("finds OS release information in edge core container", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.metadata.engines=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		fakewl := Successful(sess.Run(ctx, "busybox:latest",
			run.WithName(fakeCoreWorkloadName),
			run.WithCommand("/bin/sh", "-c",
				// Use broken VERSION_ID variable name on purpose to test "silent" fix...
				`echo '-e VERSION_ID="0.1.2.3"' > /etc/os-release-container && echo 'FOOBAR=baz' >> /etc/os-release-container && while true; do sleep 1; done`,
			),
			run.WithCombinedOutput(GinkgoWriter),
		))
		_ = Successful(fakewl.PID(ctx))

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}, "5s", "250ms").ShouldNot(BeEmpty())

		cc := findEdgeCoreContainer(r)
		Expect(cc).NotTo(BeNil())
		vars := readEdgeCoreContainerOsrelease(cc)
		Expect(vars).To(And(
			HaveKeyWithValue("VERSION_ID", "0.1.2.3"),
			HaveKeyWithValue("FOOBAR", "baz"),
		))

		Expect(Metadata(r)).To(HaveKeyWithValue(
			"industrial-edge", HaveKeyWithValue(
				"semversion", "0.1.2.3"),
		))
	})

	It("gets IED meta data from platformbox db", func(ctx context.Context) {
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.metadata.engines=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		imgid := Successful(sess.BuildImage(ctx, "tests/fakeied",
			build.WithOutput(GinkgoWriter),
		))
		fakewl := Successful(sess.Run(ctx, imgid,
			run.WithName(ieddata.EdgeIotCoreContainerName),
			run.WithCombinedOutput(GinkgoWriter),
		))
		_ = Successful(fakewl.PID(ctx))

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}).Within(5 * time.Second).ProbeEvery(250 * time.Millisecond).
			ShouldNot(BeEmpty())

		Expect(Metadata(r)).To(HaveKeyWithValue(
			"industrial-edge", And(
				HaveKeyWithValue("device-name", "iedx12345"),
				HaveKeyWithValue("developer-mode", "false"),
			)))

	})

})
