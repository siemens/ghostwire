// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"
)

// NetworkNamespaces maps all NetworkNamespace objects known to us.
type NetworkNamespaces map[species.NamespaceID]*NetworkNamespace

// NSID is the Linux-kernel type for network "namespace IDs", which must not be
// confused with the general network identifiers (consisting of device ID and
// inode number). NSIDs are 32bit unsigned int identifiers associated with a
// peer network namespace and can be either set explicitly or implicitly by the
// kernel when needed. NSIDs are scoped to the network namespace they are
// defined in. NSIDs cannot be changed anymore after they have been assigned for
// the first time.
type NSID uint32

// NSID_NONE is the Linux' kernel telling us that there ain't no such peer
// network namespace.
const NSID_NONE = ^NSID(0)

// NetworkNamespace represents a particular Linux network namespace together
// with its network interfaces, routes, open ports, et cetera. It additionally
// references the containers and non-container processes attached to this
// NetworkNamespace. Sets of containers (=initial process of container) as well
// as stand-alone (=non-container) processes are referred to as "tenants".
type NetworkNamespace struct {
	model.Namespace                       // discovered namespace details courtesy of lxkns.
	Nifs             map[int]Interface    // map of network interfaces by index number.
	NamedNifs        map[string]Interface // map of network interfaces indexed by name.
	Tenants          Tenants              // tenants of this network namespace (=processes/containers with additional information).
	Routesv4         []Route              // IPv4 routes
	Routesv6         []Route              // IPv6 routes
	Portsv4          []ProcessSocket      // sockets/open ports for IPv4 (including IPv6 sockets!)
	Portsv6          []ProcessSocket      // sockets/open ports for IPv6
	ForwardedPortsv4 []ForwardedPort      // IPv4 ports forwarded into other network namespaces
	ForwardedPortsv6 []ForwardedPort      // IPv6 ports forwarded into other network namespaces

	peerNetns map[NSID]*NetworkNamespace // NSID-to-network namespace map; required for resolving netlink relations.
}

// NetworkNamespaceList contains NetworkNamespace elements, and optionally can
// be sorted in-place using its Sort method.
type NetworkNamespaceList []*NetworkNamespace

// NewNetworkNamespace returns a new NetworkNamespace, based on the lxkns
// namespace discover information. The new NetworkNamespace will have its
// network interfaces and routes discovered. However, the relations between
// network interfaces remain unresolved at this time.
func NewNetworkNamespace(netns model.Namespace, tenantProcs []*model.Process) *NetworkNamespace {
	nns := &NetworkNamespace{
		Namespace: netns,
		Nifs:      map[int]Interface{},
		peerNetns: map[NSID]*NetworkNamespace{},
		NamedNifs: map[string]Interface{},
	}
	// Get an RTNETLINK socket wired up to this particular network namespace.
	nlh, err := nns.OpenNetlink()
	if err != nil {
		log.Warnf("cannot discover inside net:[%d], reason: %s",
			netns.ID().Ino, err.Error())
		return nil
	}
	defer nlh.Close()
	// Discover the network interfaces in this network namespace, together with
	// the addresses assigned to these network interfaces.
	nns.discoverNetworkInterfaces(nlh)
	// Index the discovered network interfaces by their names
	for _, nif := range nns.Nifs {
		nns.NamedNifs[nif.Nif().Name] = nif
	}
	// Routes
	nns.Routesv4 = nns.discoverRoutes(nlh, unix.AF_INET)
	nns.Routesv6 = nns.discoverRoutes(nlh, unix.AF_INET6)
	// Gather DNS-related information, et cetera, for the tenant processes (that
	// is, network namespace leader processes and container processes) turning
	// them into "tenants".
	tenants := make(Tenants, 0, len(tenantProcs))
	for _, proc := range tenantProcs {
		tenants = append(tenants, NewTenant(proc))
	}
	nns.Tenants = tenants

	return nns
}

// DisplayName returns a "simplified" name for "simple" display use cases, where
// it is desirable to identify network namespaces by the names of their tenants
// (containers and stand-alone processes). In case of multiple tenants, the
// display name will be the lexicographically first tenant name, or the name of
// the process with PID 1, if present.
func (n *NetworkNamespace) DisplayName() string {
	tenants := n.Tenants
	if len(tenants) == 0 {
		return n.Ref().String()
	}
	tenants.Sort()
	var name string
	proc := tenants[0].Process
	if c := proc.Container; c != nil {
		name = fmt.Sprintf("📦 %s ◉ %.8s…", c.Name, c.Engine.ID)
	} else {
		name = fmt.Sprintf("⚙️  %s(%d)", proc.Name, proc.PID)
	}
	if len(tenants) > 1 {
		return name + ", …"
	}
	return name
}

// NifList returns the (unordered) list of NetworkInterfaces in this network
// namespace.
func (n *NetworkNamespace) NifList() Interfaces {
	nifs := make(Interfaces, 0, len(n.Nifs))
	for _, nif := range n.Nifs {
		nifs = append(nifs, nif)
	}
	return nifs
}

// NifsString returns the display name of this network namespace together with
// the names of its network interface (and interface indices).
func (n *NetworkNamespace) NifsString() string {
	names := make([]string, 0, len(n.Nifs))
	for index, nif := range n.Nifs {
		names = append(names, fmt.Sprintf("%s(%d)", nif.Nif().Name, index))
	}
	return n.DisplayName() + ": " + strings.Join(names, ", ")
}

// related returns the NetworkNamespace identified by nsid in this network
// namespace, or nil. This method is used internally after all network
// namespaces with their network interfaces have been discovered and only then
// can relations between network interfaces in different network namespaces be
// properly resolved.
func (n *NetworkNamespace) related(nsid NSID) *NetworkNamespace {
	return n.peerNetns[nsid]
}

// discoverTransportPorts discovers the open (transport-layer) ports in this
// network namespace. Additionally, it relates the ports to the network
// interfaces carrying their traffic. And it notifies about IPv6 sockets
// handling IPv4 traffic.
func (n *NetworkNamespace) discoverTransportPorts(sm socketToProcessMap, allprocs model.ProcessTable) {
	// Without any attached processes, we don't have any /proc/[...]/net/xxx to
	// read the socket information for this network namespace from.
	//
	// FIXME: use mountineer, as other processes might have open fds, yet not
	// attached to this network namespace.
	if n.Ealdorman() == nil {
		return
	}
	// Now discover and process TCP und UDP sockets, for both IPv4 and IPv6.
	// Note our special handling of IPv6 sockets covering IPv4 traffic...
	addrToNifs := n.newAddrToNifMap()
	for _, proto := range []int{syscall.IPPROTO_TCP, syscall.IPPROTO_UDP} {
		// v4Ports indicates the IPv4 addresses we've seen for this protocol, so
		// that we don't create any IPv4-"mapped" pseudo entries when there are
		// both an IPv4 as well as an IPv6 socket covering AF_INET and AF_INET6,
		// and the IPv6 socket would also include the IPv4 mapped address range.
		v4Ports := map[comparableAddressPort]struct{}{}
		for _, af := range []int{unix.AF_INET, unix.AF_INET6} {
			ports := discoverSockets("/proc", n.Ealdorman().PID, af, proto, sm)
			for idx, port := range ports {
				procs := allprocs.ProcessesByPIDs(port.PIDs...)
				ports[idx].Processes = procs
				// Danger, Robinson, Danger: we're Dual Stack on Linux... ...and
				// that means that IPv4 network traffic can be handled via IPv6
				// family sockets(!), if there is no matching IPv4 family
				// socket, but a matching(!) IPv6 socket. We're not talking raw
				// sockets here, but ordinary transport-layer sockets.
				//
				// A good principle discussion can be found here:
				// https://security.stackexchange.com/questions/92081/is-receiving-ipv4-connections-on-af-inet6-sockets-insecure
				//
				// For the API, especially IPV6_V6ONLY, see here:
				// https://stackoverflow.com/questions/1618240/how-to-support-both-ipv4-and-ipv6-connections
				//
				// So we make users aware of this fact in that if we see an
				// anyaddr IPv6 socket on a port for which no matching separate
				// anyaddr IPv4 socket exists, we add one to the list by
				// ourselves.
				if af == unix.AF_INET {
					// On the first run, that always is the IPv4 run, we need to
					// remember what is already covered ... and thus visible to
					// IPv4-centric users (including "us").
					v4Ports[comparableAddressPort{addr: string(port.LocalIP), port: port.LocalPort}] = struct{}{}
				} else if (port.LocalIP.IsUnspecified() || port.LocalIP.To4() != nil) &&
					(port.RemoteIP.IsUnspecified() || port.RemoteIP.To4() != nil) {
					// Second run is always the IPv6 run: if there's no matching
					// IPv4 port but aliasing would occur, then add an (alias)
					// entry back to the IPv4 port list. This aliasing can
					// happen for both the unspecified address ("anyaddr") and
					// the IP4-mapped IPv6 addresses (::ffff:0/96).
					//
					// Gotcha: we must assume here that the IPV6_ONLY socket
					// option defaults to be *disabled*. At this time we have no
					// means to discover this socket option for a specific port.
					ipv4addr := net.IP([]byte{0, 0, 0, 0}) // DO NOT USE net.IPv4zero --> IPv4 mapped IPv6!
					if ipv4 := port.LocalIP.To4(); ipv4 != nil {
						ipv4addr = ipv4
					}
					remipv4addr := net.IP([]byte{0, 0, 0, 0}) // DO NOT USE net.IPv4zero --> IPv4 mapped IPv6!
					if remipv4 := port.RemoteIP.To4(); remipv4 != nil {
						remipv4addr = remipv4
					}
					if _, ok := v4Ports[comparableAddressPort{addr: string(ipv4addr), port: port.LocalPort}]; !ok {
						nifs := addrToNifs[string(ipv4addr)]
						n.Portsv4 = append(n.Portsv4, ProcessSocket{
							Family:          unix.AF_INET,
							Protocol:        Protocol(proto),
							LocalIP:         ipv4addr,
							LocalPort:       port.LocalPort,
							RemoteIP:        remipv4addr,
							RemotePort:      port.RemotePort,
							State:           port.State,
							SimplifiedState: port.SimplifiedState,
							PIDs:            port.PIDs,
							IPv4Mapped:      true,
							Nifs:            nifs,
							Processes:       procs,
						})
					}
				}
				// Resolve the related/"concerned" network interface(s)...
				localip := port.LocalIP
				if ipv4 := localip.To4(); ipv4 != nil {
					localip = ipv4
				}
				if nifs, ok := addrToNifs[string(localip)]; ok {
					ports[idx].Nifs = nifs // !!! "port" is a copy, not a pointer, hmpf.
				}
			}
			if af == unix.AF_INET {
				n.Portsv4 = append(n.Portsv4, ports...)
			} else {
				n.Portsv6 = append(n.Portsv6, ports...)
			}
		}
	}
}

// newAddrToNifMap returns a new map from IP addresses (including the
// unspecified addresses) assigned to network interfaces to the corresponding
// network interface(s). The map is already fully populated.
func (n *NetworkNamespace) newAddrToNifMap() map[string][]Interface {
	// Please note: while net.IP (=[]byte) is un-keyable, string is... :o ...and
	// at least these cases of conversion from []byte to string are cheap and
	// optimized by the Go compiler.
	addrToNifs := map[string][]Interface{}
	allNifs := n.NifList()
	addrToNifs[string(net.IPv4zero.To4())] = allNifs
	addrToNifs[string(net.IPv6unspecified)] = allNifs

	registerNifs := func(nif Interface, addrs Addresses) {
		for _, addr := range addrs {
			addrToNifs[string(addr.Address)] = append(addrToNifs[string(addr.Address)], nif)
		}
	}

	for _, nif := range allNifs {
		registerNifs(nif, nif.Nif().Addrsv4)
		registerNifs(nif, nif.Nif().Addrsv6)
	}

	return addrToNifs
}

// comparableAddressPort is a comparable tuple of a net.IP address (but casted
// into a string!) and a transport port. We cannot use net.IP directly, as its
// underlying []byte is deemed un-comparable by the Prophets of Go. But just
// casting []byte into string makes the data comparable, byte by byte. Oh, well.
type comparableAddressPort struct {
	addr string // ouch: incomparable map key when net.IP
	port uint16
}

// discoverNSID discovers the NSIDs of the peer network namespaces related to
// our network namespace. As this part of the discovery is run only after we
// built the full netns map, we need to create our own netlink handle here, so
// we don't expect one passed onto us.
func (n *NetworkNamespace) discoverNSIDs(allnetns NetworkNamespaces) {
	// The rtnetlink API really is stupid, because it's of the "oracle" sort
	// (not the one with a capital "O" but with a lower case "o"). It only
	// answers with a NSID or "no(pe)" when we ask it for a specific network
	// namespace. It doesn't allow simply listing all the NSIDs, not even for a
	// single network namespace. Ouch.
	nlh, err := n.OpenNetlink()
	if err != nil {
		log.Errorf("cannot discover NSIDs: %s", err.Error())
		return
	}
	defer nlh.Close()
	for _, peerNetns := range allnetns {
		if peerNetns == n.Namespace {
			continue
		}
		ref := peerNetns.Ref()
		if len(ref) == 0 {
			continue
		}
		netnspath := ref[len(ref)-1]
		var mntneer *mountineer.Mountineer
		if len(ref) > 1 {
			mntneer, err = mountineer.New(ref[:len(ref)-1], nil)
			if err != nil {
				continue
			}
			netnspath, err = mntneer.Resolve(netnspath)
			if err != nil {
				continue
			}
		}
		netnsfd, closer, err := ops.NewTypedNamespacePath(netnspath, species.CLONE_NEWNET).NsFd()
		if mntneer != nil {
			mntneer.Close()
		}
		if err != nil {
			log.Errorf("cannot access peer netns %s", ref)
			continue
		}
		nsid, err := nlh.GetNetNsIdByFd(netnsfd)
		if err != nil {
			log.Errorf("cannot determine NSID of peer netns %s, error: %s", ref, err.Error())
			continue
		}
		closer()
		if NSID(nsid) == NSID_NONE {
			continue
		}
		n.peerNetns[NSID(nsid)] = peerNetns
		log.Debugf("net:[%d] NSID %d ↦ net:[%d] %s", n.ID().Ino, nsid, peerNetns.ID().Ino, peerNetns.DisplayName())
	}
}

// OpenEthtool open a socket in this network namespace that is suitable for
// ethtool-related ioctl's. Somewhat deviating from the standard Go error return
// pattern, it doesn't return a zero fd on error, but a -1 fd ... which is
// guaranteed to be an invalid fd, whereas 0 most probably is valid.
func (n *NetworkNamespace) OpenEthtool() (fd int, err error) {
	fd = -1
	if err := n.OpenInNetworkNamespace(func() error {
		var err error
		fd, err = unix.Socket(unix.AF_INET, unix.SOCK_DGRAM, unix.IPPROTO_IP)
		return err
	}); err != nil {
		// Note: this cannot be placed inside the function passed to
		// OpenInNetworkNamespace, because we here must deal with the situation
		// where switching back the namespace failed and in this case properly
		// dispose of the socket fd now connected to the wrong network
		// namespace.
		if fd >= 0 {
			_ = unix.Close(fd)
		}
		return -1, err
	}
	return fd, nil
}

// OpenNetlink returns a netlink (route) handle for netlink queries and
// operations on this network namespace. The caller is responsible for releasing
// the resources associated with the returned netlink handle by calling
// Delete(!) on it when not needed anymore.
func (n *NetworkNamespace) OpenNetlink() (*netlink.Handle, error) {
	var nlHandle *netlink.Handle
	if err := n.OpenInNetworkNamespace(func() error {
		var err error
		nlHandle, err = netlink.NewHandle(unix.NETLINK_ROUTE)
		return err
	}); err != nil {
		if nlHandle != nil {
			nlHandle.Close() // safety net
		}
		return nil, err
	}
	return nlHandle, nil
}

// OpenInNetworkNamespace calls the supplied opener function in the context of
// this network namespace.
func (n *NetworkNamespace) OpenInNetworkNamespace(opener func() error) error {
	ref := n.Namespace.Ref()
	if len(ref) == 0 {
		return fmt.Errorf("invalid empty netns reference")
	}
	// Do we have a single network namespace path or are there multiple paths,
	// where the last path references a bind-mounted network namespace in some
	// mount namespace?
	netnspath := ref[len(ref)-1]
	if len(ref) > 1 {
		mntneer, err := mountineer.New(ref[:len(ref)-1], nil)
		if err != nil {
			return fmt.Errorf("invalid netns reference %s: %s", ref.String(), err.Error())
		}
		// we need to keep the mount namespace open only until we've got the
		// netlink handle, that is, until the end of this method.
		defer mntneer.Close()
		netnspath, err = mntneer.Resolve(netnspath)
		if err != nil {
			return fmt.Errorf("invalid netns reference %s: %s", ref.String(), err.Error())
		}
	}
	// Shortcut route in case it's our own network namespace for those running
	// the Gostwire discovery engine without sufficient capabilities, so trying
	// to switch into our own current network namespace would otherwise fail.
	// This of course begs the question, why would we do to switch into the same
	// network namespace anyway? Because.
	targetnetnsid, targeterr := ops.NamespacePath(netnspath).ID()
	if gwnetnserr == nil && targeterr == nil && gwnetnsid == targetnetnsid {
		return opener()
	}
	// Normal route where we need to temporarily switch into the target network
	// namespace in order to open a file descriptor in that context.
	var openerErr error
	visitErr := ops.Visit(func() {
		openerErr = opener()
	}, ops.NamespacePath(netnspath))
	if visitErr != nil {
		return visitErr
	}
	return openerErr
}

var (
	gwnetnsid  species.NamespaceID // our own network namespace ID when starting.
	gwnetnserr error               // there should not be any, except for a botched procfs.
)

// Fetch the ID of network namespace we (=main Go routine) are running in for
// once and all. All Go routines normally run attached to this network
// namespace, except for short and carefully locked "critical" sections.
func init() {
	gwnetnsid, gwnetnserr = ops.NamespacePath("/proc/self/ns/net").ID()
}

// discoverNetworkInterfaces discovers the network interfaces in this network
// namespace.
func (n *NetworkNamespace) discoverNetworkInterfaces(nlh *netlink.Handle) {
	links, err := nlh.LinkList()
	if err != nil {
		return
	}
	// We only open the ethtool API socket on demand when we find actual
	// physical network interfaces. This way, we don't need to switch network
	// namespaces and open the socket for containers without any physical
	// (SR-IOV VF) network interface(s) present.
	physNifsPresent := false
	ethtoolFd := -1 // optional AF_INET+SOCK_DGRAM+IPPROTO_IP socket, opened on the fly
	for _, link := range links {
		nif := NewInterface(nlh, n, link)
		n.Nifs[nif.Nif().Index] = nif
		// If this is isn't a physical network interface, then it won't have a
		// bus address anyway.
		if !nif.Nif().Physical {
			continue
		}
		// It's a physical network interface and if we haven't done so already,
		// we now open the ethtool API in this network namespace.
		if !physNifsPresent {
			var err error
			if ethtoolFd, err = n.OpenEthtool(); err == nil { // TODO: improve error handling
				physNifsPresent = true
				defer unix.Close(ethtoolFd)
			}
		}
		nif.Nif().discoverBusAddress(ethtoolFd)
	}
}

// NewNetworkNamespaces takes a set of discovered network namespaces and creates
// the Gostwire-specific NetworkNamespace objects wrapping them and supplying
// network layer-related information not discovered by lxkns.
func NewNetworkNamespaces(
	allnetns model.NamespaceMap,
	allprocs model.ProcessTable,
	containers model.Containers,
) NetworkNamespaces {
	// In order to later figure out the tenants of a network namespace, we need
	// not only to take network namespace leader processes into account but also
	// container processes, as containers can perfectly share network namespaces
	// in a configuration where the container processes aren't the namespace
	// leader processes, such as in container-in-container configurations.
	tenantProcsByNetns := tenantsOfNetnsMap(allnetns, containers)
	// Now work on the network namespaces and discover their topology and
	// configuration.
	netspaces := NetworkNamespaces{}
	for netnsid, netns := range allnetns {
		if netwns := NewNetworkNamespace(netns, tenantProcsByNetns[netns]); netwns != nil {
			netspaces[netnsid] = netwns
		}
	}
	soxProcsMap := discoverAllSockInodes("/proc")
	for nsid, netns := range netspaces {
		log.Debugf("discovering details of net:[%d]...", nsid.Ino)
		netns.discoverNSIDs(netspaces)
		netns.discoverTransportPorts(soxProcsMap, allprocs)
		netns.discoverForwardedPorts()
		log.Debugfn(func() string {
			nifNames := make([]string, 0, len(netns.Nifs))
			for _, nif := range netns.Nifs {
				name := nif.Nif().Name
				if alias := nif.Nif().Alias; alias != "" {
					name += fmt.Sprintf("(~%s)", alias)
				}
				nifNames = append(nifNames, name)
			}
			return "found nifs: " + strings.Join(nifNames, ", ")
		})
	}
	// Resolve the network interfaces topology, except for SR-IOV PFs/VFs. In
	// the case of SR-IOV we first only build a map of the discovered PFs and
	// VFs. This map indexes bus addresses to their corresponding interface
	// objects.
	for _, netns := range netspaces {
		for _, nif := range netns.Nifs {
			nif.(resolver).ResolveRelations(netspaces)
		}
	}
	// Finally resolve the SR-IOV PF/VF topology, based on the bus addresses
	// seen. Unfortunately, RTNETLINK doesn't give any netdev topology
	// information.
	resolveSRIOVTopology(netspaces)

	completeForwardedPortInformation(netspaces)

	return netspaces
}

// tenantsOfNetnsMap maps network namespaces to its tenants, that is, a list of
// processes currently attached to this network namespace.
func tenantsOfNetnsMap(
	allnetns model.NamespaceMap,
	containers model.Containers,
) map[model.Namespace][]*model.Process {
	tenantProcsByNetns := map[model.Namespace][]*model.Process{}
	for _, container := range containers {
		if container.Process != nil {
			tenantProcsByNetns[container.Process.Namespaces[model.NetNS]] =
				append(tenantProcsByNetns[container.Process.Namespaces[model.NetNS]],
					container.Process)
		}
	}
	for _, netns := range allnetns {
		tenantProcs := tenantProcsByNetns[netns]
	nextleader:
		for _, leader := range netns.Leaders() {
			for _, tenant := range tenantProcs {
				if tenant == leader {
					continue nextleader
				}
			}
			tenantProcs = append(tenantProcs, leader)
		}
		tenantProcsByNetns[netns] = tenantProcs
	}
	return tenantProcsByNetns
}

// resolveSRIOVTopology resolves the PF-VF topology, given a map with PF and VF network
// interfaces indexed by their (PCI) bus addresses.
func resolveSRIOVTopology(netspaces NetworkNamespaces) {
	sriovNifs := map[string]Interface{}
	for _, netns := range netspaces {
		for _, nif := range netns.Nifs {
			busAddr := nif.Nif().BusAddr
			if busAddr == "" {
				continue
			}
			sriovNifs[busAddr] = nif
		}
	}
	for _, nif := range sriovNifs {
		// Is this a VF? Then it has to have a physfn directory, which actually
		// is a link to the PF's device directory.
		sysfsBusPath := nif.Nif().SysfsBusPath()
		physfnLink, err := os.Readlink(sysfsBusPath + "/physfn")
		if err == nil {
			nif.Nif().SRIOVRole = PCI_SRIOV_VF
			// physfn is a symbolic link to the VF sibling in the flat
			// /sys/bus/pci/devices/ virtual structure.
			pfBusAddr := filepath.Base(physfnLink)
			pfnif, ok := sriovNifs[pfBusAddr]
			if !ok {
				continue
			}
			nif.Nif().PF = pfnif
			pfnif.Nif().Slaves = append(pfnif.Nif().Slaves, nif.Interface())
			log.Debugf("PF %s net:[%d] ↔ VF %s net:[%d]",
				pfnif.Nif().Name, pfnif.Nif().Netns.ID().Ino,
				nif.Nif().Name, nif.Nif().Netns.ID().Ino)
			continue
		}
		// Is this a PF? Then it has to have some PF-specific device nodes with
		// names starting with "sriov_".
		_, err = os.Stat(sysfsBusPath + "/sriov_numvfs")
		if err != nil {
			continue
		}
		// Nothing more to do here than to mark this as a PF, but the topology
		// will be set up only whenever we see a VF.
		nif.Nif().SRIOVRole = PCI_SRIOV_PF
	}
}

// Sorted returns the (stable) sorted list of NetworkNamespace elements.
func (m NetworkNamespaces) Sorted() NetworkNamespaceList {
	allnetns := make(NetworkNamespaceList, 0, len(m))
	for _, netns := range m {
		allnetns = append(allnetns, netns)
	}
	allnetns.Sort()
	return allnetns
}

// ByProcess looks up the NetworkNamespace for the given model.Process. If not
// found, returns nil.
func (m NetworkNamespaces) ByProcess(proc *model.Process) *NetworkNamespace {
	lxknsNetns := proc.Namespaces[model.NetNS]
	if lxknsNetns == nil {
		return nil
	}
	return m[lxknsNetns.ID()]
}

// ByContainer looks up the NetworkNamespace for the given model.Container. If
// not found, returns nil.
func (m NetworkNamespaces) ByContainer(cntr *model.Container) *NetworkNamespace {
	return m.ByProcess(cntr.Process)
}

// String returns a textual representation of the network namespaces map
// consisting of the (sorted) network namespaces' display names.
func (m NetworkNamespaces) String() string {
	names := make([]string, 0, len(m))
	for _, netns := range m {
		names = append(names, netns.DisplayName())
	}
	return strings.Join(names, " // ")
}

// Sort (stably) sorts the list of NetworkNamespace elements in-place. Please
// note that the network namespace with the initial process PID 1 will always
// come first.
func (n NetworkNamespaceList) Sort() {
	sort.SliceStable(n, orderNetworkNamespaces(n))
}

// orderNetworkNamespaces compares two NetworkNamespaces and returns true, if
// the first is lexicographically before the second.
func orderNetworkNamespaces(netns []*NetworkNamespace) func(a, b int) bool {
	return func(a, b int) bool {
		// A network namespace housing PID 1 always comes first. MINNSGA!
		// MINNSGA!!!
		initA := false
		procA := netns[a].Ealdorman()
		if procA != nil && procA.PID == 1 {
			initA = true
		}
		initB := false
		procB := netns[b].Ealdorman()
		if procB != nil && procB.PID == 1 {
			initB = true
		}
		if initA || initB {
			return initA && !initB // ..."initA < initB"
		}
		return orderingNetnsName(netns[a]) < orderingNetnsName(netns[b])
	}
}

// orderingNetnsName returns a synthetic name for the given NetworkNamespace,
// based on the lexicographically first tenant names (or bindmount reference
// name if we don't have tenants).
func orderingNetnsName(netns *NetworkNamespace) string {
	leaders := netns.Leaders()
	if len(leaders) == 0 {
		return netns.Ref().String()
	}
	names := make([]string, 0, len(leaders))
	for _, leader := range leaders {
		if c := leader.Container; c != nil {
			names = append(names, c.Name)
			continue
		}
		names = append(names, leader.Name)
	}
	sort.Strings(names)
	return names[0]
}
