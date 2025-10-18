// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/siemens/ieddata"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"
	"github.com/thediveo/whalewatcher/v2/watcher"
	"github.com/thediveo/whalewatcher/v2/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

var _ = Describe("IE App icons", Ordered, func() {

	var cizer containerizer.Containerizer
	var fakecore *morbyd.Container

	BeforeAll(func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))

		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.decorator.ieappicon=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		imgid := Successful(sess.BuildImage(ctx, "tests/fakeied"))
		fakecore = Successful(sess.Run(ctx, imgid,
			run.WithName(ieddata.EdgeIotCoreContainerName)))
		_ = Successful(fakecore.PID(ctx))

		mobyw, err := moby.New("unix:///var/run/docker.sock", nil)
		Expect(err).NotTo(HaveOccurred())
		cizerctx, cancel := context.WithCancel(context.Background())
		DeferCleanup(cancel)

		cizer = whalefriend.New(cizerctx, []watcher.Watcher{mobyw})
		DeferCleanup(cizer.Close)
		Eventually(ctx, mobyw.Ready()).
			WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
			Should(BeClosed())
	})

	It("locates the IE core runtime container", func(ctx context.Context) {
		allns := discover.Namespaces(
			discover.WithStandardDiscovery(),
			discover.WithContainerizer(cizer),
			discover.WithLabel(IEAppDiscoveryLabel, ""))
		Expect(allns.Containers).NotTo(BeEmpty())
		corePID := edgeCoreContainerPID([]*model.ContainerEngine{allns.Containers[0].Engine})
		Expect(corePID).To(Equal(model.PIDType(Successful(fakecore.PID(ctx)))))
	})

	It("loads an App Icon", func(ctx context.Context) {
		pid := model.PIDType(Successful(fakecore.PID(ctx)))
		db, err := ieddata.OpenInPID(platformboxdbName, pid)
		Expect(err).NotTo(HaveOccurred())
		apps, err := db.Apps()
		Expect(err).NotTo(HaveOccurred())
		iconData := loadAppIcon(&apps[0], pid)
		Expect(iconData).To(HavePrefix("data:image/png;base64,iVBORw0KGgoAAAAN"))
		Expect(iconData).To(HaveSuffix("Jggg=="))

		iconData2 := loadAppIcon(&apps[1], pid)
		Expect(iconData2).NotTo(BeEmpty())
		Expect(iconData2).NotTo(Equal(iconData))
	})

	Context("with App project container", func() {

		const cname = "app-B"
		const pname = "bbb"

		BeforeAll(func(ctx context.Context) {
			sess := Successful(morbyd.NewSession(ctx,
				session.WithAutoCleaning("test.decorator.ieappicon.appprojcntr=")))
			DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

			sleepyB := Successful(sess.Run(ctx, "busybox:latest",
				run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
				run.WithName(cname),
				run.WithCombinedOutput(GinkgoWriter),
				run.WithLabel(composer.ComposerProjectLabel+"="+pname),
				run.WithLabel("com_mwp_conf_foo=bar"),
			))
			_ = Successful(sleepyB.PID(ctx))

			sleepyZzz := Successful(sess.Run(ctx, "busybox:latest",
				run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
				run.WithName("zzz"),
				run.WithCombinedOutput(GinkgoWriter),
				run.WithLabel(composer.ComposerProjectLabel+"=zzz"),
				run.WithLabel("com_mwp_conf_foo=bar"),
			))
			_ = Successful(sleepyZzz.PID(ctx))
		})

		It("loads new project App Icon", func() {
			allns := discover.Namespaces(
				discover.WithStandardDiscovery(),
				discover.WithContainerizer(cizer),
				discover.WithLabel(IEAppDiscoveryLabel, ""))
			Expect(allns.Containers).NotTo(BeZero())
			projects := []ieAppProject{
				{
					Name:         pname,
					ContainerIDs: []string{cname},
				},
			}
			loadProjectIcons([]*model.ContainerEngine{allns.Containers[0].Engine}, projects)
			Expect(projects[0].IconData).To(HaveSuffix("QAAAAASUVORK5CYII="))
		})

		It("decorates container with icon", func() {
			appIcons = ieAppProjects{} // hack: wipe out cache.
			allns := discover.Namespaces(
				discover.WithStandardDiscovery(),
				discover.WithContainerizer(cizer),
				discover.WithLabel(IEAppDiscoveryLabel, ""))
			Expect(allns.Containers).To(ContainElement(HaveValue(And(
				HaveField("Name", cname),
				HaveField("Labels", HaveKeyWithValue(IconLabel, Not(BeEmpty()))),
			))))
			Expect(allns.Containers).To(ContainElement(HaveValue(And(
				HaveField("Name", "zzz"),
				HaveField("Labels", Not(HaveKey(IconLabel))),
			))))
		})

	})

})
