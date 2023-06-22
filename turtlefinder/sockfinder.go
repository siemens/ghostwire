// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"bufio"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// soAcceptCon is the state bit mask to identify listening unix domain sockets.
const soAcceptCon = 1 << 16

// soStream is the type of a connection-oriented/streaming unix domain socket.
const soStream = 1

// discoverAPISockets returns a list of (potential) listening API unix domain
// sockets for a specific process. The PID of the process must be valid in the
// current mount namespace and a correct proc filesystem must have been
// (re)mounted in this mount namespace, otherwise only an empty list will be
// returned. The easiest way is to do this with a PID valid in the initial PID
// namespace and with a correct proc in the current mount namespace that has
// full "host:pid" view.
func discoverAPISockets(pid model.PIDType) []string {
	var listeningSox = discoverListeningSox(pid)
	return matchProcSox(pid, listeningSox)
}

// socketPathsByIno maps the inode numbers of (unix domain) sockets to their
// corresponding path names. This map will not contain unix domain sockets from
// Linux' "abstract namespace" (see also:
// http://man7.org/linux/man-pages/man7/unix.7.html).
type socketPathsByIno map[uint64]string

// matchProcSox scans the open file descriptors ("fd") of the specified process
// for known listening sockets, passed in by listeningSox. The process is
// identified by its PID that needs to be usable with the proc file system
// mounted in the current mount namespace.
func matchProcSox(pid model.PIDType, listeningSox socketPathsByIno) (socketpaths []string) {
	fdbase := "/proc/" + strconv.FormatUint(uint64(pid), 10) + "/fd"
	fdentries, err := ioutil.ReadDir(fdbase)
	if err != nil {
		return
	}
	fdbase += "/"
	// Scan all directory entries below the process' /proc/[PID]/fd directory:
	// these represent the individual open file descriptors of this process.
	// They are links (rather: pseudo-symbolic links) to their corresponding
	// resources, such as file names, sockets, et cetera. For sockets, we can
	// only learn a socket's inode number, but neither its type, nor state.
	// Thus we need the sockets-by-inode dictionary to check whether a fd
	// references something of interest to us and the filesystem path it
	// points to (as usual, subject to the current mount namespace).
	for _, fdentry := range fdentries {
		fdlink, err := os.Readlink(fdbase + fdentry.Name())
		if err == nil && strings.HasPrefix(fdlink, "socket:[") {
			ino, err := strconv.ParseUint(fdlink[8:len(fdlink)-1], 10, 64)
			if err == nil {
				if soxpath, ok := listeningSox[ino]; ok {
					socketpaths = append(socketpaths, soxpath)
				}
			}
		}
	}
	return
}

// discoverListeningSox returns a map of (named) unix domain sockets in
// listening state in the mount namspace to which the specified process is
// attached to.. The map specifies for each listening unix domain socket both
// its inode number as the key and path as value.
func discoverListeningSox(pid model.PIDType) socketPathsByIno {
	sox := socketPathsByIno{}
	// Try to open the list of unix domain sockets currently present in the
	// system.
	//
	// Note 1: please note that this list is subject to mount namespace this
	// process is joined to. Some documentation and blog posts erroneously
	// indicate that this list is controlled by the current network namespace
	// (as "/proc/net/" might suggest), but without ever checking. However, when
	// thinking about it, this doesn't make any sense at all, as unix domain
	// sockets have names that are filesystem paths, so it does make sense that
	// the mount namespace gets control, but not the network namespace. /rant
	//
	// Note 2: lesser known, files in a different mount namespaces can be
	// directly accessed via the proc filesystem if there's a process attached
	// to the mount namespace. These wormholes are "/proc/[PID]/root/" and
	// predate Linux mount namespaces by quite some eons, dating back to
	// "chroot". The wormholes save us from needing to re-execute in order to
	// access a container engine API endpoint in a different mount namespace.
	// This improves performance, as we can keep in-process and even
	// aggressively parallelize talking to engines.
	//
	// It's "incontinentainers", after all.
	uf, err := os.Open("/proc/" + strconv.FormatUint(uint64(pid), 10) +
		"/net/unix")
	if err != nil {
		return nil
	}
	defer func() { _ = uf.Close() }() // otherwise gosec goes berserk over this, oh my!
	// Each line from /proc/[PID]/net/unix lists one socket with its state
	// ("flags"), type, etc. For precise field semantics, please see:
	// https://elixir.bootlin.com/linux/v5.0.3/source/net/unix/af_unix.c#L2831
	// -- this line of code generates a single line in /proc/net/unix.
	soxscan := bufio.NewScanner(uf)
	for soxscan.Scan() {
		fields := strings.Split(soxscan.Text(), " ")
		if len(fields) < 8 {
			continue
		}
		flags, err := strconv.ParseUint(fields[3], 16, 32)
		if err != nil {
			continue
		}
		soxtype, err := strconv.ParseUint(fields[4], 16, 16)
		if err != nil {
			continue
		}
		path := fields[7]
		// If this ain't ;) a unix socket in listening mode, then skip it.
		if soxtype != soStream || flags != soAcceptCon {
			continue
		}
		ino, err := strconv.ParseUint(fields[6], 10, 64)
		if err != nil {
			continue
		}
		sox[ino] = path // finally map the socket's inode number to its path.
	}
	return sox
}
