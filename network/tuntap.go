// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/vishvananda/netlink"
)

type TunTap interface {
	Interface
	TunTap() *TunTapAttrs // returns the TUN/TAP attributes
}

type TunTapMode int

const (
	TunTapModeTun TunTapMode = TunTapMode(netlink.TUNTAP_MODE_TUN)
	TunTapModeTap TunTapMode = TunTapMode(netlink.TUNTAP_MODE_TAP)
)

// TunTapAttrs represents the attributes of a TUN/TAP network interface.
type TunTapAttrs struct {
	NifAttrs
	Mode       TunTapMode
	Processors []*model.Process
}

var _ TunTap = (*TunTapAttrs)(nil)
var _ initializer = (*TunTapAttrs)(nil)

// Nif returns the common network interface attributes.
func (n *TunTapAttrs) Nif() *NifAttrs { return &n.NifAttrs }

// TunTap returns the TAP/TUN attributes.
func (n *TunTapAttrs) TunTap() *TunTapAttrs { return n }

// Init initializes this TUN/TAP Nif from information specified in the
// NetworkNamespace and lots of netlink.Link information.
func (n *TunTapAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	n.NifAttrs.Init(nlh, netns, link)
	attrs := n.Link.(*netlink.Tuntap)
	n.Mode = TunTapMode(attrs.Mode)
}

// Register our NifMaker for the "tuntap" kind.
func init() {
	plugger.Group[NifMaker]().Register(
		func() Interface {
			return &TunTapAttrs{}
		}, plugger.WithPlugin("tuntap"))
}
