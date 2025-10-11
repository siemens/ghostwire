// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import (
	"fmt"
	"slices"
	"strings"

	"github.com/thediveo/lxkns/model"
)

// Source identifies the source(s) of information in form of a bit field.
type Source uint

const (
	SourceNetlinkNetdev Source = 1 << iota
	SourceSysfs
)

func (s Source) String() string {
	sources := []string{}
	if s&SourceNetlinkNetdev != 0 {
		sources = append(sources, "netlink")
	}
	if s&SourceSysfs != 0 {
		sources = append(sources, "sysfs")
	}
	return strings.Join(sources, ",")
}

// Netdev queue attribute identifiers. Please see also the Linux kernel source:
// [netdev.h].
//
// [netdev.h]: https://elixir.bootlin.com/linux/v6.9.5/source/include/uapi/linux/netdev.h#L129
const (
	_ = iota
	NETDEV_A_QUEUE_ID
	NETDEV_A_QUEUE_IFINDEX
	NETDEV_A_QUEUE_TYPE
	NETDEV_A_QUEUE_NAPI_ID
)

// QueueType indicates whether a given netdev queue is for RX (receiving) or TX
// (transmitting).
type QueueType uint32

// Netdev queues can be either RX or TX. Please see also the Linux kernel
// source: [netdev.h].
//
// [netdev.h]: https://elixir.bootlin.com/linux/v6.9.5/source/include/uapi/linux/netdev.h#L68
const (
	NETDEV_QUEUE_TYPE_RX QueueType = iota
	NETDEV_QUEUE_TYPE_TX
)

// Netdev NAPI attribute identifiers. Please see also the Linux kernel source:
// [netdev.h].
//
// [netdev.h]: https://elixir.bootlin.com/linux/v6.9.5/source/include/uapi/linux/netdev.h#L119
const (
	_ = iota
	NETDEV_A_NAPI_IFINDEX
	NETDEV_A_NAPI_ID
	NETDEV_A_NAPI_IRQ
	NETDEV_A_NAPI_PID
)

// String returns either “RX” or “TX”, depending on the QueueType.
func (qt QueueType) String() string {
	switch qt {
	case NETDEV_QUEUE_TYPE_RX:
		return "RX"
	case NETDEV_QUEUE_TYPE_TX:
		return "TX"
	}
	return fmt.Sprintf("QueueType(%d)", qt)
}

// NetdevsByNetns maps (network) namespaces to the Netdev objects they contain.
type NetdevsByNetns map[model.Namespace][]*Netdev

// Netdev describes a network interface in terms of its queues, IRQs, and
// associated kernel threads. It can be correctly un/marshalled from/to JSON
// despite being a cyclic information model regarding Queues, NAPIs, and IRQs.
//
// In an ideal world we would get all the information we need and it would
// basically look like this:
//
//	Netdev -> Queue -> NAPI -> NAPI kernel thread
//	Netdev -> Queue -> IRQ -> IRQ number and IRQ kernel thread
//
// Unfortunately, most drivers do not fully fill in their NAPI structures and in
// turn we end up with relationships missing when using the NETLINK netdev API.
// And when using sysfs and procfs-provided information we end up with other
// relationships missing.
//
//	Netdev -> NAPI -> NAPI kernel thread
//	Netdev -> Queue -> IRQ -> IRQ number and IRQ kernel thread
//
// So we then lack the link Queue -> NAPI; that is, we know the queues and
// NAPIs, but we don't know the correct mapping. Thus, the Queue and NAPI
// objects are present with the Netdev object, but the pointers between Queue
// and NAPI objects are nil.
//
// # Standard Marshalling
//
//   - custom Queue marshalling
//   - custom NAPI marshalling
//   - standard IRQ marshalling
//
// # Custom Unmarshalling
//
//   - handles Queue unmarshalling explicitly, using jsonNetdev.
//   - handles NAPI unmarshalling explicitly, using jsonNAPI.
//   - handles IRQ marshalling explicitly, but via standard unmarshalling.
//
// # Note
//
// Storing the Queue objects in a simple slice instead of an indexed map should
// be sufficient even in case of queue monsters.
type Netdev struct {
	Source       Source         `json:"source"`    // where does this information originate from?
	Name         string         `json:"name"`      // ...of network interface in its network namespace.
	OriginalName string         `json:"orig-name"` // ...while in the host network namespace and not somewhere else in a container.
	Index        int            `json:"index"`     // ...of network interface in its network namespace.
	Queues       Queues         `json:"queues"`    // discovered RX and TX queues, with IRQ and maybe NAPI relationships.
	NAPIs        map[uint]*NAPI `json:"napis"`     // discovered NAPIs; maybe with IRQ and NAPI relationships.
	IRQs         map[uint]*IRQ  `json:"irqs"`      // discovered IRQs, with queue relationships.
}

// NAPI describes a single NAPI instance with its ID (when known), the interface
// it is registered with, an optional NAPI polling kernel thread, and the
// assigned IRQ.
//
// # Custom Marshalling
//
//   - handles Queue marshalling explicitly to emit an ID only.
//
// # Standard Unmarshalling
//
//   - the [IRQ] reference is handled during [Netdev] custom unmarshalling.
//   - the kernel thread [model.Process] reference is handled outside unmarshalling.
type NAPI struct {
	ID      uint           `json:"id"`    // known only when non-zero
	Index   int            `json:"index"` // corresponding network interface/netdev
	PID     model.PIDType  `json:"pid"`   // when non-zero, PID of NAPI kernel thread
	Kthread *model.Process `json:"-"`     // might be nil in case of stale PID
	IRQ     *IRQ           `json:"-"`     // related IRQ when known
}

// Queues is a set of RX/TX queue information for a single network interface.
type Queues []*Queue

// Queue describes a single RX or TX queue of a network interface. Where known,
// it references the IRQ and NAPI instance, and thus associated kernel threads
// when applicable.
type Queue struct {
	ID   uint      `json:"id"`   // queue number/ID; same ID for both RX and TX queues.
	Type QueueType `json:"type"` // either RX or TX
	NAPI *NAPI     `json:"-"`    // ...if known
	IRQ  *IRQ      `json:"-"`    // ...if known
}

// IRQ describes a single Interrupt with its number (“ID”) and optionally
// associated IRQ (“bottom half”) handler kernel thread.
type IRQ struct {
	Source  Source         `json:"source"`
	ID      uint           `json:"id"` // a.k.a. IRQ number
	PID     model.PIDType  `json:"pid"`
	Kthread *model.Process `json:"-"`
}

// Queue returns the Queue object of the specified type (RX/TX) and with the
// specified ID, or nil.
func (n Netdev) Queue(id uint, qtype QueueType) *Queue {
	idx := slices.IndexFunc(n.Queues, func(q *Queue) bool {
		return q.ID == id && q.Type == qtype
	})
	if idx < 0 {
		return nil
	}
	return n.Queues[idx]
}

// MaxQueueIDs returns the largest allowed queue IDs for RX and TX queues
// respectively. This assumes that a network interface driver uses continous
// queue IDs starting at zero.
func (qs Queues) MaxQueueIDs() (rxid, txid uint) {
	for _, queue := range qs {
		switch queue.Type {
		case NETDEV_QUEUE_TYPE_RX:
			rxid = max(rxid, queue.ID)
		case NETDEV_QUEUE_TYPE_TX:
			txid = max(txid, queue.ID)
		}
	}
	return
}

// SortNetdevsByName sorts Netdev objects by their names, with the exception of
// putting “lo” always first (so it sticks out like a sore toe).
func SortNetdevsByName(a, b *Netdev) int {
	isLoA := a.Name == "lo"
	isLoB := b.Name == "lo"
	if isLoA != isLoB {
		switch {
		case isLoA:
			return -1
		default:
			return 1
		}
	}
	return strings.Compare(a.Name, b.Name)
}

// SortQueuesByIndexAndType sorts Queue objects by their ID in increasing order,
// and then RX before TX.
func SortQueuesByIndexAndType(a, b *Queue) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return int(a.Type) - int(b.Type)
	}
}

// SortIRQsByID sorts IRQ objects by their IDs in increasing order.
func SortIRQsByID(a, b *IRQ) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return 0
	}
}

// SortNAPIsByID sorts NAPI objects by their IDs in increasing order.
func SortNAPIsByID(a, b *NAPI) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return 0
	}
}
