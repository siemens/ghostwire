// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package podmannet

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/siemens/turtlefinder/v2"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/build"
	"github.com/thediveo/morbyd/v2/exec"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"

	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	fedoraTag = "44"

	nifDiscoveryTimeout = 5 * time.Second
	nifDiscoveryPolling = 250 * time.Millisecond

	goroutinesUnwindTimeout = 5 * time.Second
	goroutinesUnwindPolling = 250 * time.Millisecond
)

var _ = Describe("turtle finder", Ordered, Serial, func() {

	var pindPID model.PIDType

	BeforeAll(func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(goroutinesUnwindTimeout).WithPolling(goroutinesUnwindPolling).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		By("spinning up a Docker container with a podman system service")
		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test.decorator.podmannet=")))
		DeferCleanup(func(ctx context.Context) { sess.Close(ctx) })

		// The necessary container start arguments loosely base on
		// https://www.redhat.com/sysadmin/podman-inside-container but had to be
		// heavily modified because they didn't work out as is, for whatever
		// reasons. This is now a mash-up of the args used to get the KinD
		// base-based images correctly working and some "spirit" of the before
		// mentioned RedHat blog post.
		//
		// Lesson learnt: podman in Docker is much more fragile than the podmen
		// want us to believe.
		//
		// docker run -it --rm --name pind
		//     --privileged \
		//     --cgroupns=private \
		//     --tmpfs /tmp \
		//     --tmpfs /run \
		//     --volume /var \
		//     --device=/dev/fuse \
		//   pind
		//
		// Please note that the initial build of the podman-in-Docker image is
		// really slow, as fedora installs lots of things.
		imgid := Successful(sess.BuildImage(ctx, "./_test/pind",
			build.WithOutput(GinkgoWriter),
			build.WithBuildArg("FEDORA_TAG="+fedoraTag),
		))
		pindCntr := Successful(sess.Run(ctx, imgid,
			run.WithCombinedOutput(GinkgoWriter),
			run.WithPrivileged(),
			run.WithVolume("/var/lib/containers"),
			run.WithTmpfs("/tmp"),
			run.WithTmpfs("/run"),
			run.WithDevice("/dev/fuse"),
		))
		pindPID = model.PIDType(Successful(pindCntr.PID(ctx)))

		By("waiting for systemd default target to be reached")
		// We need to wait for the container "contents" to have fully "booted",
		// because otherwise trying to pull a container image and run it gets
		// flaky. So we want to wait for systemd to reach its default target. To
		// slightly complicate things, we might be too fast so that the system
		// dbus inside the container isn't created yet and that would make
		// systemctl fail. We thus first wait for the system dbus socket to
		// appear and only then use systemctl for the container contents to
		// fully boot up...
		cmd := Successful(pindCntr.Exec(ctx,
			exec.Command("/bin/bash", "-c",
				"while [ ! -S \"/var/run/dbus/system_bus_socket\" ]; do sleep 1; done && systemctl is-system-running --wait"),
			exec.WithCombinedOutput(GinkgoWriter)))
		waitctx, waitcancel := context.WithTimeout(ctx, 10*time.Second)
		defer waitcancel()
		Expect(Successful(cmd.Wait(waitctx))).To(BeZero())

		By("creating a podman MACVLAN network")
		cmd = Successful(pindCntr.Exec(ctx,
			exec.Command(
				"podman",
				"network", "create",
				"-d", "macvlan",
				"mcwielahm",
				"-o", "parent=eth0"),
			exec.WithCombinedOutput(GinkgoWriter),
		))
		Expect(Successful(cmd.Wait(waitctx))).To(BeZero())

		By("running a canary container connected to the default 'podman' network")
		backgoundcmd := Successful(pindCntr.Exec(ctx,
			exec.Command("podman", "run", "-d", "--rm",
				"--name", "canary",
				"--net", "podman", /* WHAT?? otherwise doesn't connect the container??? */
				"busybox",
				"/bin/sh", "-c", "while true; do sleep 1; done"),
			exec.WithCombinedOutput(GinkgoWriter),
		))
		_ = Successful(backgoundcmd.PID(ctx))
	})

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(goroutinesUnwindTimeout).WithPolling(goroutinesUnwindPolling).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	It("decorates podman-managed network interfaces", func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a turtlefinder")
		ctx, cancel := context.WithCancel(ctx)
		cizer := turtlefinder.New(func() context.Context { return ctx })
		defer cancel()
		defer cizer.Close()

		By("running a full Ghostwire discovery that should pick up the podman networks")
		Eventually(ctx, func() map[int]network.Interface {
			allnetns, lxknsdisco := discover.Discover(ctx, cizer, nil)
			pindNetnsID := lxknsdisco.Processes[pindPID].
				Namespaces[model.NetNS].ID()
			return allnetns[pindNetnsID].Nifs
		}).Within(nifDiscoveryTimeout).ProbeEvery(nifDiscoveryPolling).Should(ContainElements(
			HaveField("Nif()", And(
				HaveField("Name", "eth0"),
				HaveField("Alias", "mcwielahm"))),
			HaveField("Nif()", And(
				HaveField("Name", "podman0"),
				HaveField("Alias", "podman"))),
		))

	})

})
