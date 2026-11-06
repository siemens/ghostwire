// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"bufio"
	"log/slog"
	"os"
	"strconv"
	"strings"

	"github.com/thediveo/ioctl"
	"github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// resolveTapTunProcessors checks for the presence of TAP and/or TUN network
// interfaces and then resolves their serving processes ("processors"). If no
// processors have been collected during the lxkns discovery, it runs an fd scan
// of its own.
func resolveTapTunProcessors(netspaces NetworkNamespaces, allprocs model.ProcessTable, tuntaps *TapTunProcessors) {
	if !hasTapTun(netspaces) {
		return
	}
	if tuntaps == nil {
		tuntaps = &TapTunProcessors{}
		_ = discover.Namespaces(discover.WithFdCallback(tuntaps.Visit))
	}
	for _, processor := range tuntaps.processors {
		proc := allprocs[processor.PID]
		if proc == nil {
			continue
		}
		netns := netspaces[processor.NetnsID]
		if netns == nil {
			slog.Warn("TAP/TUN serving process related to unknown netns",
				slog.String("name", proc.Name),
				slog.Int("pid", int(proc.PID)),
				slog.Uint64("netns", processor.NetnsID.Ino))
			continue
		}
		nif := netns.NamedNifs[processor.NifName]
		if nif == nil {
			continue
		}
		if nif.Nif().Kind != "tuntap" {
			slog.Error("mixed-up netdev in netns",
				slog.String("interface", nif.Nif().Name),
				slog.Uint64("netns", netns.ID().Ino))
			continue
		}
		tuntap := nif.(TunTap).TunTap()
		tuntap.Processors = append(tuntap.Processors, proc)
	}
}

// TapTunProcessors collects the processes serving TAP/TUN network interfaces
// during an lxkns discovery, when registering its [TapTunProcessors.Visit]
// method using [discover.WithFdCallback].
type TapTunProcessors struct {
	processors []tuntapProcessor
}

// tuntapProcessor describes a process serving a TAP/TUN network interface
// (netdev).
type tuntapProcessor struct {
	// PID instead of *model.Process, as the processes might have been collected
	// from a different discovery's process table than the one later resolved
	// against (such as when NewNetworkNamespaces scans on its own).
	PID     model.PIDType
	NifName string
	NetnsID species.NamespaceID
}

// getTapNetdevNetnsFd takes an open file descriptor referencing a TAP/TUN
// netdev and returns a new file descriptor referencing the network namespace
// the TAP/TUN netdev is currently placed in.
//
// For background information, please see:
// https://unix.stackexchange.com/a/743003
func getTapNetdevNetnsFd(fd int) (int, error) {
	return ioctl.RetFd(fd, ioctl.IO('T', 227))
}

// Visit checks an open file descriptor of a process for serving a TAP/TUN
// network interface and if so, collects the process, the interface name, and
// the network namespace of the interface.
func (t *TapTunProcessors) Visit(fd discover.FdEntry) {
	// TAP/TUN fds are always character devices, such as /dev/net/tun.
	if !strings.HasPrefix(fd.Target, "/dev/") {
		return
	}
	pid := fd.Process.PID
	iffName := iff("/proc/" + strconv.Itoa(int(pid)) + "/fdinfo/" + strconv.Itoa(fd.Fd))
	if iffName == "" {
		return
	}
	pidfd, err := fd.PidFd()
	if err != nil {
		return
	}
	taptunFd, err := unix.PidfdGetfd(pidfd, fd.Fd, 0)
	if err != nil {
		return
	}
	netnsFd, err := getTapNetdevNetnsFd(taptunFd)
	_ = unix.Close(taptunFd)
	if err != nil {
		return
	}
	netnsID, err := ops.NamespaceFd(netnsFd).ID()
	_ = unix.Close(netnsFd)
	if err != nil {
		return
	}
	t.processors = append(t.processors, tuntapProcessor{
		PID:     pid,
		NifName: iffName,
		NetnsID: netnsID,
	})
}

// iff returns the value of the "iff:" entry from a /proc/$PID/fdinfo/$FD pseudo
// file, if any, otherwise an empty string.
func iff(path string) string {
	const iffEntry = "iff:\t"

	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, iffEntry) {
			return line[len(iffEntry):]
		}
	}
	_ = scanner.Err()
	return ""
}

// hasTapTun returns true if any TAP or TUN network interface has been found,
// otherwise false.
func hasTapTun(netspaces NetworkNamespaces) bool {
	for _, netns := range netspaces {
		for _, nif := range netns.Nifs {
			if nif.Nif().Kind == "tuntap" {
				return true
			}
		}
	}
	return false
}
