// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"log/slog"

	"github.com/thediveo/go-plugger/v3"
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
	if n.Peer != nil {
		return
	}
	peer := n.peerNif()
	if peer == nil {
		return
	}
	n.Peer = peer
	self := n.Interface()
	peerpeer := peer.(Veth).Veth().Peer
	if peerpeer == nil || peerpeer == self {
		peer.(*VethAttrs).Peer = self
		return
	}
	warnPeerInconsistency(n.Interface(), peer, peerpeer)
}

// peerNif returns the peer network interface of a VETH or netkit pair end, or
// nil if the peer cannot be found. Please note that RTNETLINK has the ugly
// behavior to drop the IFLA_LINK attribute (=attrs.ParentIndex) if the peer
// network interface index happens to be the same as our index. However,
// IFLA_LINK_NETNSID must be present in such cases (as a pair end cannot be its
// own peer) so we can detect and properly handle this WTF.
func (n *NifAttrs) peerNif() Interface {
	attrs := n.Link.Attrs()
	netnsid := NSID(attrs.NetNsID)
	idx := attrs.ParentIndex
	if idx == 0 && netnsid == NSID_NONE {
		return nil
	}
	if idx == 0 {
		idx = n.Index // rtnetlink idio(t)syncrasy
	}
	netns := n.Netns
	if netnsid != NSID_NONE {
		netns = netns.related(netnsid)
	}
	if netns == nil {
		slog.Warn("unknown NSID in netns",
			slog.Uint64("nsid", uint64(netnsid)),
			slog.Uint64("netns", n.Netns.ID().Ino))
		return nil
	}
	return netns.Nifs[idx]
}

// warnPeerInconsistency logs that the peer of nif already has a different peer
// set.
func warnPeerInconsistency(nif, peer, peerpeer Interface) {
	slog.Warn("peer inconsistency with different peer already set",
		slog.String("interface", nif.Nif().Name),
		slog.Uint64("netns", nif.Nif().Netns.Namespace.ID().Ino),
		slog.Group("peer",
			slog.String("interface", peer.Nif().Name),
			slog.Uint64("netns", peer.Nif().Netns.Namespace.ID().Ino),
		),
		slog.Group("existing-peer",
			slog.String("interface", peerpeer.Nif().Name),
			slog.Uint64("netns", peerpeer.Nif().Netns.Namespace.ID().Ino),
		))
}

// PeerOf returns the peer network interface of a VETH or netkit network
// interface, or nil if nif isn't a pair end or its peer is unknown.
func PeerOf(nif Interface) Interface {
	switch pairend := nif.(type) {
	case Veth:
		return pairend.Veth().Peer
	case Netkit:
		return pairend.Netkit().Peer
	}
	return nil
}

// Register our NifMaker for the "veth" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &VethAttrs{}
		}, plugger.WithPlugin("veth"))
}
