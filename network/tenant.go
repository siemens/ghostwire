// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/lxkns/species"
)

// Tenant represents a process or even a container together with its
// application-layer specific communication configuration, namely configuration
// of the name resolution/DNS.
type Tenant struct {
	Process      *model.Process   // associated ealdorman process
	BoundingCaps []byte           // bounding capabilities in form of a byte string with lsb being bit 0 of the last byte.
	DNS          DnsConfiguration // DNS and name resolution configuration
}

// DnsConfiguration contains DNS/name resolution-related configuration
// information.
type DnsConfiguration struct {
	Hostname      string            `json:"uts-hostname"` // nodename returned by uname(2) and usually used instead of the gethostname() syscall.
	EtcHostname   string            `json:"etc-hostname"` // host name read from /etc/hostname, if present.
	EtcDomainname string            `json:"domainname"`   // domain name read from /etc/domainname, if present.
	Hosts         map[string]net.IP `json:"etc-hosts"`    // (host)name to IP address mapping, as read from /etc/hosts.
	Nameservers   []net.IP          `json:"nameservers"`  // list of name server IP addresses.
	Searchlist    []string          `json:"searchlist"`   // list of domain names to use in resolving names to IP addresses.
}

// Tenants is a list of Tenant elements; it allows for in-place sorting using
// its Sort method.
type Tenants []*Tenant

// NewTenant returns a new Tenant corresponding with the specified process (and
// its optional container) and with its DNS configuration discovered.
func NewTenant(proc *model.Process) *Tenant {
	t := &Tenant{Process: proc}
	t.BoundingCaps = t.caps("CapBnd")
	tenantfs, err := mountineer.New(proc.Namespaces[model.MountNS].Ref(), nil)
	if err != nil {
		return t
	}
	defer tenantfs.Close()

	t.DNS.EtcHostname = t.readSingleLine(tenantfs, "/etc/hostname")
	t.DNS.EtcDomainname = t.readSingleLine(tenantfs, "/etc/domainname")
	t.DNS.Hostname = t.uname()
	t.DNS.Hosts = t.readHosts(tenantfs)
	t.DNS.Nameservers, t.DNS.Searchlist = t.readResolvConf(tenantfs)

	return t
}

// Name returns the tenant's name, which can be either its container name,
// falling back to its process name (with PID), if there's no associated
// container.
func (t *Tenant) Name() string {
	if c := t.Process.Container; c != nil {
		return c.Name
	}
	return fmt.Sprintf("%s(%d)", t.Process.Name, t.Process.PID)
}

// readResolvConfig parses /etc/resolv.conf and returns the list of name servers
// as well as the search list defined in it. It mimics the parsing behavior
// specified in https://man7.org/linux/man-pages/man5/resolv.conf.5.html.
func (t *Tenant) readResolvConf(tfs *mountineer.Mountineer) (nameservers []net.IP, searchlist []string) {
	nameservers = []net.IP{} // ensure non-null when marshalling
	searchlist = []string{}
	f, err := tfs.Open("/etc/resolv.conf")
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		// According to resolv.conf(5), lines starting with a "#" or ";" in the
		// first column are to be treated as comments; see also:
		// https://man7.org/linux/man-pages/man5/resolv.conf.5.html
		if strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}
		switch fields[0] {
		case "nameserver": // nameserver IPADDR
			// ...can be specified multiple times
			for _, addr := range fields[1:] {
				if ip := net.ParseIP(addr); ip != nil {
					if ipv4 := ip.To4(); ipv4 != nil {
						ip = ipv4
					}
					nameservers = append(nameservers, ip)
				}
			}
		case "search": // search DOMAIN ...
			// multiple "search" and "domain" clauses are mutually exclusive,
			// with the last one standing; see also:
			// https://man7.org/linux/man-pages/man5/resolv.conf.5.html
			searchlist = fields[1:]
		case "domain": // domain DOMAIN
			// multiple "search" and "domain" clauses are mutually exclusive,
			// with the last one standing; see also:
			// https://man7.org/linux/man-pages/man5/resolv.conf.5.html
			if len(fields) != 2 {
				continue
			}
			searchlist = []string{fields[1]}
		default:
			// ...silently ignore all other lines
		}
	}

	return
}

// readHosts read /etc/hosts, if present, and returns the hostname-to-IP address
// mapping defined in it. It mimics the parsing behavior specified in
// http://man7.org/linux/man-pages/man5/hosts.5.html.
func (t *Tenant) readHosts(tfs *mountineer.Mountineer) map[string]net.IP {
	hosts := map[string]net.IP{}
	f, err := tfs.Open("/etc/hosts")
	if err != nil {
		return hosts
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		// According to hosts(5), "#" can appear anywhere in a line, starting
		// the comment; see also:
		// http://man7.org/linux/man-pages/man5/hosts.5.html
		fields := strings.Fields(strings.Split(scanner.Text(), "#")[0])
		if len(fields) < 2 {
			continue
		}
		ip := net.ParseIP(fields[0])
		if ip == nil {
			continue
		}
		if ipv4 := ip.To4(); ipv4 != nil {
			ip = ipv4
		}
		for _, name := range fields[1:] {
			hosts[name] = ip
		}
	}

	return hosts
}

// readSingleLine reads the specified file, but only its first line and ignoring
// anything else. Additionally puts a limit of 256 characters on the line read,
// which is sufficient for more or less well-formed host and domain names. This
// avoids attacks on our discovery.
func (t *Tenant) readSingleLine(tfs *mountineer.Mountineer, path string) string {
	f, err := tfs.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 256), 256)
	if !scanner.Scan() {
		return ""
	}
	return scanner.Text()
}

// uname returns the UTS-namespaced uname() of this tenant. Returns the zero
// name on any failure.
func (t *Tenant) uname() string {
	var hostname string
	utsnsref := t.Process.Namespaces[model.UTSNS].Ref()
	if len(utsnsref) != 1 {
		return ""
	}
	_ = ops.Visit(func() {
		hostname, _ = os.Hostname()
	}, ops.NewTypedNamespacePath(utsnsref[0], species.CLONE_NEWUTS))
	return hostname
}

// boundedCaps discovers the set of capabilities from the tenant's process'
// status information and returns it as a []byte.
func (t *Tenant) caps(key string) []byte {
	f, err := os.Open("/proc/" + strconv.FormatUint(uint64(t.Process.PID), 10) + "/status")
	if err != nil {
		return nil
	}
	defer func() { _ = f.Close() }()

	var capBnd []byte
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.SplitN(scanner.Text(), ":", 2)
		if len(fields) != 2 {
			continue
		}
		if fields[0] != key {
			continue
		}
		var err error
		capBnd, err = hex.DecodeString(strings.TrimSpace(fields[1]))
		if err != nil {
			return nil
		}
		break
	}

	if err := scanner.Err(); err != nil {
		return nil
	}
	return capBnd
}

// Sort the list of tenants by their names and in place. However, the init(1)
// process with PID 1 always comes first.
func (t Tenants) Sort() {
	sort.SliceStable(t, func(a, b int) bool {
		initA := t[a].Process.PID == 1
		initB := t[b].Process.PID == 1
		if initA || initB {
			return initA
		}
		return t[a].Name() < t[b].Name()
	})
}
