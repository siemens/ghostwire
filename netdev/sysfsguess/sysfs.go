// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package sysfsguess

import (
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"
)

// SysfsOfNetns takes a network namespace (reference) and returns Mountineer to
// access a mount namespace view that has a matching sysfs instance, that is, a
// sysfs instance tagged with the same network namespace. In case of (serious)
// doubt, SysfsOfNetns returns a nil Mountineer.
//
// The caller is responsible to Close() the Mountineer when it is no longer
// needed, in order to not leak resources.
//
// The caller can use the Mountineer's Resolve("/") method to get the absolute
// path through a procfs “wormhole” to where the matching sysfs is then mounted
// below.
//
// For background information, please see the Linux kernel documentation on
// [Sysfs tagging]. The gist is that upon mounting a sysfs instance, it will tag
// itself with the network namespace the mount caller at that moment is attached
// to, so that whoever later looks at this sysfs instance, will only see exactly
// this network namespace-related device information. There doesn't seem to be
// any way to explicitly retrieve this tag from user space, though. So we need
// to rely on some sanity checks, for a insufficient definition of (in)sanity.
//
// [Sysfs tagging]: https://docs.kernel.org/networking/sysfs-tagging.html
func SysfsOfNetns(netns model.Namespace) *mountineer.Mountineer {
	if netns == nil {
		return nil
	}
	// Does the given network namespace even have any process attached? If not,
	// we bail out immediately, as we cannot determine a corresponding mount
	// namespace, because we need a (container) process to make the link.
	ealdorman := netns.Ealdorman()
	if ealdorman == nil {
		return nil
	}
	// Play safe and bail out in case the most senior process of our network
	// namespace isn't also the most senior process of the mount namespace the
	// senior is attached to: this might indicate that this might be a network
	// namespace without a mount namespace with the matching sysfs instance in
	// place.
	mntns := ealdorman.Namespaces[model.MountNS]
	if mntns.Ealdorman() != ealdorman {
		return nil
	}
	m, err := mountineer.NewWithMountNamespace(mntns, nil)
	if err != nil {
		return nil
	}
	return m
}
