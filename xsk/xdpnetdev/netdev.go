// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

//go:generate go run github.com/cilium/ebpf/cmd/bpf2go bpf xsk_dispatcher.c -- -I./_headers

package xdpnetdev

import (
	"errors"
	"fmt"
	"net"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
)

// Netdev represents a “netdev” (network device, nee “network interface”) in
// some network namespace to which we can attach a default XDP program in order
// to correctly steer incoming traffic/packets to XSKs. Use [NewByName] or
// [NewByIndex] to get a Netdev object, with the XSK traffic steering eBPF
// program already attached to the specified network interface.
type Netdev struct {
	objs         bpfObjects // ebpf XDP dispatcher and XSK map eBPF objects
	netdev       link.Link  // dispatcher program link
	maxRxQueueID uint32
}

// NewByName locates the specified network interface by its name in the current
// network namespace and then attaches an XDP socket steering eBPF program to
// it, finally returning a Netdev proxy object for the network interface.
func NewByName(ifname string) (*Netdev, error) {
	iface, err := net.InterfaceByName(ifname)
	if err != nil {
		return nil, err
	}
	return NewByIndex(iface.Index)
}

// NewByIndex locates the specified network interface by its index in the
// current network namespace and then attaches an XDP socket steering eBPF
// program to it, finally returning a Netdev proxy object for the network
// interface.
//
// Make sure to call [Netdev.Release] to unload the eBPF XSK steering program
// when not needed anymore.
//
// Call [Netdev.AddXsk] to register an XDP socket file descriptor to receive
// incomming traffic on a particular RX queue of the netdev.
func NewByIndex(ifindex int) (*Netdev, error) {
	return newByIndex(ifindex, false)
}

// newByIndex loads either an XDP socket steering eBPF program for the specified
// network interface, or alternatively an all traffic dropping eBPF program when
// specifying dropall as true (in order to support certain unit tests).
//
// Make sure to call [Netdev.Release] to unload the eBPF XSK steering program
// when not needed anymore.
//
// Call [Netdev.AddXsk] to register an XDP socket file descriptor to receive
// incomming traffic on a particular RX queue of the netdev.
func newByIndex(ifindex int, dropall bool) (n *Netdev, err error) {
	n = &Netdev{
		maxRxQueueID: MaxRxQueueID(ifindex),
	}

	if err := loadBpfObjects(&n.objs, nil); err != nil {
		var veriferr *ebpf.VerifierError
		if !errors.As(err, &veriferr) {
			return nil, fmt.Errorf("cannot load XDP program, reason: %w",
				err)
		}
		details := veriferr.Error()
		return nil, fmt.Errorf("cannot load XDP program:\n%s\nreason: %w",
			details, err)
	}
	defer func() {
		if err != nil {
			_ = n.objs.Close()
			n = nil
		}
	}()

	xdpopts := link.XDPOptions{
		Program:   n.objs.XskDispatcher,
		Interface: ifindex,
	}
	if dropall {
		xdpopts.Program = n.objs.XskDropper
	}
	n.netdev, err = link.AttachXDP(xdpopts)
	if err != nil {
		err = fmt.Errorf("cannot attach XDP program to netdev with ifindex %d, reason: %w",
			ifindex, err)
		return
	}
	return
}

// Release the XDP program and associated resources attached to this netdev.
func (n *Netdev) Release() {
	_ = n.netdev.Close()
	_ = n.objs.Close()
}

// AddXsk registers an XSK, identified by its file descriptor, to receive RX
// traffic on the specified netdev RX queue.
func (n *Netdev) AddXsk(queueid int, fd int) error {
	if queueid < 0 || queueid > int(n.maxRxQueueID) {
		return fmt.Errorf("queue index %d out of range", queueid)
	}
	return n.objs.XsksMap.Put(uint32(queueid), uint32(fd))
}

// RemoveXsk unregisters from an XSK for the specified netdev RX queue.
func (n *Netdev) RemoveXsk(queueid int) error {
	if queueid < 0 || queueid > int(n.maxRxQueueID) {
		return fmt.Errorf("queue index %d out of range", queueid)
	}
	return n.objs.XsksMap.Delete(uint32(queueid))
}
