// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/vishvananda/netlink"
)

// Netkit represents a netkit network interface, that is, either one end of a
// BPF-programmable netkit pair, or a single-device netkit without any peer.
// Similar to VETH, be prepared to trip upon a paired netkit without a peer (nil
// peer).
type Netkit interface {
	Interface
	Netkit() *NetkitAttrs // returns the netkit attributes.
}

// NetkitAttrs represents the attributes of one end of a netkit pair.
type NetkitAttrs struct {
	NifAttrs
	Peer       Interface // other end of the netkit pair; always nil if Single.
	Mode       NetkitMode
	Policy     NetkitPolicy // default policy of this end when no BPF program is attached
	PeerPolicy NetkitPolicy // default policy of the peer end; meaningless if Single.
	Primary    bool         // primary end, usually in the host netns
	Single     bool         // single-device mode without any peer
}

// NetkitMode specifies whether a netkit pair operates in L2 or L3 mode.
type NetkitMode netlink.NetkitMode

// NetkitPolicy specifies the default forwarding policy of a netkit end.
type NetkitPolicy netlink.NetkitPolicy

var _ Netkit = (*NetkitAttrs)(nil)
var _ resolver = (*NetkitAttrs)(nil)    // Hmpf.
var _ initializer = (*NetkitAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *NetkitAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Netkit returns the netkit attributes.
func (n *NetkitAttrs) Netkit() *NetkitAttrs { return n }

// Init initializes this netkit Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *NetkitAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	nk := link.(*netlink.Netkit)
	n.Mode = NetkitMode(nk.Mode)
	n.Policy = NetkitPolicy(nk.Policy)
	n.PeerPolicy = NetkitPolicy(nk.PeerPolicy)
	n.Primary = nk.IsPrimary()
	n.Single = isSingleNetkit(n.Primary, nk.Attrs())
}

// isSingleNetkit returns true if a netkit is in single-device mode. Our
// netlink package doesn't decode the IFLA_NETKIT_PAIRING attribute, but a
// single-device netkit is always primary and RTNETLINK then reports neither a
// peer ifindex (IFLA_LINK is 0) nor a peer netns (IFLA_LINK_NETNSID).
func isSingleNetkit(primary bool, attrs *netlink.LinkAttrs) bool {
	return primary && attrs.ParentIndex == 0 && NSID(attrs.NetNsID) == NSID_NONE
}

// ResolveRelations resolves relations to the other peer network interface.
func (n *NetkitAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	if n.Peer != nil || n.Single {
		return
	}
	peer, ok := n.peerNif().(*NetkitAttrs)
	if !ok {
		return
	}
	n.Peer = peer
	self := n.Interface()
	if peer.Peer == nil || peer.Peer == self {
		peer.Peer = self
		return
	}
	warnPeerInconsistency(self, peer, peer.Peer)
}

// String returns a short text for this netkit mode.
func (m NetkitMode) String() string {
	return netkitModeIdentifiers[m]
}

var netkitModeIdentifiers = map[NetkitMode]string{
	NetkitMode(netlink.NETKIT_MODE_L2): "L2",
	NetkitMode(netlink.NETKIT_MODE_L3): "L3",
}

// String returns a short text for this netkit policy.
func (p NetkitPolicy) String() string {
	return netkitPolicyIdentifiers[p]
}

var netkitPolicyIdentifiers = map[NetkitPolicy]string{
	NetkitPolicy(netlink.NETKIT_POLICY_FORWARD):   "forward",
	NetkitPolicy(netlink.NETKIT_POLICY_BLACKHOLE): "blackhole",
}

// Register our NifMaker for the "netkit" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &NetkitAttrs{}
		}, plugger.WithPlugin("netkit"))
}
