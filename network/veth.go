// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/vishvananda/netlink"
)

// Veth represents an VETH network interface, and especially the peer-to-peer
// relationships between exactly two Veths. However, beware that since network
// discovery will never be an atomic operation, be prepared to trip upon an Veth
// without a peer (nil peer).
type Veth interface {
	Interface
	Veth() *VethAttrs // returns the veth attributes.
}

// VethAttrs represents the attributes of an VETH peer-to-peer network interface
// (one end of the pair).
type VethAttrs struct {
	NifAttrs
	Peer Interface // other end of the VETH "wire"
}

var _ Veth = (*VethAttrs)(nil)
var _ resolver = (*VethAttrs)(nil)    // Hmpf.
var _ initializer = (*VethAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *VethAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Veth returns the veth attributes.
func (n *VethAttrs) Veth() *VethAttrs { return n }

// Init initializes this VETH Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *VethAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	// nothing more to be done here; instead we will postpone any work until
	// resolving the peer relation.
}

// ResolveRelations resolves relations to the other peer network interface.
func (n *VethAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	attrs := n.Link.Attrs()
	if n.Peer == nil {
		// Find out who our peer interface is and then relate us to our peer and
		// vice versa. Please note that RTNETLINK has the ugly behavior to drop
		// the IFLA_LINK attribute (=attrs.ParentIndex) if the peer network
		// interface index happens to be the same as our index. However,
		// IFLA_LINK_NETNSID must be present in such cases (as VETH cannot be
		// its own peer) so we can detect and properly handle this WTF.
		netnsid := NSID(attrs.NetNsID)
		if idx := attrs.ParentIndex; idx != 0 || netnsid != NSID_NONE {
			if idx == 0 {
				idx = n.Index // rtnetlink idio(t)syncrasy
			}
			netns := n.Netns
			if netnsid != NSID_NONE {
				netns = netns.related(netnsid)
			}
			if netns != nil {
				if peer := netns.Nifs[idx]; peer != nil {
					n.Peer = peer
					self := n.Interface()
					if peerpeer := peer.(Veth).Veth().Peer; peerpeer == nil || peerpeer == self {
						peer.(*VethAttrs).Peer = self
					} else {
						// There's already a different peer set for our peer, so our
						// peer's peer isn't us.
						log.Warnf("VETH peer inconsistency for %s in net:[%d]: peer %s in net:[%d] has different peer %s in net:[%d] already set",
							n.Name, n.Netns.Namespace.ID().Ino,
							peer.Nif().Name, peer.Nif().Netns.Namespace.ID().Ino,
							peerpeer.Nif().Name, peerpeer.Nif().Netns.Namespace.ID().Ino)
					}
				}
			} else {
				log.Warnf("unknown NSID %d in net:[%d]", netnsid, n.Netns.ID().Ino)
			}
		}
	}
}

// Register our NifMaker for the "veth" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &VethAttrs{}
		}, plugger.WithPlugin("veth"))
}
