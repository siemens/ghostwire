// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/siemens/ghostwire/v2/network"
	"golang.org/x/exp/slices"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/netdb"
)

type ipvxPorts struct {
	IPv4 ports `json:"ipv4"`
	IPv6 ports `json:"ipv6"`
}

type ports []network.ProcessSocket

// port describes actually a local socket, but anyway...
type port struct {
	Family            network.AddressFamily `json:"family"`
	Protocol          string                `json:"protocol"`
	LocalAddress      net.IP                `json:"local-address"`
	LocalPort         uint16                `json:"local-port"`
	LocalServiceName  string                `json:"local-servicename"`
	RemoteAddress     net.IP                `json:"remote-address"`
	RemotePort        uint16                `json:"remote-port"`
	RemoteServiceName string                `json:"remote-servicename"`
	State             string                `json:"state"`
	Macrostate        string                `json:"macrostate"`
	Owners            []owner               `json:"owners"`
	NifRefs           []string              `json:"network-interface-idrefs"`
}

// owner describes a process attached to a particular transport port (rather,
// attached to a socket that relates to the port).
type owner struct {
	PID          model.PIDType `json:"pid"`
	Cmdline      string        `json:"cmdline"`
	ContainerRef string        `json:"container-idref,omitempty"`
}

func (p ports) MarshalJSON() ([]byte, error) {
	prts := make([]port, 0, len(p))
	for _, prt := range p {
		protocol := strings.ToLower(prt.Protocol.String())
		// Gather the JSON document-local network interface identifier
		// references for the network interfaces covered by this socket...
		nifrefs := make([]string, 0, len(prt.Nifs))
		for _, nif := range prt.Nifs {
			nifrefs = append(nifrefs, nifID(nif))
		}
		// Gather the "ownership" information about the processes using a socket
		// for this port...
		owners := make([]owner, 0, len(prt.Processes))
		for _, proc := range prt.Processes {
			owners = append(owners, owner{
				PID:          proc.PID,
				Cmdline:      strings.Join(proc.Cmdline, " "),
				ContainerRef: cntrID(leader(proc)),
			})
		}
		prts = append(prts, port{
			Family:            prt.Family,
			Protocol:          protocol,
			LocalAddress:      prt.LocalIP,
			LocalPort:         prt.LocalPort,
			LocalServiceName:  serviceName(prt.LocalPort, protocol),
			RemoteAddress:     prt.RemoteIP,
			RemotePort:        prt.RemotePort,
			RemoteServiceName: serviceName(prt.RemotePort, protocol),
			State:             prt.State.String(),
			Macrostate:        prt.SimplifiedState.String(),
			Owners:            owners,
			NifRefs:           nifrefs,
		})
	}
	return json.Marshal(prts)
}

// serviceName returns the name (if known) for the service on the given port and
// transport protocol name ("tcp" or "udp"). Otherwise, it returns "".
func serviceName(port uint16, protocol string) string {
	if service := netdb.ServiceByPort(int(port), protocol); service != nil {
		return service.Name
	}
	return ""
}

// leader returns the leader process for a specified process as seen in the
// network namespace of this process, otherwise the originally specified
// process.
func leader(someproc *model.Process) *model.Process {
	netns := someproc.Namespaces[model.NetNS]
	if netns == nil {
		return someproc
	}
	leaders := netns.LeaderPIDs()
	for proc := someproc; proc != nil; proc = proc.Parent {
		if slices.Contains(leaders, proc.PID) {
			return proc
		}
	}
	return someproc
}
