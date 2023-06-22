// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"context"
	"errors"
	"os"
	"time"

	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"

	"github.com/siemens/ghostwire/v2/decorator/dockernet"
	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/ghostwire/v2/turtlefinder"
	"github.com/siemens/mobydig/messymoby"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	"github.com/onsi/gomega/types"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

// DcTestUpArgs specifies docker-compose CLI args for setting up the test
// harness.
var DcTestUpArgs = []string{
	"-f", "./test/neighborhood/docker-compose.yaml",
	"up",
	"-d",
	"--scale", "service-a=2",
}

// DcTestDnArgs specifies docker-compose CLI args for tearing down the test
// harness.
var DcTestDnArgs = []string{
	"-f", "./test/neighborhood/docker-compose.yaml",
	"down",
	"-t", "1",
}

var _ = Describe("docker network neighborhood services", Ordered, func() {

	var allnetns network.NetworkNamespaces
	var lxknsdisco *lxknsdiscover.Result

	BeforeAll(NodeTimeout(30*time.Second), func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		By("Cleaning up test containers and networks")
		messymoby.Cleanup(ctx)
		By("Tearing down any left-over test harness")
		messymoby.DockerCompose(ctx, DcTestDnArgs...)
		By("Bringing up the test harness")
		messymoby.DockerCompose(ctx, DcTestUpArgs...)

		DeferCleanup(NodeTimeout(30*time.Second), func(ctx context.Context) {
			By("Tearing down our test harness")
			messymoby.DockerCompose(ctx, DcTestDnArgs...)
			By("Cleaning up test containers and networks")
			messymoby.Cleanup(ctx)
		})

		By("running a discovery")
		cizer := turtlefinder.New(func() context.Context { return ctx })
		defer cizer.Close()
		Eventually(func() model.Containers {
			allnetns, lxknsdisco = discover.Discover(ctx, cizer, nil)
			return lxknsdisco.Containers
		}).Should(ContainElements(
			HaveField("Name", "neighborhood-service-a-1"),
			HaveField("Name", "neighborhood-service-a-2"),
			HaveField("Name", "neighborhood-service-b-1"),
			HaveField("Name", "neighborhood-service-c-1"),
		))
	})

	BeforeEach(func() {
		goodgos := Goroutines()
		goodfds := Filedescriptors()
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(3 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})
	})

	type suckcess types.GomegaMatcher
	var (
		toFail    suckcess = HaveOccurred()
		toSucceed suckcess = Succeed()
	)

	DescribeTable("discovers networks attached to a particular container",
		func(suckcess suckcess, startname string, netnames []string) {
			Expect(allnetns).NotTo(BeEmpty())
			var start *model.Container
			found, err := ContainElement(HaveField("Name", startname),
				&start).Match(lxknsdisco.Containers)
			if err == nil && !found {
				err = errors.New("no container match")
			}
			Expect(err).To(suckcess, "missing start container %q", startname)
			expectedNetnames := []types.GomegaMatcher{}
			for _, netname := range netnames {
				expectedNetnames = append(expectedNetnames,
					HaveField("Nif().Labels", HaveKeyWithValue(dockernet.GostwireNetworkNameKey, netname)))
			}
			attachees, err := attachedNetworks(allnetns, start)
			Expect(err).To(suckcess)
			Expect(attachees).To(HaveLen(len(netnames)))
			Expect(attachees).To(ConsistOf(expectedNetnames), "expected networks named %v", netnames)
		},
		Entry("nil container", toFail, "", []string{}),
		Entry("non-existing container", toFail, "4", []string{}), // https://xkcd.com/221/
		Entry("at service-a_1", toSucceed, "neighborhood-service-a-1", []string{
			"ghostwire-test-net-a",
		}),
		Entry("at service-b_1", toSucceed, "neighborhood-service-b-1", []string{
			"ghostwire-test-net-a", "ghostwire-test-net-b",
		}),
		Entry("at service-c_1", toSucceed, "neighborhood-service-c-1", []string{
			"ghostwire-test-net-a", "ghostwire-test-net-b",
		}),
	)

	DescribeTable("discovers neighborhood services",
		func(suckcess suckcess, startname string, services Services) {
			Expect(allnetns).NotTo(BeEmpty())
			// Okay, we're going to the meta level here (not that "metaverse"
			// oldety) by reusing Gomega matchers to filter out test
			// configuration information and at the same time cross-checking the
			// configuration.
			var start *model.Container
			found, err := ContainElement(HaveField("Name", startname),
				&start).Match(lxknsdisco.Containers)
			if err == nil && !found {
				err = errors.New("no container match")
			}
			Expect(err).To(suckcess, "missing start container %q", startname)
			expectedServices := []types.GomegaMatcher{}
			for _, service := range services {
				var expectedCntrNames []string
				for _, servant := range service.Servants {
					Expect(servant.Name).NotTo(BeEmpty())
					expectedCntrNames = append(expectedCntrNames, servant.Name)
				}
				Expect(expectedCntrNames).NotTo(BeEmpty())
				expectedServices = append(expectedServices, And(
					HaveField("Name", service.Name),
					HaveField("Servants", HaveEach(
						HaveField("Name", BeElementOf(expectedCntrNames)))),
					HaveField("NetworkLabels", ConsistOf(service.NetworkLabels)),
				))
			}
			actualServices, err := NeighborhoodServices(allnetns, start)
			Expect(err).To(suckcess)
			Expect(actualServices).To(ConsistOf(expectedServices))
		},
		Entry("nil container", toFail, "", nil),
		Entry("non-existing container", toFail, "4", nil), // https://xkcd.com/221/
		Entry("of service-x-1", toSucceed, "neighborhood-service-x-1", Services{
			{
				Name: "service-x",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-x-1"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-X",
				},
			},
		}),
		Entry("of service-a-1", toSucceed, "neighborhood-service-a-1", Services{
			{
				Name: "service-a",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-a-1"},
					&model.Container{Name: "neighborhood-service-a-2"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
				},
			},
			{
				Name: "service-b",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-b-1"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
				},
			},
			{
				Name: "service-c",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-c-1"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
				},
			},
		}),
		Entry("of service-b-1", toSucceed, "neighborhood-service-b-1", Services{
			{
				Name: "service-a",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-a-1"},
					&model.Container{Name: "neighborhood-service-a-2"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
				},
			},
			{
				Name: "service-b",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-b-1"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
					"ghostwire-test-net-b",
				},
			},
			{
				Name: "service-c",
				Servants: Containers{
					&model.Container{Name: "neighborhood-service-c-1"},
				},
				NetworkLabels: []string{
					"ghostwire-test-net-a",
					"ghostwire-test-net-b",
				},
			},
		}),
		Entry("of service-z-1", toSucceed, "neighborhood-service-z-1", Services{
			// nothing should be found when on the default network (docker0)
		}),
	)

})
