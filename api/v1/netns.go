// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"net"
	"strconv"
	"strings"

	"github.com/siemens/ghostwire/v2/network"
	"github.com/siemens/turtlefinder"
	"github.com/siemens/turtlefinder/activator/podman"
	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/engineclient/moby"
	"github.com/thediveo/whalewatcher/watcher/containerd"
	"github.com/thediveo/whalewatcher/watcher/cri"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var titler = cases.Title(language.Und)

// privilegedContainerLabelName is the label name used by this engine in the v1
// REST API to signal that a (Docker) container is privileged. This label must
// have an empty value if present; the label must be omitted completely if the
// container is not privileged (in the strict Docker sense, not the sense of
// having lots of interesting capabilities).
const privilegedContainerLabelName = "gostwire/container/privileged"

// networkNamespaces is a set of network namespaces, together with the JSON
// document-local (container) group IDs, implementing JSON marshalling. The
// (container) group IDs are automatically generated during marshalling.
type networkNamespaces struct {
	network.NetworkNamespaces
	gids map[*model.Group]string
}

// newNetworkNamespace returns a new NetworkNamespaces object, to be used for
// marshalling the information about a bunch of network namespaces (and
// associated containers) into JSON.
func newNetworkNamespace(n network.NetworkNamespaces) networkNamespaces {
	return networkNamespaces{
		NetworkNamespaces: n,
		gids:              map[*model.Group]string{},
	}
}

// groupID returns the JSON document-local identifier for the given container
// group.
func (n *networkNamespaces) groupID(group *model.Group) string {
	if id, ok := n.gids[group]; ok {
		return id
	}
	id := "group-" + strconv.Itoa(len(n.gids))
	n.gids[group] = id
	return id
}

// cntrID returns the JSON document-local identifier for the given process
// (standalone or pertaining to a container).
func cntrID(proc *model.Process) string {
	return cntrIDFromPID(proc.PID)
}

// cntrID returns the JSON document-local identifier for the given PID
// (standalone or pertaining to a container).
func cntrIDFromPID(pid model.PIDType) string {
	return "cont-" + strconv.FormatUint(uint64(pid), 10)
}

// nifID returns the JSON document-local identifier for the given network
// interface. Returns "" in case of a nil NetworkInterface pointer.
func nifID(nif network.Interface) string {
	if nif == nil {
		return ""
	}
	return "nif-" + strconv.FormatUint(nif.Interface().Nif().Netns.ID().Ino, 10) +
		"-" + strconv.FormatInt(int64(nif.Nif().Index), 10)
}

func (n *networkNamespaces) MarshalJSON() ([]byte, error) {
	b := bytes.Buffer{}
	b.WriteRune('[')
	first := true
	for _, netns := range n.NetworkNamespaces {
		if first {
			first = false
		} else {
			b.WriteRune(',')
		}
		jb, err := (*networkNamespace)(netns).marshal(n)
		if err != nil {
			return nil, err
		}
		b.Write(jb)
	}
	b.WriteRune(']')
	return b.Bytes(), nil
}

type networkNamespace network.NetworkNamespace

// NetworkNamespaceJSON describes discovery details for a single network
// namespace, marshallable into JSON.
type NetworkNamespaceJSON struct {
	ContainerGroups   []*containerGroup  `json:"container-groups"`
	Containers        []container        `json:"containers"`
	ID                string             `json:"id"`
	NetnsID           uint64             `json:"netnsid"`
	NetworkInterfaces []networkInterface `json:"network-interfaces"`
	Routes            ipvxRoutes         `json:"routes"`
	TransportPorts    ipvxPorts          `json:"transport-ports"`
	ForwardedPorts    ipvxForwardedPorts `json:"forwarded-ports"`
}

// mashal emits all the API v1 information about a single network namespace in
// JSON textual format.
func (n *networkNamespace) marshal(allnetns *networkNamespaces) ([]byte, error) {
	// the v1 REST API only knows about network namespace-sharing pod groups. It
	// doesn't care about composer groups, that is left to the receiving client
	// to understand on the basis of container labels.
	podgroups := map[string]*containerGroup{}
	cntrs := make([]container, 0)
	// Set up the "containers" and also the groups, if necessary.
	if len(n.Tenants) != 0 {
		for _, tenant := range n.Tenants {
			if tenant.Process.PPID == 0 && tenant.Process.PID == 2 {
				// skip kthreadd(2) in order to not bedazzle users.
				continue
			}
			bndcaps := hex.EncodeToString(tenant.BoundingCaps)
			if len(bndcaps) == 0 {
				bndcaps = "0x0"
			} else {
				bndcaps = "0x" + bndcaps
			}
			if c := tenant.Process.Container; c != nil {
				var gid string
				for _, g := range c.Groups {
					// Please note that Ghostwire v1 only supports "pod"-type
					// groups.
					if g.Type != kuhbernetes.PodGroupType {
						continue
					}
					gid = allnetns.groupID(g)
					if _, ok := podgroups[gid]; !ok {
						jg := &containerGroup{
							ID:       gid,
							Name:     g.Name,
							Type:     "pod",
							TypeText: "Pod",
						}
						podgroups[gid] = jg
					}
					podgroups[gid].ContainerIDs = append(podgroups[gid].ContainerIDs, cntrID(tenant.Process))
					break
				}
				status := "running"
				if c.Paused {
					status = "paused"
				}
				if _, ok := c.Labels[moby.PrivilegedLabel]; ok {
					// add the required magic label; as we don't want to modify
					// the original data, we need to first make a shallow copy
					// of the container object and then a shallow duplicate of
					// the map. Only afterwards we can add the required new
					// label.
					var cc = *c
					c = &cc
					label := model.Labels{}
					for k, v := range c.Labels {
						label[k] = v
					}
					c.Labels = label
					c.Labels[privilegedContainerLabelName] = ""
				}
				typ := v1ContainerType(c.Type)
				cntrs = append(cntrs, container{
					Cmdline:  strings.Join(tenant.Process.Cmdline, " "),
					DNS:      newDns(tenant),
					GroupID:  gid,
					ID:       cntrID(tenant.Process),
					Name:     c.Name,
					Prefix:   c.Labels[turtlefinder.TurtlefinderContainerPrefixLabelName],
					Labels:   c.Labels,
					API:      c.Engine.API,
					PID:      c.PID,
					PIDNS:    c.Process.Namespaces[model.PIDNS].ID().Ino,
					PIDNSRef: pidnsID(c.Process.Namespaces[model.PIDNS]),
					CapBnd:   bndcaps,
					Status:   status,
					Type:     typ,
					TypeText: titler.String(typ),
				})
			} else {
				// It's a stand-alone process
				status := "running"
				if tenant.Process.FridgeFrozen {
					status = "paused"
				}
				cntrs = append(cntrs, container{
					Cmdline:  strings.Join(tenant.Process.Cmdline, " "),
					DNS:      newDns(tenant),
					ID:       cntrID(tenant.Process),
					Name:     tenant.Name(),
					Prefix:   "", // never prefixed: not a container.
					PID:      tenant.Process.PID,
					PIDNS:    tenant.Process.Namespaces[model.PIDNS].ID().Ino,
					PIDNSRef: pidnsID(tenant.Process.Namespaces[model.PIDNS]),
					CapBnd:   bndcaps,
					Status:   status,
					Type:     "proc",
					TypeText: "Process",
				})
			}
		}
	} else {
		// special case: either stand-alone processes nor containers, so
		// Ghostwire v1 emits a bind-mount "container" instead.
		cntrs = append(cntrs, container{
			ID:       "bmnt-" + strconv.FormatUint(n.ID().Ino, 10),
			Type:     "bindmount",
			TypeText: "Bind-mounted",
			Name:     strings.Join(n.Ref(), ":"),
		})
	}
	grps := make([]*containerGroup, 0, len(podgroups))
	for _, group := range podgroups {
		grps = append(grps, group)
	}
	nifs := make([]networkInterface, 0, len(n.Nifs))
	for _, nif := range n.Nifs {
		nifs = append(nifs, newNif(nif))
	}
	return json.Marshal(&NetworkNamespaceJSON{
		ContainerGroups:   grps,
		Containers:        cntrs,
		ID:                "netns-" + strconv.FormatUint(n.ID().Ino, 10),
		NetnsID:           n.ID().Ino,
		NetworkInterfaces: nifs,
		Routes: ipvxRoutes{
			IPv4: n.Routesv4,
			IPv6: n.Routesv6,
		},
		TransportPorts: ipvxPorts{
			IPv4: n.Portsv4,
			IPv6: n.Portsv6,
		},
		ForwardedPorts: ipvxForwardedPorts{
			IPv4: n.ForwardedPortsv4,
			IPv6: n.ForwardedPortsv6,
		},
	})
}

// containerGroup represents a Ghostwire API v1 container group. It covers only
// Kubernetes pods.
type containerGroup struct {
	ContainerIDs []string `json:"container-idrefs"`
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	TypeText     string   `json:"type-text"`
}

type container struct {
	Cmdline  string        `json:"cmdline"`
	DNS      dns           `json:"dns"`
	GroupID  string        `json:"group"`
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Prefix   string        `json:"prefix"`
	Labels   model.Labels  `json:"labels,omitempty"`
	API      string        `json:"api-path"`
	PID      model.PIDType `json:"pid"`
	PIDNS    uint64        `json:"pidns"`
	PIDNSRef string        `json:"pidns-idref"`
	CapBnd   string        `json:"capbnd,omitempty"`
	Status   string        `json:"status"`
	Type     string        `json:"type"`
	TypeText string        `json:"type-text"`
}

type dns struct {
	*network.DnsConfiguration
	EtcHosts []namedIP `json:"etc-hosts"`
}

type namedIP struct {
	Name string `json:"name"`
	IP   net.IP `json:"address"`
}

func newDns(t *network.Tenant) dns {
	etchosts := make([]namedIP, 0, len(t.DNS.Hosts))
	for name, ip := range t.DNS.Hosts {
		etchosts = append(etchosts, namedIP{
			Name: name,
			IP:   ip,
		})
	}
	return dns{
		DnsConfiguration: &t.DNS,
		EtcHosts:         etchosts,
	}
}

func v1ContainerType(typ string) string {
	if v1type, ok := containerTypesToV1[typ]; ok {
		return v1type
	}
	return typ
}

// containerTypesToV1 maps the container type identifiers used by lxkns to the
// old Ghostwire v1 API identifiers, where applicable. Lxkns container type
// identifiers usually use a domain name, such as "docker.com", instead of plain
// and unscoped identifiers, such as "docker".
var containerTypesToV1 = map[string]string{
	moby.Type:       "docker",
	containerd.Type: "containerd",
	podman.Type:     "podman",
	cri.Type:        "CRI",
}
