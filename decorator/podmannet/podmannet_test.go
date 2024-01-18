// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package podmannet

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/turtlefinder"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/success"
)

const (
	fedoraTag = "39"

	pindName      = "ghostwire-pind"
	pindImageName = "siemens/ghostwire-pind"

	nifDiscoveryTimeout = 5 * time.Second
	nifDiscoveryPolling = 250 * time.Millisecond

	goroutinesUnwindTimeout = 2 * time.Second
	goroutinesUnwindPolling = 250 * time.Millisecond
)

var _ = Describe("turtle finder", Ordered, Serial, func() {

	var pindCntr *dockertest.Resource

	BeforeAll(func() {
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
		pool := Successful(dockertest.NewPool("unix:///run/docker.sock"))
		_ = pool.RemoveContainerByName(pindName)
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
		Expect(pool.Client.BuildImage(docker.BuildImageOptions{
			Name:       pindImageName,
			ContextDir: "./_test/pind",
			Dockerfile: "Dockerfile",
			BuildArgs: []docker.BuildArg{
				{Name: "FEDORA_TAG", Value: fedoraTag},
			},
			OutputStream: io.Discard,
		})).To(Succeed())
		pindCntr = Successful(pool.RunWithOptions(
			&dockertest.RunOptions{
				Name:       pindName,
				Repository: pindImageName,
				Privileged: true,
				Mounts: []string{
					"/var/lib/containers", // well, this actually is an unnamed volume
				},
				Tty: false,
			}, func(hc *docker.HostConfig) {
				hc.Init = false
				hc.Tmpfs = map[string]string{
					"/tmp": "",
					"/run": "",
				}
				hc.Devices = []docker.Device{
					{PathOnHost: "/dev/fuse"},
				}
			}))

		By("waiting for systemd default target to be reached")
		// We need to wait for the container "contents" to have fully "booted",
		// because otherwise trying to pull a container image and run it gets
		// flaky. So we want to wait for systemd to reach its default target. To
		// slightly complicate things, we might be too fast so that the system
		// dbus inside the container isn't created yet and that would make
		// systemctl fail. We thus first wait for the system dbus socket to
		// appear and only then use systemctl for the container contents to
		// fully boot up...
		Expect(pindCntr.Exec([]string{
			"/bin/bash", "-c",
			"while [ ! -S \"/var/run/dbus/system_bus_socket\" ]; do sleep 1; done" +
				" && systemctl is-system-running --wait",
		}, dockertest.ExecOptions{
			StdOut: GinkgoWriter,
			StdErr: GinkgoWriter,
		})).Error().To(Succeed())

		By("creating a podman MACVLAN network")
		Expect(pindCntr.Exec([]string{
			"podman",
			"network", "create",
			"-d", "macvlan",
			"mcwielahm",
			"-o", "parent=eth0",
		}, dockertest.ExecOptions{
			StdOut: GinkgoWriter,
			StdErr: GinkgoWriter,
		})).Error().To(Succeed())

		By("running a canary container connected to the default 'podman' network")
		Expect(pindCntr.Exec([]string{
			"podman", "run", "-d", "--rm",
			"--name", "canary",
			"--net", "podman", /* WHAT?? otherwise doesn't connect the container??? */
			"busybox",
			"/bin/sh", "-c", "while true; do sleep 1; done",
		}, dockertest.ExecOptions{
			StdOut: GinkgoWriter,
			StdErr: GinkgoWriter,
		})).Error().To(Succeed())

		DeferCleanup(func() {
			By("removing the podman-in-Docker container")
			Expect(pool.Purge(pindCntr)).To(Succeed())
		})
	})

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(goroutinesUnwindTimeout).WithPolling(goroutinesUnwindPolling).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
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
			pindNetnsID := lxknsdisco.Processes[model.PIDType(pindCntr.Container.State.Pid)].
				Namespaces[model.NetNS].ID()
			return allnetns[pindNetnsID].Nifs
		}).Within(nifDiscoveryPolling).ProbeEvery(nifDiscoveryPolling).Should(ContainElements(
			HaveField("Nif()", And(
				HaveField("Name", "eth0"),
				HaveField("Alias", "mcwielahm"))),
			HaveField("Nif()", And(
				HaveField("Name", "podman0"),
				HaveField("Alias", "podman"))),
		))

	})

})
