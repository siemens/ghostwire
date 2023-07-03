// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"net"
	"strings"

	"github.com/siemens/ghostwire/v2/network"

	"github.com/thediveo/lxkns/model"
)

// networkInterface is the API v1 JSON representation of an individual network
// interface.
type networkInterface struct {
	ID            string            `json:"id"`
	Kind          string            `json:"kind"`
	Name          string            `json:"name"`
	Alias         string            `json:"alias,omitempty"`
	Index         int               `json:"index"`
	Addresses     addresses         `json:"addresses"`
	Operstate     string            `json:"operstate"`
	Physical      bool              `json:"physical"`
	Promiscuous   bool              `json:"promisc"`
	Labels        model.Labels      `json:"labels,omitempty"`
	Master        *nifRef           `json:"master,omitempty"`
	MacvlanMaster *nifRef           `json:"macvlan,omitempty"` // admittedly stupid naming.
	Macvlans      []*nifRef         `json:"macvlans,omitempty"`
	Slaves        []*nifRef         `json:"slaves,omitempty"`
	Peer          *peerNifRef       `json:"peer,omitempty"`
	TunTap        *tuntapConfig     `json:"tuntap,omitempty"`
	Vxlan         *vxlanConfig      `json:"vxlan,omitempty"`
	Vlan          *vlanConfig       `json:"vlan,omitempty"`
	SRIOVRole     network.SRIOVRole `json:"sr-iov-role,omitempty"`
	PF            *nifRef           `json:"pf,omitempty"`
}

type addresses struct {
	MAC  string            `json:"mac"`
	IPv4 []network.Address `json:"ipv4"`
	IPv6 []network.Address `json:"ipv6"`
}

// vxlanConfig is optional and carries VXLAN-specific network interface
// information.
type vxlanConfig struct {
	UnderlayID      string          `json:"idref"`
	VID             uint32          `json:"vid"`
	ArpProxy        bool            `json:"arp-proxy"`
	Source          *sourceIP       `json:"source,omitempty"`
	SourcePortRange sourcePortRange `json:"source-portrange,omitempty"`
	Remote          *remoteIP       `json:"remote,omitempty"`
	RemotePort      uint16          `json:"remote-port"`
}

type processor owner

// tuntapConfig is optional and carries TUN/TAP-specific network interface
// information, especially whether it is a TAP or a TUN.
type tuntapConfig struct {
	Mode       string      `json:"mode"`
	Processors []processor `json:"processors"`
}

type sourceIP struct {
	IPv4 net.IP `json:"source_ipv4,omitempty"`
	IPv6 net.IP `json:"source_ipv6,omitempty"`
}

type sourcePortRange struct {
	Low  uint16 `json:"low"`
	High uint16 `json:"high"`
}

type remoteIP struct {
	IPv4 net.IP `json:"remote_ipv4,omitempty"`
	IPv6 net.IP `json:"remote_ipv6,omitempty"`
}

type vlanConfig struct {
	VID          uint16 `json:"vid"`
	VlanProtocol int    `json:"vlan-protocol"`
}

// newNif returns a properly set up JSON networkInterface, given a discovered
// network.NetworkInterface.
func newNif(nif network.Interface) networkInterface {
	// Handle network interface having a bridge master.
	var master *nifRef
	if nif.Nif().Bridge != nil {
		master = newNifRef(nif.Nif().Bridge)
	}
	// Handle a VETH peer, if present.
	var peer *peerNifRef
	if veth, ok := nif.(network.Veth); ok {
		if p := veth.Veth().Peer; p != nil {
			peer = &peerNifRef{
				ID:    nifID(p),
				Index: p.Nif().Index,
				Name:  p.Nif().Name,
			}
		}
	}
	// Handle a MACVLAN master.
	var macvlanmaster *nifRef
	if macvlan, ok := nif.(network.Macvlan); ok {
		macvlanmaster = newNifRef(macvlan.Macvlan().Master)
	}
	// Handle slaves of bridges, but also of other masters, and even of PFs.
	var slaves []*nifRef
	if bridge, ok := nif.(network.Bridge); ok {
		for _, port := range bridge.Bridge().Ports {
			slaves = append(slaves, newNifRef(port))
		}
	}
	var macvlans []*nifRef
	for _, nif := range nif.Nif().Slaves {
		if nif.Nif().Kind == "macvlan" {
			macvlans = append(macvlans, newNifRef(nif))
			continue
		}
		slaves = append(slaves, newNifRef(nif))
	}
	// Handle a VF.
	var pf *nifRef
	if pfnif := nif.Nif().PF; pfnif != nil {
		pf = newNifRef(pfnif)
	}
	// Handle a VXLAN.
	var vxlancfg *vxlanConfig
	if vxlan, ok := nif.(network.Vxlan); ok {
		vx := vxlan.Vxlan()
		var srcportrange sourcePortRange
		if vx.SourcePortLow != 0 {
			srcportrange = sourcePortRange{
				Low:  vx.SourcePortLow,
				High: vx.SourcePortHigh,
			}
		}
		var src *sourceIP
		if len(vx.Sourcev4) != 0 || len(vx.Sourcev6) != 0 {
			src = &sourceIP{
				IPv4: vx.Sourcev4,
				IPv6: vx.Sourcev6,
			}
		}
		var rem *remoteIP
		if len(vx.Groupv4) != 0 || len(vx.Groupv6) != 0 {
			rem = &remoteIP{
				IPv4: vx.Groupv4,
				IPv6: vx.Groupv6,
			}
		}
		vxlancfg = &vxlanConfig{
			UnderlayID:      nifID(vx.Master),
			VID:             vx.VID,
			ArpProxy:        vx.ArpProxy,
			Source:          src,
			SourcePortRange: srcportrange,
			Remote:          rem,
			RemotePort:      vx.DestinationPort,
		}
	}
	// Handle a TUN/TAP.
	var tuntapcfg *tuntapConfig
	if tuntap, ok := nif.(network.TunTap); ok {
		tt := tuntap.TunTap()
		tuntapcfg = &tuntapConfig{}
		switch tt.Mode {
		case network.TunTapModeTap:
			tuntapcfg.Mode = "tap"
		case network.TunTapModeTun:
			tuntapcfg.Mode = "tun"
		}
		processors := make([]processor, 0, len(tt.Processors))
		for _, proc := range tt.Processors {
			processors = append(processors, processor{
				PID:          proc.PID,
				Cmdline:      strings.Join(proc.Cmdline, " "),
				ContainerRef: cntrID(leader(proc)),
			})
		}
		tuntapcfg.Processors = processors
	}
	// Handle a VLAN.
	var vlancfg *vlanConfig
	if vlan, ok := nif.(network.Vlan); ok {
		vl := vlan.Vlan()
		vlancfg = &vlanConfig{
			VID:          vl.VID,
			VlanProtocol: int(vl.VlanProtocol),
		}
		master = newNifRef(vl.Master)
	}

	nifa := nif.Nif()
	return networkInterface{
		ID:    nifID(nif),
		Name:  nifa.Name,
		Alias: nifa.Alias,
		Index: nifa.Index,
		Kind:  nifa.Kind,
		Addresses: addresses{
			MAC:  nifa.L2Addr.String(),
			IPv4: nifa.Addrsv4,
			IPv6: nifa.Addrsv6,
		},
		Physical:      nifa.Physical,
		Promiscuous:   nifa.Promiscuous,
		Operstate:     nifa.State.Name(),
		Labels:        nifa.Labels,
		Master:        master,
		MacvlanMaster: macvlanmaster,
		Macvlans:      macvlans,
		Peer:          peer,
		Slaves:        slaves,
		Vxlan:         vxlancfg,
		TunTap:        tuntapcfg,
		Vlan:          vlancfg,
		SRIOVRole:     nifa.SRIOVRole,
		PF:            pf,
	}
}

// nifRef represents an intra-JSON document network interface reference.
type nifRef struct {
	ID    string `json:"idref"`
	Index int    `json:"index"`
	Name  string `json:"name"`
}

// peerNifRef represents an intra-JSON document VETH network interface
// reference. And yes, that has been a bad idea right from the beginning, why
// did I do this?
type peerNifRef struct {
	ID    string `json:"peer-idref"`
	Index int    `json:"peer-index"`
	Name  string `json:"peer-name"`
}

// newNifRef returns a new intra-JSON document network interface reference.
func newNifRef(nif network.Interface) *nifRef {
	if nif == nil {
		return nil
	}
	return &nifRef{
		ID:    nifID(nif),
		Index: nif.Nif().Index,
		Name:  nif.Nif().Name,
	}
}
