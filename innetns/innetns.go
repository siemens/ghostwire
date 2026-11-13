// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package innetns

import (
	"errors"
	"fmt"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
)

// ErrInvalidReference is wrapped by errors from [Run] for invalid netns
// references, including stale ones.
var ErrInvalidReference = errors.New("invalid netns reference")

// Run the specified function while temporarily attached in the specified
// network namespace (and with the OS-level thread locked during the execution
// of fn).
func Run(netns model.Namespace, fn func() error) error {
	if netns == nil {
		return errors.New("nil namespace")
	}
	if netns.Type() != species.CLONE_NEWNET {
		return fmt.Errorf("invalid non-netns namespace of type %s", netns.Type().String())
	}
	ref := netns.Ref()
	if len(ref) == 0 {
		return errors.New("invalid empty netns reference")
	}
	// Do we have a single network namespace path or are there multiple paths,
	// where the last path references a bind-mounted network namespace in some
	// mount namespace?
	netnspath := ref[len(ref)-1]
	if len(ref) > 1 {
		mntneer, err := mountineer.New(ref[:len(ref)-1], nil)
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrInvalidReference, ref.String(), err)
		}
		// we need to keep the mount namespace open only until we've got the
		// netlink handle, that is, until the end of this method.
		defer mntneer.Close()
		netnspath, err = mntneer.Resolve(netnspath)
		if err != nil {
			return fmt.Errorf("%w %s: %w", ErrInvalidReference, ref.String(), err)
		}
	}
	// Open the referenced network namespace only once, so that we can check it
	// is still the expected one (fd-based references might have gone stale)
	// and then switch into exactly this one.
	netnsfd, closer, err := ops.NamespacePath(netnspath).NsFd()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrInvalidReference, ref.String(), err)
	}
	defer closer()
	targetnetnsid, err := ops.NamespaceFd(netnsfd).ID()
	if err != nil {
		return fmt.Errorf("%w %s: %w", ErrInvalidReference, ref.String(), err)
	}
	if targetnetnsid != netns.ID() {
		return fmt.Errorf("%w %s: stale, expected net:[%d], got net:[%d]",
			ErrInvalidReference, ref.String(), netns.ID().Ino, targetnetnsid.Ino)
	}
	// Shortcut route in case it's our own network namespace for those running
	// this without sufficient capabilities, so trying to switch into our own
	// current network namespace would otherwise fail. This of course begs the
	// question, why would we do to switch into the same network namespace
	// anyway? Because.
	if ourNetnsErr == nil && ourNetnsID == targetnetnsid {
		return fn()
	}
	// Normal route where we need to temporarily switch into the target network
	// namespace in order to open a file descriptor in that context.
	var openerErr error
	visitErr := ops.Visit(func() {
		openerErr = fn()
	}, ops.NamespaceFd(netnsfd))
	if visitErr != nil {
		return visitErr
	}
	return openerErr
}

var (
	ourNetnsID  species.NamespaceID // our own network namespace ID when starting.
	ourNetnsErr error               // there should not be any, except for a botched procfs.
)

// Fetch the ID of the network namespace we (=main Go routine) are running in,
// for once and all. All Go routines normally run attached to this network
// namespace, except for short and carefully locked "critical" sections.
func init() {
	ourNetnsID, ourNetnsErr = ops.NamespacePath("/proc/self/ns/net").ID()
}
