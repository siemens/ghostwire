// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"net"
	"sort"
	"strings"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// Interface represents Gostwire's perspective on the common properties
// of a network interface inside a network namespace. For some (virtual) type of
// network interfaces, additional nif type-specific interfaces give more
// details, such as the Peer interface about the peer relations between the two
// VETH ends of an VETH "wire".
//
// Due to the design of Go interfaces versus (pointer) receivers we have to deal
// with the ugly situation were in a (pointer) receiver method with a *NifAttrs
// receiver and we need the Interface instead to whatever "derived" type
// like VethAttrs is implementing it. The Interface() method is a
// convience to encapsulate our work-around that looks up the original
// Interface from the nif list of the owning NetworkNamespace.
type Interface interface {
	Nif() *NifAttrs       // returns the common network interface attributes.
	Interface() Interface // returns the NetworkInterface interface for this, erm, network interface.
}

// Interfaces is a list of network Interface elements, and can optionally be
// sorted in-place.
type Interfaces []Interface

// NifAttrs defines Gostwire's view on the common attributes of any network
// interface. This also includes some relations between network interfaces.
type NifAttrs struct {
	Netns       *NetworkNamespace // network namespace this interface belongs to.
	Kind        string            // kind of interface.
	Name        string            // interface name.
	Alias       string            // alias name.
	Index       int               // interface index.
	State       OperState         // operational state.
	Physical    bool              // or more metaphorical: it has an associated driver.
	Promiscuous bool              // does snoop all traffic?
	Labels      model.Labels      // optional labels attached by Gostwire decorators.
	L2Addr      net.HardwareAddr  // data-link layer (aka "hardware") address.
	Addrsv4     Addresses         // assigned IPv4 network addresses.
	Addrsv6     Addresses         // assigned IPv6 network addresses.
	SRIOVRole   SRIOVRole         // ...when network interface is an SR-IOV PF or VF.

	// Relations with other network interfaces
	Bridge Interface  // when interface is a "port" of a bridge interface.
	Slaves Interfaces // MACVLANs, VXLANs, VFs, others (but not VETH peers).
	PF     Interface  // when interface is an SR-IOV VF.

	// Low-level, not available after unmarshalling.
	Link    netlink.Link // low-level netlink information about this interface.
	BusAddr string
}

// SRIOVRole identifies network interfaces that are SR-IOV PFs or VFs.
type SRIOVRole uint8

const (
	PCI_NIC      SRIOVRole = iota // Not a PCI NIC or SR-IOV isn't enabled.
	PCI_SRIOV_PF                  // network interface is a PCI PF.
	PCI_SRIOV_VF                  // network interface is a PCI VF.
)

// NifMaker returns an instance of something implementing network.Interface.
type NifMaker func() Interface

// resolver instructs a network Interface to resolve its relations with other
// network interfaces during the discovery phase when all NetworkNamespaces and
// their network interfaces have been discovered.
type resolver interface {
	ResolveRelations(NetworkNamespaces) // resolve relations between network interfaces
}

// initializer instructs a network Interface to properly initialize (set up)
// itself from the information given.
type initializer interface {
	Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) // initializes a network interface-based type
}

var _ Interface = (*NifAttrs)(nil)
var _ resolver = (*NifAttrs)(nil)
var _ initializer = (*NifAttrs)(nil)

// Nif returns the common network interface attributes.
func (n *NifAttrs) Nif() *NifAttrs { return n }

// NewInterface returns a new network.Interface that is Gostwire's view take on
// network interfaces in network namespaces. The network interface returned has
// the correct underlying Go type (such as BridgeAttrs, VethAttrs, ...) based on
// the specific kind of link. If there is no dedicated Gostwire type, then the
// generic NifAttr type is created and returned instead.
func NewInterface(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) Interface {
	if len(nifMakers) == 0 {
		collectNifMakers()
	}
	var nif Interface
	if nifmaker, ok := nifMakers[link.Type()]; ok {
		nif = nifmaker()
	} else {
		nif = &NifAttrs{}
	}
	nif.(initializer).Init(nlh, netns, link)
	return nif
}

// collectNifMakers populates the map with the NifMakers for the particular kind
// of network interface.
func collectNifMakers() {
	// We first need to get and cache the registered kind-specific NifMakers; we
	// can't do so in an init(), as we cannot be sure that all registrations
	// have been done when that init() would run.
	for _, nm := range plugger.Group[NifMaker]().PluginsSymbols() {
		nifMakers[nm.Plugin] = nm.S
	}
	log.Infof("available specialized network interface plugins for kinds: %s",
		strings.Join(plugger.Group[NifMaker]().Plugins(), ", "))
}

// nifMakers maps link type names to their "maker" (factories) functions.
var nifMakers = map[string]NifMaker{}

// Interface returns the network.Interface for this network.Interface (sic!).
// This is specific to Go, due to the Go design with its interfaces versus
// (pointer) receivers.
func (n *NifAttrs) Interface() Interface {
	return n.Netns.Nifs[n.Index]
}

// Init initializes this Nif from information specified in the NetworkNamespace
// and lots of netlink.Link information.
func (n *NifAttrs) Init(nlh *netlink.Handle, netns *NetworkNamespace, link netlink.Link) {
	attrs := link.Attrs()
	l2addr := attrs.HardwareAddr
	if len(l2addr) == 0 {
		l2addr = []byte{0, 0, 0, 0, 0, 0}
	}
	kind := link.Type()
	if kind == "device" {
		kind = ""
	}
	// Get the IP addresses assigned to this network interface.
	addrsv4 := []Address{}
	addrsv6 := []Address{}
	nlAddrs, err := nlh.AddrList(link, 0) // ...include all address families
	if err == nil {
		for _, addr := range nlAddrs {
			var family int
			switch len(addr.IP) {
			case 4:
				family = unix.AF_INET
			case 16:
				family = unix.AF_INET6
			default:
				continue
			}
			prefixlen, _ := addr.Mask.Size()
			a := Address{
				Family:            family,
				Address:           addr.IP,
				PrefixLength:      uint(prefixlen),
				Scope:             addr.Scope,
				Index:             addr.LinkIndex,
				PreferredLifetime: uint32(addr.PreferedLft),
				ValidLifetime:     uint32(addr.ValidLft),
			}
			if family == unix.AF_INET6 {
				addrsv6 = append(addrsv6, a)
			} else {
				addrsv4 = append(addrsv4, a)
			}
		}
	}
	// Final base initialization.
	*n = NifAttrs{
		Netns:       netns,
		Kind:        kind,
		Name:        attrs.Name,
		Alias:       attrs.Alias,
		Index:       attrs.Index,
		State:       OperState(attrs.OperState),
		Physical:    link.Type() == "device" && (attrs.Flags&net.FlagLoopback == 0),
		Promiscuous: attrs.Flags&unix.IFF_PROMISC != 0,
		Labels:      model.Labels{},
		L2Addr:      l2addr,
		Addrsv4:     addrsv4,
		Addrsv6:     addrsv6,
		Link:        link,
	}
}

// HasAddress returns true if the specified IP address is one of the addresses
// assigned to this interface, else false. The specified address is expected to
// be in canonical format, that is of length 4 for an IPv4 address (and being an
// IPv4-mapped IPv6 address).
func (n *NifAttrs) HasAddress(ip net.IP) bool {
	if len(ip) == net.IPv6len {
		for _, addr := range n.Addrsv6 {
			if ip.Equal(addr.Address) {
				return true
			}
		}
		return false
	}
	for _, addr := range n.Addrsv4 {
		if ip.Equal(addr.Address) {
			return true
		}
	}
	return false
}

// DiscoverPhys discovers the bus address of physical network interfaces and
// updates the interface's BusAddr field. We don't do bus address discovery in
// Init because we don't want to allocate "private" local ethtool API file
// descriptors over and over again individually for each physical network
// interface. Instead, NetworkNamespace.discoverNetworkInterfaces allocates an
// ethtool API file descriptor only once per network namespace, and only when
// phyiscal network interfaces are present in a network namespace.
func (n *NifAttrs) discoverBusAddress(ethtoolFd int) {
	driverInfo, err := unix.IoctlGetEthtoolDrvinfo(ethtoolFd, n.Name)
	if err != nil {
		log.Errorf("cannot query ethtool API driver information for nif %q, reason: %s", n.Name, err.Error())
		return
	}
	// Fun fact: the ethtool API returns "N/A" for the bus address when trying
	// to use it on a virtual network interface :)
	n.BusAddr = strings.TrimRight(string(driverInfo.Bus_info[:]), "\x00")
	log.Debugf("physical network interface %s has device bus address: %s", n.Name, n.BusAddr)
}

// SysfsBusPath returns a device directory path for the physical device of this
// network interface somewhere deeper inside /sys/bus/.
//
// Note: this is currently needed only for PCI(e) devices and thus does not work
// correctly for other busses, such as USB.
func (n *NifAttrs) SysfsBusPath() string {
	return "/sys/bus/pci/devices/" + n.BusAddr
}

// ResolveRelations resolves relations to other network interfaces. Please note
// that this doesn't include the PF-VF topology, as we're to resolve that
// topology separately.
func (n *NifAttrs) ResolveRelations(allns NetworkNamespaces) {
	// Could this be a bridge "port" interface? Its bridge can only be in the
	// same network namespace.
	idx := n.Link.Attrs().MasterIndex
	if idx != 0 {
		if bridge := n.Netns.Nifs[idx]; bridge != nil && bridge.Nif().Kind == "bridge" {
			n.Bridge = bridge
			brattrs := bridge.(*BridgeAttrs)
			// Go AWAY, that's flawed object-oriented design! Because we're here
			// *NifAttrs, we're thus not network.Interface anymore. And
			// therefore we can't simply "cast" back from *NifAttrs to
			// network.Interface, because a network.Interface pointer actually
			// now says: "I'm a *NifAttrs satisfying network.Interface". It has
			// forgotten what ever original type it was that embedded the
			// NifAttrs. Oh, bummer.
			brattrs.Ports = append(brattrs.Ports, n.Interface())
		} else if bridge == nil {
			log.Warnf("missing bridge network interface idx %d", idx)
		} else {
			log.Warnf("master network interface is not a bridge, but of type '%s'", bridge.Nif().Kind)
		}
	}
}

// AddLabels adds in (merges) the passed labels with the existing labels
// assigned to this network interface. Added labels take precedence over
// existing labels, replacing them in case of conflict.
func (n *NifAttrs) AddLabels(labels model.Labels) {
	for key, val := range labels {
		n.Labels[key] = val
	}
}

// Sort sorts the list of Interface elements in-place.
func (i Interfaces) Sort() {
	sort.SliceStable(i, func(a, b int) bool {
		nameA := i[a].Nif().Name
		nameB := i[b].Nif().Name
		if alo := nameA == "lo"; alo || nameB == "lo" {
			return alo
		}
		return nameA < nameB
	})
}

// OfKind returns only the Interfaces of the specified kind.
func (i Interfaces) OfKind(kind string) Interfaces {
	kindifs := make(Interfaces, 0, len(i))
	for _, nif := range i {
		if nif.Nif().Kind != kind {
			continue
		}
		kindifs = append(kindifs, nif)
	}
	return kindifs
}
