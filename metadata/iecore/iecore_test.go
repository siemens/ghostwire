// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package iecore

import (
	"context"
	"errors"
	"os"
	"sync"
	"time"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/turtlefinder"
	"github.com/siemens/ieddata"

	"github.com/cenkalti/backoff/v4"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
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

var _ = Describe("Industrial Edge core/runtime metadata", func() {

	var cizer containerizer.Containerizer

	BeforeEach(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over

		ctx, cancel := context.WithCancel(context.Background())
		cizer = turtlefinder.New(func() context.Context { return ctx })

		// Ensure that separate tests start a full metadata discovery over and
		// over again.
		once = &sync.Once{}

		DeferCleanup(func() {
			cancel()
			cizer.Close()
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
		})
	})

	It("doesn't crash without edge iot core container", NodeTimeout(30*time.Second), func(ctx context.Context) {
		dummy := &model.Container{}
		Expect(readEdgeCoreContainerOsrelease(dummy)).To(BeNil())

		r := gostwire.Discover(ctx, cizer, nil)

		dummy.Process = &model.Process{}
		dummy.Process.Namespaces[model.MountNS] = &brokenMntNs{}
		Expect(readEdgeCoreContainerOsrelease(dummy)).To(BeNil())

		Expect(findEdgeCoreContainer(r)).To(BeNil())
		Expect(gatherMetadata(r)).To(BeNil())
	})

	It("survives fake edge core container", NodeTimeout(30*time.Second), func(ctx context.Context) {
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = pool.RemoveContainerByName(fakeCoreWorkloadName) }()

		fakewl, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       fakeCoreWorkloadName,
			Cmd:        []string{"/bin/sleep", "120s"},
		})
		Expect(err).NotTo(HaveOccurred(), "container %s", fakeCoreWorkloadName)
		defer fakewl.Close()

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}, "5s", "250ms").ShouldNot(BeEmpty())

		cc := findEdgeCoreContainer(r)
		Expect(cc).NotTo(BeNil())
		Expect(readEdgeCoreContainerOsrelease(cc)).To(BeNil())
	})

	It("finds OS release information in edge core container", NodeTimeout(30*time.Second), func(ctx context.Context) {
		pool, err := dockertest.NewPool("")
		Expect(err).NotTo(HaveOccurred())
		defer func() { _ = pool.RemoveContainerByName(fakeCoreWorkloadName) }()

		fakewl, err := pool.RunWithOptions(&dockertest.RunOptions{
			Repository: "busybox",
			Tag:        "latest",
			Name:       fakeCoreWorkloadName,
			Cmd: []string{
				"/bin/sh", "-c",
				// Use broken VERSION_ID variable name on purpose to test "silent" fix...
				`echo '-e VERSION_ID="0.1.2.3"' > /etc/os-release-container && echo 'FOOBAR=baz' >> /etc/os-release-container && /bin/sleep 120s`,
			},
		})
		Expect(err).NotTo(HaveOccurred(), "container %", fakewl)
		defer fakewl.Close()

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

	It("gets IED meta data from platformbox db", NodeTimeout(30*time.Second), func(ctx context.Context) {
		pool, err := dockertest.NewPool("unix:///var/run/docker.sock")
		Expect(err).NotTo(HaveOccurred())

		_ = pool.Client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    ieddata.EdgeIotCoreContainerName,
			Force: true,
		})
		_, err = pool.BuildAndRunWithBuildOptions(
			&dockertest.BuildOptions{
				Dockerfile: "Dockerfile",
				ContextDir: "tests/fakeied",
			},
			&dockertest.RunOptions{
				Name: ieddata.EdgeIotCoreContainerName,
			})
		Expect(err).NotTo(HaveOccurred())
		defer func() {
			_ = pool.Client.RemoveContainer(docker.RemoveContainerOptions{
				ID:    ieddata.EdgeIotCoreContainerName,
				Force: true,
			})
		}()

		// wait for the container to become properly alive...
		err = backoff.Retry(func() error {
			c, err := pool.Client.InspectContainer(ieddata.EdgeIotCoreContainerName)
			if err != nil {
				return err
			}
			if !c.State.Running {
				return errors.New("edge core not running")
			}
			return nil
		}, backoff.WithMaxRetries(backoff.NewConstantBackOff(250*time.Millisecond), 5*(1000/250)))
		Expect(err).NotTo(HaveOccurred())

		var r gostwire.DiscoveryResult
		Eventually(func() model.Containers {
			r = gostwire.Discover(ctx, cizer, nil)
			return r.Lxkns.Containers
		}, "5s", "250ms").ShouldNot(BeEmpty())

		Expect(Metadata(r)).To(HaveKeyWithValue(
			"industrial-edge", And(
				HaveKeyWithValue("device-name", "iedx12345"),
				HaveKeyWithValue("developer-mode", "false"),
			)))

	})

})
