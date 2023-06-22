// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/siemens/ieddata"

	"github.com/cenkalti/backoff/v4"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/containerizer/whalefriend"
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("IE App icons", Ordered, func() {

	var pool *dockertest.Pool
	var fakecore *dockertest.Resource

	var cizer containerizer.Containerizer
	var cancel context.CancelFunc

	BeforeAll(func() {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		var err error
		pool, err = dockertest.NewPool("unix:///var/run/docker.sock")
		if err != nil {
			panic(err)
		}
		_ = pool.Client.RemoveContainer(docker.RemoveContainerOptions{
			ID:    ieddata.EdgeIotCoreContainerName,
			Force: true,
		})
		fakecore, err = pool.BuildAndRunWithBuildOptions(
			&dockertest.BuildOptions{
				Dockerfile: "Dockerfile",
				ContextDir: "tests/fakeied",
			},
			&dockertest.RunOptions{
				Name: ieddata.EdgeIotCoreContainerName,
			})
		Expect(err).NotTo(HaveOccurred())

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

		mobyw, err := moby.New("unix:///var/run/docker.sock", nil)
		Expect(err).NotTo(HaveOccurred())
		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		cizer = whalefriend.New(ctx, []watcher.Watcher{mobyw})
		Eventually(mobyw.Ready()).
			WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
			Should(BeClosed())
	})

	AfterAll(func() {
		if cancel != nil {
			cancel()
		}
		if cizer != nil {
			cizer.Close()
		}
		if pool != nil {
			_ = pool.Client.RemoveContainer(docker.RemoveContainerOptions{
				ID:    ieddata.EdgeIotCoreContainerName,
				Force: true,
			})
		}
	})

	It("locates the IE core runtime container", func() {
		allns := discover.Namespaces(
			discover.WithStandardDiscovery(),
			discover.WithContainerizer(cizer),
			discover.WithLabel(IEAppDiscoveryLabel, ""))
		Expect(allns.Containers).NotTo(BeEmpty())
		corePID := edgeCoreContainerPID([]*model.ContainerEngine{allns.Containers[0].Engine})
		Expect(corePID).To(Equal(model.PIDType(fakecore.Container.State.Pid)))
	})

	It("loads an App Icon", func() {
		db, err := ieddata.OpenInPID(platformboxdbName, model.PIDType(fakecore.Container.State.Pid))
		Expect(err).NotTo(HaveOccurred())
		apps, err := db.Apps()
		Expect(err).NotTo(HaveOccurred())
		iconData := loadAppIcon(&apps[0], model.PIDType(fakecore.Container.State.Pid))
		Expect(iconData).To(HavePrefix("data:image/png;base64,iVBORw0KGgoAAAAN"))
		Expect(iconData).To(HaveSuffix("Jggg=="))

		iconData2 := loadAppIcon(&apps[1], model.PIDType(fakecore.Container.State.Pid))
		Expect(iconData2).NotTo(BeEmpty())
		Expect(iconData2).NotTo(Equal(iconData))
	})

	Context("with App project container", func() {

		const cname = "app-B"
		const pname = "bbb"

		var sleepyB, sleepyZ *dockertest.Resource

		BeforeAll(func() {
			_ = pool.RemoveContainerByName(cname)
			var err error
			sleepyB, err = pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       cname,
				Cmd:        []string{"/bin/sleep", "60s"},
				Labels: map[string]string{
					composer.ComposerProjectLabel: pname,
					"com_mwp_conf_foo":            "bar", // just some label with com_mwp_conf prefix.
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				c, err := pool.Client.InspectContainer(sleepyB.Container.ID)
				Expect(err).NotTo(HaveOccurred(), "container %s", sleepyB.Container.Name[1:])
				return c.State.Running
			}, "5s", "100ms").Should(BeTrue(), "container %s", sleepyB.Container.Name[1:])

			_ = pool.RemoveContainerByName("zzz")
			sleepyZ, err = pool.RunWithOptions(&dockertest.RunOptions{
				Repository: "busybox",
				Tag:        "latest",
				Name:       "zzz",
				Cmd:        []string{"/bin/sleep", "60s"},
				Labels: map[string]string{
					composer.ComposerProjectLabel: "zzz",
					"com_mwp_conf_foo":            "bar", // just some label with com_mwp_conf prefix.
				},
			})
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				c, err := pool.Client.InspectContainer(sleepyZ.Container.ID)
				Expect(err).NotTo(HaveOccurred(), "container %s", sleepyZ.Container.Name[1:])
				return c.State.Running
			}, "5s", "100ms").Should(BeTrue(), "container %s", sleepyZ.Container.Name[1:])
		})

		AfterAll(func() {
			if sleepyB != nil {
				_ = pool.Purge(sleepyB)
			}
			if sleepyZ != nil {
				_ = pool.Purge(sleepyZ)
			}
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
