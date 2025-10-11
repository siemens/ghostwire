// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package passedthrough

import (
	"slices"
	"strings"

	"github.com/vishvananda/netlink"
)

// AltNameKVDelemiter defines the delimiter used in network interface alternate
// names to separate values from keys.
//
// Please note that the “ip-route” command places restrictions on what
// characters are allowed in alternative names, deviating from the “completely
// relaxed” Linux kernel that bascially accepts anything as long as it doesn't
// bust the 64k limit for alias names. Thus, for the sake of tools like
// “ip-route” don't use forward slashes “/” in alternative names.
const AltNameKVDelemiter = "="

// AltNameOriginalIfnamePrefix is the key (prefix) of an “alternative name” of
// the network interface that specifies the original network interface name,
// even when renamed by Docker's libnetwork while moved into a (networking)
// sandbox.
//
// This alternative name is set or regenerated (made sure to be attached) by
// cooperating Docker network drivers...
//   - when a Docker custom network using such a “passthrough”-like driver is created.
//   - when a cooperating driver (re)starts for currently existing custom
//     passthrough networks.
//   - when an endpoint (the only one allowed) is created on the custom
//     passthrough network.
//
// The alternative name is never intentially removed by cooperating drivers, so
// it is even present if a network interface is currently in its original place
// with its original name.
const AltNameOriginalIfnamePrefix = "siemens.iedge.orig-ifname" + AltNameKVDelemiter

// OriginalIfname returns the original name of an interface that currently might
// be roaming in some other network namespace, or "" if not applicable.
func OriginalIfname(nif netlink.Link) string {
	idx := slices.IndexFunc(nif.Attrs().AltNames, hasOriginalIfname)
	if idx < 0 {
		return ""
	}
	origIfname, _ := strings.CutPrefix(nif.Attrs().AltNames[idx], AltNameOriginalIfnamePrefix)
	return origIfname
}

// hasOriginalIfname returns true if the alternative name specifies the
// original ifname of a (host) network interface roaming around.
func hasOriginalIfname(altname string) bool {
	return strings.HasPrefix(altname, AltNameOriginalIfnamePrefix)
}
