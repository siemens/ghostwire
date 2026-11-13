// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// NAPIKthreadMap indexes NAPI kernel threads by either their names the
// interfaces had when the NAPI kthreads were created, or by their PIDs
// (depending on use case).
//
// Unfortunately, that doesn't work well with containers when a workload does
// NAPI kthread tuning on its own: “don't do that!” and especially not on a
// generic “eth0” or such like.
//
// This guesswork indexing works best if NAPI kthread tuning was once done only
// on netdevs that currently reside in the host network namespace and especially
// not after playing shenanigans with netdev names in the host network
// namespace; meaning: there's always a way to break this scheme and that's why
// we need the NETLINK netdev API ... which at this time requires kernel 6.8+
// and netdev drivers that actually fill in the needed information (a rarety at
// the time of this writing).
type NAPIKthreadMap map[string][]*NAPI

// NewNAPIKthreadsMap returns a map of NAPI kernel threads indexed by their
// (original) network interface names, based on the passed process table.
func NewNAPIKthreadsMap(procs model.ProcessTable) NAPIKthreadMap {
	return NewKthreadsMap(procs, kvOfNAPIKthread)
}

// FillInNAPIKthreads fills in NAPI information for netdevs, based on the passed
// NAPI kernel thread map. Because the NAPI kthread map is based on (original)
// netdev names, but not network namespace'd (because the NAPI kthread naming
// scheme is in need of improvement and we can't yet rely on the NETLINK netdev
// API) callers must be very careful and pass in only netdevs from the host
// network namespace, or if from other network namespaces only when we know of
// their original name.
//
// Unfortunately, we cannot link the Queue objects to their respective NAPIs, as
// this information is not available in the sysfs.
func FillInNAPIKthreads(napikthreads NAPIKthreadMap, netdevs []*Netdev) {
	for _, ndev := range netdevs {
		name := ndev.Name
		if ndev.OriginalName != "" {
			name = ndev.OriginalName
		}
		napis := napikthreads[name]
		if napis == nil {
			continue
		}
		if ndev.NAPIs == nil {
			ndev.NAPIs = map[uint]*NAPI{}
		}
		for _, napi := range napis {
			napi.Index = ndev.Index
			ndev.NAPIs[napi.ID] = napi
		}
	}
}

const napiKthreadnamePrefix = "napi/"

// kvOfNAPIKthread returns its verdict whether a kernel thread is a NAPI
// kthread, and in case it is, the interface name it belongs to as the key to
// use for mapping this kthread.
func kvOfNAPIKthread(kthread *model.Process) (ifname string, napi *NAPI, ok bool) {
	if !strings.HasPrefix(kthread.Name, napiKthreadnamePrefix) {
		return
	}
	// Format: "napi/$IFNAME-$NAPIID"
	idx := strings.LastIndex(kthread.Name, "-")
	if idx <= len(napiKthreadnamePrefix) {
		return
	}
	ifname = kthread.Name[len(napiKthreadnamePrefix):idx]
	napiID, err := strconv.ParseUint(kthread.Name[idx+1:], 10, 32)
	if err != nil {
		return
	}
	return ifname, &NAPI{
		ID:      uint(napiID),
		PID:     kthread.PID,
		Kthread: kthread,
	}, true
}
