// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/siemens/ghostwire/v2/internal/discover"
	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/mobydig/messymoby"
	"github.com/siemens/turtlefinder"
	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	"github.com/onsi/gomega/types"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
)

var _ = Describe("docker neighborhood services digging", Ordered, func() {

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

	DescribeTable("discovers and streams neighborhood services",
		func(suckcess suckcess, startname string, validate func(<-chan JSONTextualRepresentation)) {
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
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			verdictCh, err := DigNeighborhoodServices(ctx, allnetns, start, 10, 10)
			Expect(err).To(suckcess)
			if validate != nil {
				validate(verdictCh)
				Eventually(verdictCh).WithTimeout(20 * time.Second).Should(BeClosed())
			}
		},
		Entry("nil container", toFail, "", nil),
		Entry("non-existing container", toFail, "4", nil), // https://xkcd.com/221/
		Entry("streaming digging and verification verdicts", toSucceed, "neighborhood-service-a-1",
			func(ch <-chan JSONTextualRepresentation) {
				Eventually(ch).WithTimeout(10 * time.Second).Should(
					Receive(MatchJSON(`{"fqdn":"service-c.","quality":"unverified"}`)))
				Eventually(ch).WithTimeout(10 * time.Second).Should(
					Receive(withoutAddress(MatchJSON(`{"fqdn":"service-c.","quality":"verifying"}`))))
				Eventually(ch).WithTimeout(10 * time.Second).Should(
					Receive(withoutAddress(MatchJSON(`{"fqdn":"service-c.","quality":"verified"}`))))
			}),
	)

})

// withoutAddress removes any "address" field that might be present before doing
// further matching.
func withoutAddress(m types.GomegaMatcher) types.GomegaMatcher {
	return WithTransform(func(j JSONTextualRepresentation) (string, error) {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal([]byte(j.String()), &raw); err != nil {
			return "", err
		}
		delete(raw, "address")
		b, _ := json.Marshal(raw)
		return string(b), nil
	}, m)
}
