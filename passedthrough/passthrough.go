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

// AltNameDockerNetworkPrefix is the key (prefix) of an “alternative name” of
// the network interface that specifies the Docker network ID while the network
// interface is passed through. It is removed when the network interface isn't
// currently passed through.
const AltNameDockerNetworkPrefix = "siemens.iedge.docker-nw" + AltNameKVDelemiter

// OriginalIfname returns the original name of an interface that currently might
// be roaming in some other network namespace, or "" if not applicable.
func OriginalIfname(nif netlink.Link) string {
	origIfname, _ := AltnameValueOk(nif, AltNameOriginalIfnamePrefix)
	return origIfname
}

// DockernetID returns the Docker network ID of an interface while being
// passed-through; otherwise, it returns an empty string.
func DockernetID(nif netlink.Link) string {
	id, ok := AltnameValueOk(nif, AltNameDockerNetworkPrefix)
	if !ok {
		return ""
	}
	id, _, _ = strings.Cut(id, ".")
	return id
}

// Altname returns the alternate name of the passed network interface that
// matches the passed suffix, otherwise an empty string.
func Altname(nif netlink.Link, prefix string) string {
	altnames := nif.Attrs().AltNames
	if idx := slices.IndexFunc(altnames, func(altname string) bool {
		return strings.HasPrefix(altname, prefix)
	}); idx >= 0 {
		return altnames[idx]
	}
	return ""
}

// AltnameValueOk returns the value of the alternate name of the passed network
// interface and true, otherwise it returns an empty value and false.
func AltnameValueOk(nif netlink.Link, prefix string) (string, bool) {
	altname := Altname(nif, prefix)
	if altname == "" {
		return "", false
	}
	return strings.TrimPrefix(altname, prefix), true
}
