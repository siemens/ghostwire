// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package engines

import (
	"context"
	"os"
	"time"

	"github.com/ory/dockertest"
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/turtlefinder"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/engineclient/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

const sleepyName = "sleepy-engines-test"

var _ = Describe("container engines metadata", func() {

	var cizer containerizer.Containerizer

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over

		ctx, cancel := context.WithCancel(context.Background())
		cizer = turtlefinder.New(func() context.Context { return ctx })

		DeferCleanup(func() {
			cancel()
			cizer.Close()
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("finds OS release information in edge core container", NodeTimeout(30*time.Second), func(ctx context.Context) {
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = pool.RemoveContainerByName(sleepyName) }()

		sleepy, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       sleepyName,
			Cmd:        []string{"/bin/sleep", "60s"},
		})
		Expect(err).NotTo(HaveOccurred(), "container %s", sleepyName)
		defer sleepy.Close()

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}, "5s", "250ms").ShouldNot(BeEmpty())

		Expect(Metadata(r)).To(HaveKeyWithValue(
			"container-engines", ContainElement(And(
				HaveField("ID", Not(BeEmpty())),
				HaveField("Type", moby.Type),
				HaveField("Version", Not(BeEmpty())),
			))))
	})

})
