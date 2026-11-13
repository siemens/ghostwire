// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"errors"
	"fmt"
	"log/slog"
	"maps"
	"net"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/thediveo/nonstd/xiter"
	"github.com/thediveo/notwork/dummy"
	"github.com/thediveo/notwork/nlhandle"
	"github.com/thediveo/notwork/veth"
	"github.com/thediveo/spacetest/netns"
	"github.com/vishvananda/netlink"

	"github.com/siemens/ghostwire/v2/innetns"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gleak"
	. "github.com/thediveo/fdooze"
	. "github.com/thediveo/namspill"
	. "github.com/thediveo/success"
)

var _ = Describe("network namespace", func() {

	BeforeEach(func() {
		goodfds := Filedescriptors()
		goodgos := Goroutines() // avoid other failed goroutine tests to spill over
		DeferCleanup(func() {
			Eventually(Goroutines).WithTimeout(2 * time.Second).WithPolling(250 * time.Millisecond).
				ShouldNot(HaveLeaked(goodgos))
			Expect(Filedescriptors()).NotTo(HaveLeakedFds(goodfds))
			Expect(Tasks()).To(BeUniformlyNamespaced())
		})

		DeferCleanup(slog.SetDefault, slog.Default())
		slog.SetDefault(slog.New(slog.NewTextHandler(GinkgoWriter, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})))
	})

	Context("with an almost lonely network namespace", func() {

		var allnetns NetworkNamespaces
		var testNetns *NetworkNamespace
		var gromit netlink.Link

		BeforeEach(func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			By("creating a new transient network namespace and a VETH pair from here to there")
			tmpNetns := netns.NewTransient()
			_, gromit = veth.NewTransient(veth.WithPeerNamespace(tmpNetns))

			allnetns, _ = discoverRedux()
			Expect(allnetns).To(HaveKey(species.NamespaceIDfromInode(netns.Ino(tmpNetns))))

			testNetns = allnetns[species.NamespaceIDfromInode(netns.Ino(tmpNetns))]
			Expect(testNetns).NotTo(BeNil())
		})

		It("found the related NetworkNamespace via their discovered NSIDs", func() {
			initnetnsid, err := ops.NamespacePath("/proc/self/ns/net").ID()
			Expect(err).NotTo(HaveOccurred())
			initialNetns := allnetns[initnetnsid]
			Expect(initialNetns).NotTo(BeNil())

			Expect(initialNetns.peerNetns).To(ContainElement(testNetns))
			Expect(testNetns.peerNetns).To(ContainElement(initialNetns))
		})

		It("lists nifs in new network namespace", func() {
			Expect(testNetns.NifList()).To(ConsistOf(
				HaveInterfaceName("lo"),
				HaveInterfaceName(gromit.Attrs().Name)))
		})

		It("names a tenant-less network namespace after its reference", func() {
			Expect(testNetns.Tenants).To(BeEmpty())
			Expect(testNetns.DisplayName()).To(Equal(testNetns.Ref().String()))
			Expect(orderingNetnsName(testNetns)).To(Equal(testNetns.Ref().String()))
		})

		When("sorting", func() {

			It("sorts the initial netns first", func() {
				initnetnsid, err := ops.NamespacePath("/proc/1/ns/net").ID()
				Expect(err).NotTo(HaveOccurred())
				initialNetns := allnetns[initnetnsid]
				Expect(initialNetns).NotTo(BeNil())

				anotherNetns := Allright(
					xiter.FirstOk(
						xiter.Filter(maps.Values(allnetns),
							func(n *NetworkNamespace) bool { return n != initialNetns })))
				Expect(anotherNetns).NotTo(BeNil())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, anotherNetns})(0, 1)).To(BeTrue())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{anotherNetns, initialNetns})(0, 1)).To(BeFalse())
				Expect(orderNetworkNamespaces([]*NetworkNamespace{initialNetns, initialNetns})(0, 1)).To(BeFalse())
			})

		})

	})

	Context("looking up and listing", func() {

		var allnetns NetworkNamespaces
		var ourproc *model.Process
		var ournetns *NetworkNamespace

		BeforeEach(func() {
			var disco *discover.Result
			allnetns, disco = discoverRedux()
			ourproc = disco.Processes[model.PIDType(os.Getpid())]
			Expect(ourproc).NotTo(BeNil())
			ournetns = allnetns[Successful(ops.NamespacePath("/proc/self/ns/net").ID())]
			Expect(ournetns).NotTo(BeNil())
		})

		It("looks up the network namespace of a process", func() {
			Expect(allnetns.ByProcess(ourproc)).To(BeIdenticalTo(ournetns))
			Expect(allnetns.ByProcess(&model.Process{})).To(BeNil())
		})

		It("looks up the network namespace of a container", func() {
			Expect(allnetns.ByContainer(&model.Container{Process: ourproc})).To(BeIdenticalTo(ournetns))
		})

		It("returns a sorted list", func() {
			sorted := allnetns.Sorted()
			Expect(sorted).To(ConsistOf(slices.Collect(maps.Values(allnetns))))
			for idx := 1; idx < len(sorted); idx++ {
				Expect(orderNetworkNamespaces(sorted)(idx, idx-1)).To(BeFalse())
			}
		})

		It("returns the display names", func() {
			Expect(ournetns.DisplayName()).NotTo(BeEmpty())
			Expect(allnetns.String()).To(ContainSubstring(ournetns.DisplayName()))
			Expect(strings.Count(allnetns.String(), " // ")).To(Equal(len(allnetns) - 1))
		})

		It("relates open ports to their processes", func() {
			listener := Successful(net.Listen("tcp4", "127.0.0.1:0"))
			DeferCleanup(listener.Close)
			port := uint16(AssignableTo[*net.TCPAddr](listener.Addr()).Port)

			allnetns, _ := discoverRedux()
			ournetns := allnetns[Successful(ops.NamespacePath("/proc/self/ns/net").ID())]
			Expect(ournetns.Portsv4).To(ContainElement(And(
				HaveField("LocalPort", port),
				HaveField("SimplifiedState", Listening),
				HaveField("PIDs", ContainElement(model.PIDType(os.Getpid()))))))
		})

		It("relates open ports to their processes without being passed socket processes", func() {
			listener := Successful(net.Listen("tcp4", "127.0.0.1:0"))
			DeferCleanup(listener.Close)
			port := uint16(AssignableTo[*net.TCPAddr](listener.Addr()).Port)

			disco := discover.Namespaces(
				discover.FromProcs(),
				discover.WithNamespaceTypes(
					species.CLONE_NEWNET|species.CLONE_NEWPID|species.CLONE_NEWNS|species.CLONE_NEWUTS))
			allnetns := NewNetworkNamespaces(
				disco.Namespaces[model.NetNS], disco.Processes, disco.Containers)
			ournetns := allnetns[Successful(ops.NamespacePath("/proc/self/ns/net").ID())]
			Expect(ournetns.Portsv4).To(ContainElement(And(
				HaveField("LocalPort", port),
				HaveField("SimplifiedState", Listening),
				HaveField("PIDs", ContainElement(model.PIDType(os.Getpid()))))))
		})

		It("discovers ports in network namespaces without processes", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			netnsfd := netns.NewTransient()
			var listener net.Listener
			netns.Execute(netnsfd, func() {
				listener = Successful(net.Listen("tcp4", "0.0.0.0:0"))
			})
			DeferCleanup(listener.Close)
			port := uint16(AssignableTo[*net.TCPAddr](listener.Addr()).Port)

			allnetns, _ := discoverRedux()
			portnetns := allnetns[species.NamespaceIDfromInode(netns.Ino(netnsfd))]
			Expect(portnetns).NotTo(BeNil())
			Expect(portnetns.Ealdorman()).To(BeNil())
			Expect(portnetns.Portsv4).To(ContainElement(And(
				HaveField("LocalPort", port),
				HaveField("SimplifiedState", Listening),
				HaveField("PIDs", ContainElement(model.PIDType(os.Getpid()))))))
		})

		It("discovers network interface aliases", func() {
			if os.Getuid() != 0 {
				Skip("needs root")
			}

			netnsfd := netns.NewTransient()
			dmy := dummy.NewTransient(dummy.InNamespace(netnsfd))
			Expect(nlhandle.New(netnsfd).LinkSetAlias(dmy, "lost-in-space")).To(Succeed())

			allnetns, _ := discoverRedux()
			dmynetns := allnetns[species.NamespaceIDfromInode(netns.Ino(netnsfd))]
			Expect(dmynetns).NotTo(BeNil())
			Expect(dmynetns.NamedNifs[dmy.Attrs().Name].Nif().Alias).To(Equal("lost-in-space"))
		})

	})

	Context("failing to work in network namespaces", func() {

		It("logs invalid references only as warnings", func() {
			Expect(netnsErrLevel(fmt.Errorf("%w: foo", innetns.ErrInvalidReference))).To(Equal(slog.LevelWarn))
			Expect(netnsErrLevel(errors.New("foo"))).To(Equal(slog.LevelError))
		})

		It("runs in the network namespace using the deprecated API", func() {
			canary := errors.New("canary")
			Expect(ownNetworkNamespace().OpenInNetworkNamespace(func() error { return canary })).
				To(BeIdenticalTo(canary))
		})

		It("doesn't discover inside an inaccessible network namespace", func() {
			Expect(NewNetworkNamespace(&mockedNetns{ref: model.NamespaceRef{"/nonexisting"}}, nil)).To(BeNil())

			netns := mockedNetworkNamespace("/nonexisting")
			Expect(netns.OpenNetlink()).Error().To(MatchError(innetns.ErrInvalidReference))
			fd, err := netns.OpenEthtool()
			Expect(err).To(MatchError(innetns.ErrInvalidReference))
			Expect(fd).To(Equal(-1))

			netns.discoverTransportPorts(nil, nil)
			Expect(netns.Portsv4).To(BeEmpty())
			netns.discoverNSIDs(NetworkNamespaces{})
			Expect(netns.peerNetns).To(BeEmpty())
		})

		It("skips peer network namespaces it can't determine NSIDs for", func() {
			ownnetns := ownNetworkNamespace()
			allnetns := NetworkNamespaces{ownnetns.ID(): ownnetns}
			for idx, ref := range []model.NamespaceRef{
				{},
				{"/nonexisting", "/foo"},
				{"/proc/self/ns/mnt", "/nonexisting"},
				{"/nonexisting"},
				{"/dev/null"}, // not a network namespace, so no NSID.
			} {
				allnetns[species.NamespaceIDfromInode(uint64(1000+idx))] = mockedNetworkNamespace(ref...)
			}
			ownnetns.discoverNSIDs(allnetns)
			Expect(ownnetns.peerNetns).To(BeEmpty())
		})

	})

})
