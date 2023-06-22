// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"context"
	"os"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

var _ = Describe("turtle finder", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	// This is an ugly test with respect to goroutine leakage, as it runs a
	// discovery and then very quickly cancels the context, so watchers might
	// still be in their spin-up phase.
	It("finds docker and containerd", func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		watcherctx, watchercancel := context.WithCancel(context.Background())
		tf := New(func() context.Context { return watcherctx })
		Expect(tf).NotTo(BeNil())
		defer watchercancel()
		defer tf.Close()

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		lxdisco := discover.Namespaces(discover.WithFullDiscovery())
		_ = tf.Containers(ctx, lxdisco.Processes, lxdisco.PIDMap)
		engines := tf.Engines()
		Expect(engines).To(ContainElements(
			HaveEngine(moby.Type, `^unix:///proc/\d+/root/run/docker.sock$`),
			HaveEngine(containerd.Type, `^unix:///proc/\d+/root/run/containerd/containerd.sock$`),
		))
	})

})
