// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package nlnetdev

import (
	"errors"
	"fmt"

	"github.com/mdlayher/genetlink"
	"github.com/mdlayher/netlink"

	"github.com/siemens/ghostwire/v2/netdev/rxtxlayout"
)

// netdevMapOfQueues maps network interface indices to sets of RX/TX queues for
// their respective network interfaces. This type is used with the NETLINK
// netdev API dump operation as this operation returns all queues flat with
// interface indices, so we bring some useful order into the flat dump.
type netdevMapOfQueues map[uint32]Queues

// Queues is a set of RX/TX queue information for a single particular network
// interface (as referenced in the IfIndex field).
type Queues []*Queue

// Queue describes an individual netdev RX or TX Queue on the NETLINK netdev API
// level. This is later transformed into the rxtxlayout model with proper
// references between queues and the other netdev model elements.
type Queue struct {
	ID      uint32               // queue ID
	Type    rxtxlayout.QueueType // RX or TX
	IfIndex uint32               // ifindex of network interface with this queue
	NapiID  uint32               // NAPI ID, or zero
}

// netdevsQueues returns the RX/TX queue sets for all network interfaces; because
// that is what the NETLINK netdev DUMP operation does.
//
// Important: network interfaces in “down” operational state won't be listed;
// the rationale is that NAPIs aren't registered at all in operstate “down”.
func (c *Conn) netdevsQueues() (netdevMapOfQueues, error) {
	queues, err := c.queues(0)
	if err != nil {
		return nil, err
	}
	m := netdevMapOfQueues{}
	for _, queue := range queues {
		m[queue.IfIndex] = append(m[queue.IfIndex], queue)
	}
	return m, nil
}

// Queues returns the list of Queue objects describing the RX and TX queues of
// of a specific network interface.
func (c *Conn) Queues(ifindex uint32) (Queues, error) {
	if ifindex == 0 {
		return nil, fmt.Errorf("invalid ifindex %d", ifindex)
	}
	return c.queues(ifindex)
}

// MaxQueueIDs returns the largest allowed queue IDs for RX and TX queues
// respectively. This assumes that a network interface driver uses continous
// queue IDs starting at zero.
func (qs Queues) MaxQueueIDs() (rxid, txid uint) {
	for _, queue := range qs {
		switch queue.Type {
		case rxtxlayout.NETDEV_QUEUE_TYPE_RX:
			rxid = max(rxid, uint(queue.ID))
		case rxtxlayout.NETDEV_QUEUE_TYPE_TX:
			txid = max(txid, uint(queue.ID))
		}
	}
	return
}

// queues return the list of Queue objects for either all network interfaces in
// the network namespace the connection leads to (ifindex==0), or for the
// specific network interface only specified by ifindex.
func (c *Conn) queues(ifindex uint32) (Queues, error) {
	var data []byte
	if ifindex != 0 {
		enco := netlink.NewAttributeEncoder()
		enco.Uint32(rxtxlayout.NETDEV_A_QUEUE_IFINDEX, ifindex)
		var err error
		data, err = enco.Encode()
		if err != nil {
			return nil, fmt.Errorf("cannot encode interface index attribute, reason: %w", err)
		}
	}
	req := genetlink.Message{
		Header: genetlink.Header{
			Command: NETDEV_CMD_QUEUE_GET,
			Version: c.family.Version,
		},
		Data: data,
	}
	replies, err := c.conn.Execute(req, c.family.ID, netlink.Request|netlink.Dump)
	if err != nil {
		return nil, fmt.Errorf("cannot query netdev queue information, reason %w", err)
	}
	queues := make(Queues, 0, len(replies))
	for _, msg := range replies {
		queue, err := newQueue(msg.Data)
		if err != nil {
			return nil, err
		}
		queues = append(queues, queue)
	}
	return queues, nil
}

// newQueue returns a Queue object from the supplied binary data, or an error if
// decoding the binary data fails.
func newQueue(b []byte) (*Queue, error) {
	deco, err := netlink.NewAttributeDecoder(b)
	if err != nil {
		return nil, fmt.Errorf("cannot decode netdev queue information, reason: %w", err)
	}

	const (
		hasID = 1 << iota
		hasIFINDEX
		hasTYPE
	)
	const mandatory = hasID | hasIFINDEX | hasTYPE

	queue := &Queue{}
	seen := 0
	for deco.Next() {
		switch deco.Type() {
		case rxtxlayout.NETDEV_A_QUEUE_ID:
			queue.ID = deco.Uint32()
			seen |= hasID
		case rxtxlayout.NETDEV_A_QUEUE_TYPE:
			queue.Type = rxtxlayout.QueueType(deco.Uint32())
			seen |= hasTYPE
		case rxtxlayout.NETDEV_A_QUEUE_IFINDEX:
			queue.IfIndex = deco.Uint32()
			seen |= hasIFINDEX
		case rxtxlayout.NETDEV_A_QUEUE_NAPI_ID:
			queue.NapiID = deco.Uint32()
		}
	}
	if err := deco.Err(); err != nil {
		return nil, fmt.Errorf("cannot decode netdev queue information, reason: %w", err)
	}
	if seen&mandatory != mandatory {
		return nil, errors.New("incomplete netdev queue information")
	}
	return queue, nil
}
