// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/vishvananda/netlink"
)

// GostwireInternalBridgeKey specifies the label key indicating if a bridge
// belongs to a Docker network that has been configured as "internal". If
// present, it has an empty value.
const GostwireInternalBridgeKey = "gostwire/bridge/internal"

// Bridge represents a bridge network interface, and especially the bridge-port
// relationships.
type Bridge interface {
	Interface
	Bridge() *BridgeAttrs // returns the bridge attributes.
}

// BridgeAttrs represents the attributes of a bridge network interface.
type BridgeAttrs struct {
	NifAttrs
	Ports []Interface // "enslaved" network interfaces acting as bridge ports
}

var _ Bridge = (*BridgeAttrs)(nil)
var _ resolver = (*BridgeAttrs)(nil)    // Hmpf.
var _ initializer = (*BridgeAttrs)(nil) // Hmpf.

// Nif returns the common network interface attributes.
func (n *BridgeAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// Bridge returns the bridge attributes.
func (n *BridgeAttrs) Bridge() *BridgeAttrs { return n }

// Init initializes this Bridge Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *BridgeAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	// nothing more to be done here.
}

// ResolveRelations resolves relations to the enslaved "port" network
// interfaces. Please note that as only the ports of a bridge indicate their
// bridge, resolving these relations can only and mut be done in the generic
// resolution of the Nif base type.
func (n *BridgeAttrs) ResolveRelations(allns NetworkNamespaces) {
	n.NifAttrs.ResolveRelations(allns)
}

// Register our NifMaker for the "bridge" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &BridgeAttrs{}
		}, plugger.WithPlugin("bridge"))
}
