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
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"golang.org/x/sys/unix"
)

// resolveTapTunProcessors checks for the presence of TAP and/or TUN network
// interfaces and then resolves their serving processes ("processors").
func resolveTapTunProcessors(netspaces NetworkNamespaces, allprocs model.ProcessTable) {
	if !hasTapTun(netspaces) {
		return
	}
	processors := discoverProcessors(allprocs)
	for _, processor := range processors {
		netns := netspaces[processor.NetnsID]
		if netns == nil {
			slog.Warn("TAP/TUN serving process related to unknown netns",
				slog.String("name", processor.Process.Name),
				slog.Int("pid", int(processor.Process.PID)),
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
		tuntap.Processors = append(tuntap.Processors, processor.Process)
	}
}

// tuntapProcessor describes a Process serving a TAP/TUN network interface
// (netdev).
type tuntapProcessor struct {
	Process *model.Process
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

// discoverProcessors returns a list of processes that have file descriptors
// referencing TAP/TUN network devices.
func discoverProcessors(allprocs model.ProcessTable) []tuntapProcessor {
	var processors []tuntapProcessor
	for pid, proc := range allprocs {
		base := "/proc/" + strconv.Itoa(int(pid)) + "/fdinfo"
		fdinfoEntries, err := os.ReadDir(base)
		if err != nil {
			continue
		}
		var pidfd int // as we have stdin, stdout, stderr always connected, this fd cannot be 0.
	scanFds:
		for _, fdInfoEntry := range fdinfoEntries {
			iffName := iff(base + "/" + fdInfoEntry.Name())
			if iffName == "" {
				continue
			}

			if pidfd == 0 {
				pidfd, err = unix.PidfdOpen(int(pid), 0)
				if err != nil {
					break scanFds
				}
			}

			fd, err := strconv.ParseUint(fdInfoEntry.Name(), 10, strconv.IntSize)
			if err != nil {
				continue
			}
			taptunFd, err := unix.PidfdGetfd(pidfd, int(fd), 0)
			if err != nil {
				continue
			}
			netnsFd, err := getTapNetdevNetnsFd(taptunFd)
			_ = unix.Close(taptunFd)
			if err != nil {
				continue
			}
			netnsID, err := ops.NamespaceFd(netnsFd).ID()
			_ = unix.Close(netnsFd)
			if err != nil {
				continue
			}

			processors = append(processors, tuntapProcessor{
				Process: proc,
				NifName: iffName,
				NetnsID: netnsID,
			})
		}
		if pidfd > 0 {
			_ = unix.Close(pidfd)
			pidfd = 0
		}
	}
	return processors
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
