// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/vishvananda/netlink"
)

// Vxlan represents a VXLAN overlay network interface with quite some gory
// configuration details and the relation to it "master" network interface.
type Vxlan interface {
	Interface
	Vxlan() *VxlanAttrs // returns the vxlan attributes.
}

// VxlanAttrs represents the attributes of a VXLAN network interface.
type VxlanAttrs struct {
	NifAttrs
	Master          Interface // master network interface
	VID             uint32    // VXLAN ID
	Groupv4         net.IP
	Groupv6         net.IP
	Sourcev4        net.IP
	Sourcev6        net.IP
	DestinationPort uint16
	SourcePortLow   uint16
	SourcePortHigh  uint16
	TTL             uint8
	TOS             uint8
	ArpProxy        bool
}

var _ Vxlan = (*VxlanAttrs)(nil)
var _ resolver = (*VxlanAttrs)(nil)    // Hmpf.
var _ initializer = (*VxlanAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *VxlanAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Vxlan returns the macvlan attributes.
func (n *VxlanAttrs) Vxlan() *VxlanAttrs { return n }

// Init initializes this VXLAN Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *VxlanAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	attrs := n.Link.(*netlink.Vxlan)
	n.VID = uint32(attrs.VxlanId)
	if groupv4 := attrs.Group.To4(); groupv4 != nil {
		n.Groupv4 = groupv4
	} else {
		n.Groupv6 = attrs.Group
	}
	if srcv4 := attrs.SrcAddr.To4(); srcv4 != nil {
		n.Sourcev4 = srcv4
	} else {
		n.Sourcev6 = attrs.SrcAddr
	}
	n.DestinationPort = uint16(attrs.Port)
	n.SourcePortLow = uint16(attrs.PortLow)
	n.SourcePortHigh = uint16(attrs.PortHigh)
	n.TTL = uint8(attrs.TTL)
	n.TOS = uint8(attrs.TOS)
	n.ArpProxy = attrs.Proxy
}

// ResolveRelations resolves relations to the master (underlay) network
// interface.
func (n *VxlanAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	attrs := n.Link.(*netlink.Vxlan)
	// Find out who our master interface is and then relate us to our master and
	// vice versa. Please note that RTNETLINK has the ugly behavior to drop the
	// IFLA_VXLAN_LINK attribute (=attrs.VtepDevIndex) if the master network
	// interface index happens to be the same as our index. However,
	// IFLA_LINK_NETNSID must be present in such cases (as VXLAN cannot be its
	// own master) so we can detect and properly handle this WTF.
	netnsid := NSID(attrs.NetNsID)
	if idx := attrs.VtepDevIndex; idx != 0 || netnsid != NSID_NONE {
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

// Register our NifMaker for the "vxlan" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &VxlanAttrs{}
		}, plugger.WithPlugin("vxlan"))
}
