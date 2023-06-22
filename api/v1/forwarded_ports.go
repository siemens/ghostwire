// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"encoding/json"
	"net"
	"strings"

	"github.com/siemens/ghostwire/v2/network"
	"golang.org/x/sys/unix"
)

type ipvxForwardedPorts struct {
	IPv4 forwardedPorts `json:"ipv4"`
	IPv6 forwardedPorts `json:"ipv6"`
}

type forwardedPorts []network.ForwardedPort

type forwardedPort struct {
	Family             network.AddressFamily `json:"family"`
	Protocol           string                `json:"protocol"`
	IP                 net.IP                `json:"ip"`
	Port               uint16                `json:"port"`
	ServiceName        string                `json:"servicename"`
	ForwardIP          net.IP                `json:"forward-ip"`
	ForwardPort        uint16                `json:"forward-port"`
	ForwardServiceName string                `json:"forward-servicename"`
	NetnsID            uint64                `json:"netnsid"`
	Owners             []owner               `json:"owners"`
	NifRefs            []string              `json:"network-interface-refs"`
}

func (f forwardedPorts) MarshalJSON() ([]byte, error) {
	prts := make([]forwardedPort, 0, len(f))
	for _, fp := range f {
		fam := network.AddressFamily(unix.AF_INET)
		if len(fp.IP) == net.IPv6len {
			fam = network.AddressFamily(unix.AF_INET6)
		}
		var netnsid uint64
		if fp.DestinationNetns != nil {
			netnsid = fp.DestinationNetns.ID().Ino
		}
		// Gather the JSON document-local network interface identifier
		// references for the network interfaces covered by this forwarded
		// port...
		nifrefs := make([]string, 0, len(fp.Nifs))
		for _, nif := range fp.Nifs {
			nifrefs = append(nifrefs, nifID(nif))
		}
		// Gather the "ownership" information about the processes using a socket
		// for this forwarded port...
		owners := make([]owner, 0, len(fp.Processes))
		for _, proc := range fp.Processes {
			owners = append(owners, owner{
				PID:          proc.PID,
				Cmdline:      strings.Join(proc.Cmdline, " "),
				ContainerRef: cntrID(leader(proc)),
			})
		}
		prts = append(prts, forwardedPort{
			Family:             fam,
			Protocol:           strings.ToLower(fp.Protocol.String()),
			IP:                 fp.IP,
			Port:               fp.PortMin,
			ServiceName:        serviceName(fp.PortMin, fp.ForwardedPortRange.Protocol),
			ForwardIP:          fp.ForwardIP,
			ForwardPort:        fp.ForwardPortMin,
			ForwardServiceName: serviceName(fp.ForwardPortMin, fp.ForwardedPortRange.Protocol),
			NetnsID:            netnsid,
			Owners:             owners,
			NifRefs:            nifrefs,
		})
	}
	return json.Marshal(prts)
}
