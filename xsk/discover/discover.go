// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

import (
	"net"

	lxknsdiscover "github.com/thediveo/lxkns/discover"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/species"
	"github.com/vishvananda/netlink"
)

// XSK represents a single open XDP socket with its relationship to one or more processes.
type XSK struct {
	Netns     model.Namespace
	Processes []*model.Process
	Diag      *netlink.XDPDiagInfoResp
	Nifname   string // name of the network interface in Netns the XSK is bound to.
}

// NetnsXSKs organizes XSK objects by their network namespaces (in particular,
// network namespace identifiers).
type NetnsXSKs map[species.NamespaceID][]XSK

// AllXSKs discovers all XDP sockets and how they relate to processes (and
// tasks), given a set of discovered network namespaces.
//
// Please note that the XSK discovery needs to scan the process file system, and
// the open file descriptors of all processes: it gathers socket inode numbers
// to match them against the XSK inode numbers we found in the dump.
//
// The discovery of AF_XDP family sockets differs from the discovery of
// especially sockets of the IP address families: there is no information
// available in /proc/net for XDP sockets. Instead, we need to “dump” XSK
// information using a NETLINK-based API. NETLINK sockets are scoped as all the
// other socket families to the network namespace the creating task/process was
// attached to when calling the socket syscall.
func AllXSKs(r *lxknsdiscover.Result) NetnsXSKs {
	netnsxsks := NetnsXSKs{}
	for _, netns := range r.Namespaces[model.NetNS] {
		for _, xsk := range discoverXsks(netns, r.SocketProcessMap, r.Processes) {
			netnsid := xsk.Netns.ID()
			netnsxsks[netnsid] = append(netnsxsks[netnsid], xsk)
		}
	}
	return netnsxsks
}

// discoverXsks discovers the XDP sockets (and their diagnosis information) in
// the specified network namespaces, returning a list of discovered XSKs. The
// discovery keeps silent in the face of errors.
func discoverXsks(netns model.Namespace, inosockmap lxknsdiscover.SocketProcesses, procstable model.ProcessTable) []XSK {
	ref := netns.Ref()
	if len(ref) != 1 {
		return nil
	}
	netnsref := ops.NewTypedNamespacePath(ref[0], species.CLONE_NEWNET)

	var xsks []XSK
	var xsksinfo []*netlink.XDPDiagInfoResp
	var dialErr error

	if err := ops.Visit(func() {
		// First dump the XSKs in this network namespace...
		xsksinfo, _ = netlink.SocketDiagXDP()
		// ...then update the XSK map and at this time also resolve the network
		// interface names from the ifindices we were only given so far.
		//
		// net.Interfaces uses RTNETLINK on Linux, so we don't need here another
		// non-stdlib dependency.
		nifs, _ := net.Interfaces()
		nifNamesByIndex := make(map[int]string, len(nifs))
		for _, nif := range nifs {
			nifNamesByIndex[nif.Index] = nif.Name
		}
		xsks = make([]XSK, 0, len(xsksinfo))
		for _, xsk := range xsksinfo {
			pids := inosockmap[uint64(xsk.XDPDiagMsg.Ino)]
			procs := make([]*model.Process, 0, len(pids))
			for _, pid := range pids {
				proc := procstable[pid]
				if proc == nil {
					continue
				}
				procs = append(procs, proc)
			}
			xsks = append(xsks, XSK{
				Netns:     netns,
				Diag:      xsk,
				Nifname:   nifNamesByIndex[int(xsk.XDPInfo.Ifindex)],
				Processes: procs,
			})
		}
	}, netnsref); err != nil || dialErr != nil {
		return nil
	}
	return xsks
}
