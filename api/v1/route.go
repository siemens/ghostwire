// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"fmt"
	"net"

	"github.com/siemens/ghostwire/v2/network"
)

// ipvxRoutes is the set of IPv4 and IPv6 routes discovered in a single network
// namespace.
type ipvxRoutes struct {
	IPv4 routes `json:"ipv4"`
	IPv6 routes `json:"ipv6"`
}

// routes represents a JSON marshallable list of routes (for a single address
// family)
type routes []network.Route

// route describes a single route and is marshallable to JSON.
type route struct {
	Type           string                `json:"type"`
	Family         network.AddressFamily `json:"family"`
	Destination    net.IP                `json:"destination"`
	DestinationLen int                   `json:"destination-prefixlen"`
	Index          int                   `json:"index,omitempty"`
	NifRef         string                `json:"network-interface-idref,omitempty"`
	NextHop        net.IP                `json:"next-hop,omitempty"`
	Preference     string                `json:"preference"`
	Priority       int                   `json:"priority"`
	Table          int                   `json:"table"`
}

// MarshalJSON marshals a list of routes into JSON format.
func (r routes) MarshalJSON() ([]byte, error) {
	rts := make([]route, 0, len(r))
	for _, rt := range r {
		rts = append(rts, route{
			Type:           rt.Type.String(),
			Family:         rt.Family,
			Destination:    rt.Destination.IP,
			DestinationLen: rt.DestinationPrefixLen,
			Index:          rt.Index,
			NifRef:         nifID(rt.Nif),
			NextHop:        rt.NextHop,
			Preference:     fmt.Sprintf("%02b", rt.Preference&0x03),
			Priority:       rt.Priority,
			Table:          rt.Table,
		})
	}
	return json.Marshal(rts)
}
