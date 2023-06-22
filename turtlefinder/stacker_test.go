// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"context"
	"os"
	"time"

	"github.com/siemens/ghostwire/v2/util"

	"os/exec"

	"github.com/onsi/gomega/gexec"
	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("turtles and elephants", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("get prefixed and stacked, not stucked", NodeTimeout(60*time.Second), func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("creating a turtle finder")
		watcherctx, watchercancel := context.WithCancel(ctx)
		tf := New(func() context.Context { return watcherctx })
		Expect(tf).NotTo(BeNil())
		defer watchercancel()
		defer tf.Close()

		By("stopping any old left-over containerized container engine")
		session, err := gexec.Start(exec.Command("./test/cind/teardown.sh"),
			GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(ctx, session).Should(gexec.Exit())

		By("discovering engines")
		// to get out of an import cycle we need to do the discover ourselves
		// ... but we can trim this down to only those steps we need here.
		discover := func() *lxknsdiscover.Result {
			return lxknsdiscover.Namespaces(
				lxknsdiscover.FromProcs(),
				lxknsdiscover.FromBindmounts(),
				lxknsdiscover.WithNamespaceTypes(
					species.CLONE_NEWNET|species.CLONE_NEWPID|species.CLONE_NEWNS|species.CLONE_NEWUTS),
				lxknsdiscover.WithHierarchy(),
				lxknsdiscover.WithContainerizer(tf),
				lxknsdiscover.WithPIDMapper(),
			)
		}

		_ = discover()
		Expect(tf.Engines()).To(ContainElements(
			HaveEngine(moby.Type, `^unix:///proc/\d+/root/run/docker.sock$`),
			HaveEngine(containerd.Type, `^unix:///proc/\d+/root/run/containerd/containerd.sock$`),
		))

		By("starting an additional container engine in a container")
		session, err = gexec.Start(exec.Command("./test/cind/setup.sh"),
			GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(ctx, session).Should(gexec.Exit(0))
		defer func() { // safety net
			session, err := gexec.Start(exec.Command("./test/cind/teardown.sh"),
				GinkgoWriter, GinkgoWriter)
			Expect(err).NotTo(HaveOccurred())
			Eventually(session).Within(20 * time.Second).ProbeEvery(100 * time.Millisecond).Should(gexec.Exit())
		}()

		By("turtle finder catching up")
		Eventually(ctx, func() int {
			_ = discover()
			return tf.EngineCount()
		}, "5s", "250ms").Should(Equal(3))

		containers := discover().Containers
		Expect(containers).To(ContainElement(
			SatisfyAll(
				util.HaveContainerNameID("cind-sleepy"),
				WithTransform(
					func(actual *model.Container) model.Labels { return actual.Labels },
					HaveKeyWithValue(GostwireContainerPrefixLabelName, "containerd-in-docker")))))

		By("stopping the containerized container engine")
		session, err = gexec.Start(exec.Command("./test/cind/teardown.sh"),
			GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(ctx, session).Should(gexec.Exit())

		By("running a final engine discovery")
		Eventually(ctx, func() int {
			_ = discover()
			return tf.EngineCount()
		}, "5s", "250ms").Should(Equal(2))
	})

})
