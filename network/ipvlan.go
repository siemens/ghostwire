// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package network

import (
	"log/slog"

	"github.com/thediveo/go-plugger/v3"
	"github.com/vishvananda/netlink"
)

// Ipvlan represents an IPVLAN network interface with its configuration details
// and the relation to its "master" network interface.
type Ipvlan interface {
	Interface
	Ipvlan() *IpvlanAttrs // returns the ipvlan attributes.
}

// IpvlanAttrs represents the attributes of an IPVLAN network interface.
type IpvlanAttrs struct {
	NifAttrs
	Master Interface // master (hardware) network interface
	Mode   IpvlanMode
	Flag   IpvlanFlag
}

// IpvlanMode specifies the IPVLAN operating mode L2, L3, or L3S.
type IpvlanMode netlink.IPVlanMode

// IpvlanFlag specifies the IPVLAN switching behavior (bridge, private, VEPA).
type IpvlanFlag netlink.IPVlanFlag

var _ Ipvlan = (*IpvlanAttrs)(nil)
var _ resolver = (*IpvlanAttrs)(nil)    // Hmpf.
var _ initializer = (*IpvlanAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *IpvlanAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Ipvlan returns the ipvlan attributes.
func (n *IpvlanAttrs) Ipvlan() *IpvlanAttrs { return n }

// Init initializes this IPVLAN Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *IpvlanAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	ipvlan := link.(*netlink.IPVlan)
	n.Mode = IpvlanMode(ipvlan.Mode)
	n.Flag = IpvlanFlag(ipvlan.Flag)
}

// ResolveRelations resolves relations to the master network interface.
func (n *IpvlanAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
	attrs := n.Link.Attrs()
	// Same rtnetlink IFLA_LINK quirk as with MACVLANs: when the master's index
	// equals ours, only IFLA_LINK_NETNSID will be present.
	netnsid := NSID(attrs.NetNsID)
	idx := attrs.ParentIndex
	if idx == 0 && netnsid == NSID_NONE {
		return
	}
	if idx == 0 {
		idx = n.Index // rtnetlink idio(t)syncrasy
	}
	netns := n.Netns
	if netnsid != NSID_NONE {
		netns = netns.related(netnsid)
	}
	if netns == nil {
		slog.Warn("unknown NSID",
			slog.Uint64("nsid", uint64(netnsid)),
			slog.Uint64("netns", n.Netns.ID().Ino))
		return
	}
	if master := netns.Nifs[idx]; master != nil {
		n.Master = master
		master.Nif().Slaves = append(master.Nif().Slaves, n.Interface())
	}
}

// String returns a short text for this IPVLAN mode.
func (m IpvlanMode) String() string {
	return ipvlanModeIdentifiers[m]
}

var ipvlanModeIdentifiers = map[IpvlanMode]string{
	IpvlanMode(netlink.IPVLAN_MODE_L2):  "L2",
	IpvlanMode(netlink.IPVLAN_MODE_L3):  "L3",
	IpvlanMode(netlink.IPVLAN_MODE_L3S): "L3S",
}

// String returns a short text for this IPVLAN flag.
func (f IpvlanFlag) String() string {
	return ipvlanFlagIdentifiers[f]
}

var ipvlanFlagIdentifiers = map[IpvlanFlag]string{
	IpvlanFlag(netlink.IPVLAN_FLAG_BRIDGE):  "bridge",
	IpvlanFlag(netlink.IPVLAN_FLAG_PRIVATE): "private",
	IpvlanFlag(netlink.IPVLAN_FLAG_VEPA):    "VEPA",
}

// Register our NifMaker for the "ipvlan" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &IpvlanAttrs{}
		}, plugger.WithPlugin("ipvlan"))
}
