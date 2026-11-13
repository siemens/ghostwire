// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"github.com/thediveo/lxkns/model"
)

const kthreaddPID = model.PIDType(2)

// NewKthreadMap takes a process table as well as a filter function, where the
// filter function returns map keys only for those kernel threads to be mapped
// (and otherwise false), and then returns the resulting map from keys to kernel
// threads.
//
// The passed-in process table must contain the “kthreadd” parent of kernel
// threads with PID 2, otherwise NewKthreadMap returns a nil map.
//
// And yes, despite their name Linux “kernel threads” in procfs user space are
// visible as processes, not as tasks.
func NewKthreadMap[K comparable, V any](procs model.ProcessTable, fn func(kthread *model.Process) (K, V, bool)) map[K]V {
	kthreadd, ok := procs[kthreaddPID]
	if !ok {
		return nil
	}
	m := map[K]V{}
	for _, kthread := range kthreadd.Children {
		key, value, ok := fn(kthread)
		if !ok {
			continue
		}
		m[key] = value
	}
	return m
}

// NewKthreadsMap is like NewKthreadMap, but for mapping from a key to a slice
// of values, instead of just a single value.
func NewKthreadsMap[K comparable, V any](procs model.ProcessTable, fn func(kthread *model.Process) (K, V, bool)) map[K][]V {
	kthreadd, ok := procs[kthreaddPID]
	if !ok {
		return nil
	}
	m := map[K][]V{}
	for _, kthread := range kthreadd.Children {
		key, value, ok := fn(kthread)
		if !ok {
			continue
		}
		m[key] = append(m[key], value)
	}
	return m
}
