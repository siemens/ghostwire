// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"context"
	"time"

	"github.com/siemens/ghostwire/v2/util"

	"github.com/onsi/gomega/types"
	"github.com/ory/dockertest/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
)

// testEngineWorkloadName specifies the name of a Docker container test
// workload, so we're sure that there is a well-defined container to be found.
const testEngineWorkloadName = "gw-turtles-testengine-workload"

func HaveEngine(typ string, apiregex string) types.GomegaMatcher {
	return And(
		HaveField("Type", typ),
		HaveField("API", MatchRegexp(apiregex)))
}

var _ = Describe("container engine", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("tracks an engine", NodeTimeout(30*time.Second), func(ctx context.Context) {
		w, err := moby.New("", nil)
		Expect(err).NotTo(HaveOccurred())
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		e := NewEngine(ctx, w)
		Expect(e.ID).NotTo(BeZero())

		Consistently(e.IsAlive).Should(BeTrue())

		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		_ = pool.RemoveContainerByName(testEngineWorkloadName)
		_, err = pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       testEngineWorkloadName,
			Cmd:        []string{"/bin/sleep", "120s"},
		})
		Expect(err).NotTo(HaveOccurred(), "creating container %s", testEngineWorkloadName)
		defer func() { _ = pool.RemoveContainerByName(testEngineWorkloadName) }()

		// Give leeway for the container workload discovery to reflect the
		// correct situation even under heavy system load. And remember to pass
		// a function to Eventually, not a result ;)
		Eventually(func() []*model.Container {
			return e.Containers(ctx)
		}, "10s", "500ms").Should(
			ContainElement(util.HaveContainerNameID(testEngineWorkloadName)),
			"missing container %s", testEngineWorkloadName)

		cancel()
		Eventually(e.IsAlive).Should(BeFalse())
	})

})
