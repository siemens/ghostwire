// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/vishvananda/netlink"
)

// Vlan represents a VLAN network interface that in turn represents a specific
// IEEE 802.1Q VLAN ID.
type Vlan interface {
	Interface
	Vlan() *VlanAttrs // returns the VLAN attributes.
}

// VlanAttrs represents the attributes of a VLAN network interface.
type VlanAttrs struct {
	NifAttrs
	Master       Interface // master network interface
	VID          uint16    // VLAN ID 1..4094
	VlanProtocol netlink.VlanProtocol
}

var _ Vlan = (*VlanAttrs)(nil)
var _ resolver = (*VlanAttrs)(nil)    // Hmpf.
var _ initializer = (*VlanAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *VlanAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Vlan returns the VLAN attributes.
func (n *VlanAttrs) Vlan() *VlanAttrs { return n }

// Init initializes this VLAN Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *VlanAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	attrs := n.Link.(*netlink.Vlan)
	n.VID = uint16(attrs.VlanId)
	n.VlanProtocol = attrs.VlanProtocol
}

// ResolveRelations resolves relations to the master (underlay) network
// interface.
func (n *VlanAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	attrs := n.Link.(*netlink.Vlan)
	// Find out who our master interface is and then relate us to our master and
	// vice versa. Please note that RTNETLINK has the ugly behavior to drop the
	// IFLA_LINK attribute if the master network interface index happens to be
	// the same as our index. However, IFLA_LINK_NETNSID must be present in such
	// cases (as VLAN cannot be its own master) so we can detect and properly
	// handle this WTF.
	netnsid := NSID(attrs.NetNsID)
	if idx := attrs.ParentIndex; idx != 0 || netnsid != NSID_NONE {
		if idx == 0 {
			idx = n.Index // rtnetlink idio(t)syncrasy
		}
		netns := n.Netns
		if NSID(attrs.NetNsID) != NSID_NONE {
			netns = netns.related(netnsid)
		}
		if netns != nil {
			if master := netns.Nifs[idx]; master != nil {
				n.Master = master
				master.Nif().Slaves = append(master.Nif().Slaves, n.Interface())
			}
		} else {
			log.Warnf("unknown NSID %d in net:[%d]", netnsid, n.Netns.ID().Ino)
		}
	}
}

// Register our NifMaker for the "vlan" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &VlanAttrs{}
		}, plugger.WithPlugin("vlan"))
}
