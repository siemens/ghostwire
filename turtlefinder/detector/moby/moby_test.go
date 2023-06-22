// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package moby

import (
	"context"
	"os"
	"time"

	detect "github.com/siemens/ghostwire/v2/turtlefinder/detector"

	"github.com/ory/dockertest/v3"
	"github.com/thediveo/go-plugger/v3"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

// testWorkloadName specifies the name of a Docker container test workload, so
// we're sure that there is a well-defined container to be found.
const testWorkloadName = "gw-turtles-docker-watch-workload"

var _ = Describe("Docker turtle watcher", func() {

	var pool *dockertest.Pool

	BeforeEach(NodeTimeout(30*time.Second), func(_ context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		var err error
		pool, err = dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		_ = pool.RemoveContainerByName(testWorkloadName)
		_, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       testWorkloadName,
			Cmd:        []string{"/bin/sleep", "120s"},
		})
		Expect(err).NotTo(HaveOccurred(), "creating container %s", testWorkloadName)

		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(NodeTimeout(30*time.Second), func(_ context.Context) {
			_ = pool.RemoveContainerByName(testWorkloadName)
			Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	Context("with test workload", func() {

		It("registers correctly", func() {
			Expect(plugger.Group[detect.Detector]().Plugins()).To(
				ContainElement("dockerd"))
		})

		It("tries unsuccessfully", NodeTimeout(30*time.Second), func(ctx context.Context) {
			d := &Detector{}
			Expect(d.NewWatcher(ctx, 0, []string{"/etc/rumpelpumpel"})).To(BeNil())
		})

		It("watches successfully", NodeTimeout(30*time.Second), func(ctx context.Context) {
			d := &Detector{}
			w := d.NewWatcher(ctx, 0, []string{"/etc/rumpelpumpel", "/var/run/docker/metrics.sock", "/var/run/docker.sock"})
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
