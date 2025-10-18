// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"context"
	"log/slog"
	"maps"
	"os"
	"slices"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/morbyd/v2"
	"github.com/thediveo/morbyd/v2/run"
	"github.com/thediveo/morbyd/v2/session"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

func discoverRedux() (NetworkNamespaces, *discover.Result) {
	discoverednetns := discover.Namespaces(
		discover.FromProcs(),
		discover.FromTasks(),
		discover.FromFds(),
		discover.WithNamespaceTypes(
			species.CLONE_NEWNET|species.CLONE_NEWPID|species.CLONE_NEWNS|species.CLONE_NEWUTS),
		discover.WithHierarchy(),
		discover.WithPIDMapper(),
	)
	allnetns := NewNetworkNamespaces(
		discoverednetns.Namespaces[model.NetNS],
		discoverednetns.Processes,
		discoverednetns.Containers)
	return allnetns, discoverednetns
}

var _ = Describe("bridge network interfaces", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(5 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	It("discovers bridge", func(ctx context.Context) {
		if os.Getuid() != 0 {
			Skip("needs root")
		}

		sess := Successful(morbyd.NewSession(ctx,
			session.WithAutoCleaning("test=ghostwire.network")))
		DeferCleanup(func(ctx context.Context) {
			sess.Close(ctx)
		})

		By("creating a test bridge network")
		netw := Successful(sess.CreateNetwork(ctx,
			"test-network-bridge"))

		By("creating a test workload and connecting it to the test network")
		cntr := Successful(sess.Run(ctx,
			"busybox",
			run.WithNetwork(netw.ID),
			run.WithCommand("/bin/sh", "-c", "while true; do sleep 1; done"),
		))
		Expect(cntr.PID(ctx)).Error().NotTo(HaveOccurred())

		By("running a discovery")
		allnetns, lxknsdisco := discoverRedux()
		Expect(allnetns).NotTo(BeEmpty())

		// Expect a bridge to be present, with a nif name derived from the
		// network ID.
		brname := "br-" + netw.ID[0:12]
		hostnetnsid := lxknsdisco.Processes[model.PIDType(os.Getpid())].Namespaces[model.NetNS].ID()
		hostnetns := allnetns[hostnetnsid]
		Expect(hostnetns).NotTo(BeNil())
		Expect(hostnetns.NamedNifs).To(HaveKey(brname),
			"known nifs: %s", slices.Collect(maps.Keys(hostnetns.NamedNifs)))
		nif := hostnetns.NamedNifs[brname]
		Expect(nif.Nif().Kind).To(Equal("bridge"))

		// Expect a single port: an veth. We leave more precise checks up to the
		// VETH unit tests.
		ports := nif.(Bridge).Bridge().Ports
		Expect(ports).To(HaveLen(1))
		port := ports[0]
		Expect(port.Nif().Kind).To(Equal("veth"))
	})

})
