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

// netdevMapOfNAPIs maps network interface indices to sets of NAPI information
// for the respective network interfaces.  This type is used with the NETLINK
// netdev API dump operation as this operation returns all queues flat with
// interface indices, so we bring some useful order into the flat dump.
type netdevMapOfNAPIs map[uint32]napis

// napis is a set of NAPI instance(s) information for a single specific network
// interface (as referenced in the IfIndex field).
type napis []*nAPI

// nAPI describes a single NAPI instance for a network interface. This type is
// on the NETLINK netdev API level and has to be transformed to the rxtxlayout
// model.
type nAPI struct {
	ID      uint32 // NAPI ID
	IfIndex uint32 // ifindex of network interface with this NAPI
	HasIRQ  bool   // is IRQ information available?
	IRQ     uint32 // IRQ, if any; IRQ0 is a valid value, so check HasIRQ
	PID     uint32 // optional nAPI kthread PID
}

// nAPIs returns the nAPI sets for all network interfaces.
func (c *Conn) nAPIs() (netdevMapOfNAPIs, error) {
	nAPIs, err := c.napis(0)
	if err != nil {
		return nil, err
	}
	m := netdevMapOfNAPIs{}
	for _, nAPI := range nAPIs {
		m[nAPI.IfIndex] = append(m[nAPI.IfIndex], nAPI)
	}
	return m, nil
}

func (c *Conn) napis(ifindex uint32) ([]*nAPI, error) {
	var data []byte
	if ifindex != 0 {
		enco := netlink.NewAttributeEncoder()
		enco.Uint32(rxtxlayout.NETDEV_A_NAPI_IFINDEX, ifindex)
		var err error
		data, err = enco.Encode()
		if err != nil {
			return nil, fmt.Errorf("cannot encode interface index attribute, reason: %w", err)
		}
	}
	req := genetlink.Message{
		Header: genetlink.Header{
			Command: NETDEV_CMD_NAPI_GET,
			Version: c.family.Version,
		},
		Data: data,
	}
	replies, err := c.conn.Execute(req, c.family.ID, netlink.Request|netlink.Dump)
	if err != nil {
		return nil, fmt.Errorf("cannot query netdev nAPI information, reason %w", err)
	}
	nAPIs := make([]*nAPI, 0, len(replies))
	for _, msg := range replies {
		nAPI, err := newNAPI(msg.Data)
		if err != nil {
			return nil, err
		}
		nAPIs = append(nAPIs, nAPI)
	}
	return nAPIs, nil
}

// newNAPI returns a nAPI object from the supplied binary data, or an error if
// decoding the binary data fails.
func newNAPI(b []byte) (*nAPI, error) {
	deco, err := netlink.NewAttributeDecoder(b)
	if err != nil {
		return nil, fmt.Errorf("cannot decode netdev NAPI information, reason: %w", err)
	}

	const (
		hasID = 1 << iota
		hasIFINDEX
	)
	const mandatory = hasID | hasIFINDEX

	nAPI := &nAPI{}
	seen := 0
	for deco.Next() {
		switch deco.Type() {
		case rxtxlayout.NETDEV_A_NAPI_ID:
			nAPI.ID = deco.Uint32()
			seen |= hasID
		case rxtxlayout.NETDEV_A_NAPI_IFINDEX:
			nAPI.IfIndex = deco.Uint32()
			seen |= hasIFINDEX
		case rxtxlayout.NETDEV_A_NAPI_IRQ:
			nAPI.HasIRQ = true
			nAPI.IRQ = deco.Uint32()
		case rxtxlayout.NETDEV_A_NAPI_PID:
			nAPI.PID = deco.Uint32()
		}
	}
	if err := deco.Err(); err != nil {
		return nil, fmt.Errorf("cannot decode netdev NAPI information, reason: %w", err)
	}
	if seen&mandatory != mandatory {
		return nil, errors.New("incomplete netdev NAPI information")
	}
	return nAPI, nil
}
