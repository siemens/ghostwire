// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"runtime"
	"testing"

	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/thediveo/success"
)

func init() {
	// avoid M0 ending up wedged as it was used during a throw-away namespace
	// switch, but as M0 is special it cannot be killed.
	runtime.LockOSThread()
}

func TestGostwireNetwork(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ghostwire/network package")
}

// ownNetworkNamespace returns a NetworkNamespace without any discovered details
// that wraps the lxkns network namespace we're currently attached to. Unlike
// mockedNetworkNamespace it can actually be entered and queried via netlink,
// yet without the cost of a full discovery.
func ownNetworkNamespace() *NetworkNamespace {
	GinkgoHelper()
	allns := discover.Namespaces(discover.FromProcs(), discover.WithNamespaceTypes(species.CLONE_NEWNET))
	ownnetns := allns.Namespaces[model.NetNS][Successful(ops.NamespacePath("/proc/self/ns/net").ID())]
	Expect(ownnetns).NotTo(BeNil())
	return &NetworkNamespace{
		Namespace: ownnetns,
		Nifs:      map[int]Interface{},
		NamedNifs: map[string]Interface{},
		peerNetns: map[NSID]*NetworkNamespace{},
	}
}

// mockedNetns is a fake lxkns network namespace with arbitrary ID and
// reference; calling any other method panics.
type mockedNetns struct {
	model.Namespace
	ref model.NamespaceRef
}

func (m *mockedNetns) ID() species.NamespaceID     { return species.NamespaceIDfromInode(666) }
func (m *mockedNetns) Type() species.NamespaceType { return species.CLONE_NEWNET }
func (m *mockedNetns) Ref() model.NamespaceRef     { return m.ref }

// mockedNetworkNamespace returns a NetworkNamespace for a fake lxkns network
// namespace with the specified reference.
func mockedNetworkNamespace(ref ...string) *NetworkNamespace {
	return &NetworkNamespace{
		Namespace: &mockedNetns{ref: ref},
		Nifs:      map[int]Interface{},
		NamedNifs: map[string]Interface{},
		peerNetns: map[NSID]*NetworkNamespace{},
	}
}
