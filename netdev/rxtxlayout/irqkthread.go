// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// IRQKthreadMap indexes IRQ kernel threads by their IRQ ID/number. Use
// [NewIRQKthreadsMap] to build one from a [model.ProcessTable]. This map is
// then used with [FillInIRQKthreads] to link [IRQ] objects with their
// corresponding kthread [model.Process] objects. And yes, even if they are
// termed “kernel threads” they appear in the procfs as child processes (and not
// tasks!) of the “kthreadd” parent process with PID 2.
type IRQKthreadMap = map[uint]*model.Process

// NewIRQKthreadsMap returns a map of IRQ kernel threads indexed by their IRQ
// numbers, based on the passed process table.
//
// IRQ kthreads are identified by their names in form of “irq/$IRQ-$SOMETHING”.
func NewIRQKthreadsMap(procs model.ProcessTable) IRQKthreadMap {
	return NewKthreadMap(procs, kvOfIRQKthread)
}

// FillInIRQKthreads links IRQ kernel threads to the IRQs for the passed
// netdevs. It expects a map of IRQ kernel threads, indexed by IRQ ID/number, as
// returned by MapOfIRQKthreads. As netdevs are network namespace'd, but IRQs
// are system-wide, indexing IRQ kernel threads is a separate one-time
// operation, whereas FillInIRQKthreads is potentially called multiple times,
// for each network namespace once.
func FillInIRQKthreads(irqkthreads IRQKthreadMap, netdevs []*Netdev) {
	for _, ndev := range netdevs {
		for _, irq := range ndev.IRQs {
			kthread, ok := irqkthreads[irq.ID]
			if !ok {
				continue
			}
			irq.PID = kthread.PID
			irq.Kthread = kthread
		}
	}
}

const irqKthreadnamePrefix = "irq/"

// kvOfIRQKthread returns its verdict whether a kernel thread is an IRQ kthread,
// and in case it is, the IRQ number as the key to use for mapping this
// particular kthread.
func kvOfIRQKthread(kthread *model.Process) (irq uint, kt *model.Process, ok bool) {
	if !strings.HasPrefix(kthread.Name, irqKthreadnamePrefix) {
		return
	}
	irqfield, _, ok := strings.Cut(kthread.Name[len(irqKthreadnamePrefix):], "-")
	if !ok {
		return
	}
	irqno, err := strconv.ParseUint(irqfield, 10, 32)
	if err != nil {
		return 0, nil, false
	}
	return uint(irqno), kthread, true
}
