// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package containerd

import (
	"context"
	"os"
	"time"

	"github.com/siemens/ghostwire/v2/test/nerdctl"
	detect "github.com/siemens/ghostwire/v2/turtlefinder/detector"

	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

// testWorkloadName specifies the name of a Docker container test workload, so
// we're sure that there is a well-defined container to be found.
const testWorkloadName = "gw-turtles-containerd-watch-workload"

var _ = Describe("containerd turtle watcher", func() {

	Context("with test workload", Ordered, func() {

		BeforeAll(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}
			nerdctl.SkipWithout()

			nerdctl.NerdctlIgnore("rm", "-f", testWorkloadName)
			nerdctl.Nerdctl(
				"run", "-d",
				"--name", testWorkloadName,
				"busybox", "/bin/sleep", "120s")

			goodfds := Filedescriptors()
			goodgos := Goroutines() // avoid other failed goroutine tests to spill over
			DeferCleanup(func() {
				nerdctl.NerdctlIgnore("rm", "-f", testWorkloadName)
				Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
					ShouldNot(HaveLeaked(goodgos))
				Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			})
		})

		It("registers correctly", func() {
			Expect(plugger.Group[detect.Detector]().Plugins()).To(
				ContainElement("containerd"))
		})

		It("tries unsuccessfully", NodeTimeout(30*time.Second), func(ctx context.Context) {
			d := &Detector{}
			Expect(d.NewWatcher(ctx, 0, []string{"/etc/rumpelpumpel"})).To(BeNil())
		})

		It("watches successfully", NodeTimeout(30*time.Second), func(ctx context.Context) {
			d := &Detector{}
			w := d.NewWatcher(ctx, 0, []string{
				"/etc/rumpelpumpel",
				"/var/run/containerd/containerd.sock.ttrpc",
				"/var/run/containerd/containerd.sock",
			})
			Expect(w).NotTo(BeNil())
			defer w.Close()
			go func() { // ...will be ended by cancelling the context
				_ = w.Watch(ctx)
			}()
			Eventually(w.Portfolio().Project("").ContainerNames,
				"5s", "250ms").Should(ContainElement(testWorkloadName))
		})

	})

})
