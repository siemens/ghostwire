// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/vishvananda/netlink"
)

// Macvlan represents a MACVLAN network interface with quite some gory
// configuration details and the relation to it "master" network interface (a
// "hardware") interface.
type Macvlan interface {
	Interface
	Macvlan() *MacvlanAttrs // returns the macvlan attributes.
}

// MacvlanAttrs represents the attributes of a MACVLAN network interface.
type MacvlanAttrs struct {
	NifAttrs
	Master Interface // master (hardware) network interface
	Mode   MacvlanMode
}

// MacvlanMode specifies the switching mode with respect to other MACVLANs and
// the rest of the world, excluding the master network interface though.
type MacvlanMode netlink.MacvlanMode

var _ Macvlan = (*MacvlanAttrs)(nil)
var _ resolver = (*MacvlanAttrs)(nil)    // Hmpf.
var _ initializer = (*MacvlanAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *MacvlanAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Macvlan returns the macvlan attributes.
func (n *MacvlanAttrs) Macvlan() *MacvlanAttrs { return n }

// Init initializes this MACVLAN Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *MacvlanAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	n.Mode = MacvlanMode(link.(*netlink.Macvlan).Mode)
}

// ResolveRelations resolves relations to the master network interface.
func (n *MacvlanAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	attrs := n.Link.Attrs()
	// Find out who our master interface is and then relate us to our master and
	// vice versa. Please note that RTNETLINK has the ugly behavior to drop the
	// IFLA_LINK attribute (=attrs.ParentIndex) if the master network interface
	// index happens to be the same as our index. However, IFLA_LINK_NETNSID
	// must be present in such cases (as MACVLAN cannot be its own master) so we
	// can detect and properly handle this WTF.
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

// String returns a short text for this MACVLAN (switching) mode.
func (m MacvlanMode) String() string {
	return macvlanModeIdentifiers[m]
}

var macvlanModeIdentifiers = map[MacvlanMode]string{
	MacvlanMode(netlink.MACVLAN_MODE_DEFAULT):  "default",
	MacvlanMode(netlink.MACVLAN_MODE_PRIVATE):  "private",
	MacvlanMode(netlink.MACVLAN_MODE_VEPA):     "VEPA",
	MacvlanMode(netlink.MACVLAN_MODE_BRIDGE):   "bridge",
	MacvlanMode(netlink.MACVLAN_MODE_PASSTHRU): "passthrough",
	MacvlanMode(netlink.MACVLAN_MODE_SOURCE):   "source",
}

// Register our NifMaker for the "macvlan" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &MacvlanAttrs{}
		}, plugger.WithPlugin("macvlan"))
}
