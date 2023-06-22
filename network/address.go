// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"encoding/hex"
	"fmt"
	"net"
	"sort"
	"syscall"

	"golang.org/x/sys/unix"
)

// Address represents a network-layer address with associated information.
type Address struct {
	Family            int    `json:"family"`
	Address           net.IP `json:"address"`
	PrefixLength      uint   `json:"prefixlen"`
	PreferredLifetime uint32 `json:"preferred-lifetime"`
	ValidLifetime     uint32 `json:"valid-lifetime"`
	Scope             int    `json:"scope"`
	Index             int    `json:"index"` // index of network interface the address is assigned to.
}

// Addresses is an unordered list of Address elements.
type Addresses []Address

// Exploded returns the textural IP address representation with 0 padding and
// without any compression. For instance, "000.000.000.000" and
// "0000:0000:0000:0000:0000:0000:0000:0000". The exploded form is mainly of use
// when sorting IP addresses, as this avoids the need for (hex) number-aware
// sorting functions.
func (a Address) Exploded() string {
	return exploded(a.Address, a.Family)
}

// Exploded returns the textural IP address representation with 0 padding and
// without any compression. For instance, "000.000.000.000" and
// "0000:0000:0000:0000:0000:0000:0000:0000". The exploded form is mainly of use
// when sorting IP addresses, as this avoids the need for (hex) number-aware
// sorting functions.
func Exploded(ip net.IP) string {
	if len(ip) == net.IPv4len {
		return exploded(ip, unix.AF_INET)
	}
	return exploded(ip, unix.AF_INET6)
}

// exploded returns the textural IP address representation with 0 padding and
// without any compression, with cross-checking against the address family
// given.
func exploded(ip net.IP, af int) string {
	switch af {
	case unix.AF_INET:
		if octets := ip.To4(); octets != nil {
			return fmt.Sprintf("%03d.%03d.%03d.%03d",
				octets[0], octets[1], octets[2], octets[3])
		}
		panic(ip.String() + " is not an IPPROTO_IP address")
	case unix.AF_INET6:
		hexdigits := make([]byte, 16*2)
		_ = hex.Encode(hexdigits, ip)
		return string(hexdigits[0:4]) + ":" +
			string(hexdigits[4:8]) + ":" +
			string(hexdigits[8:12]) + ":" +
			string(hexdigits[12:16]) + ":" +
			string(hexdigits[16:20]) + ":" +
			string(hexdigits[20:24]) + ":" +
			string(hexdigits[24:28]) + ":" +
			string(hexdigits[28:])
	}
	return ip.String()
}

// Sort sorts a list of Addresses in-place. IPv4 addresses are always sorted
// before IPv6 addresses. IP addresses within the same family (v4 or v6) are
// sorted based on their Exploded form. An address with a shorter prefix length
// sort before the same address with a longer prefix length.
func (as Addresses) Sort() {
	sort.SliceStable(as, func(a, b int) bool {
		if as[a].Family != as[b].Family {
			return as[a].Family < as[b].Family // IPv4 comes/came before IPv6
		}
		expA := as[a].Exploded()
		expB := as[b].Exploded()
		switch {
		case expA < expB:
			return true
		case expA > expB:
			return false
		default:
			return as[a].PrefixLength < as[b].PrefixLength
		}
	})
}

// AddressFamily represents an AF_ address family. Additionally, AdressFamily
// can be String-ified into the text strings "IPv4" or "IPv6".
type AddressFamily int

// String returns either "IPv4" or "IPv6" for an AF_INET(6) address family.
func (af AddressFamily) String() string {
	switch af {
	case unix.AF_INET:
		return "IPv4"
	case unix.AF_INET6:
		return "IPv6"
	default:
		return fmt.Sprintf("AddressFamily(%d)", af)
	}
}

// Protocol represents a (transport) protocol number for TCP or UDP.
// Additionally, Protocol can be String-ified into the text strings "TCP" and
// "UDP".
type Protocol int

// String returns either "TCP" or "UDP" for the given protocol number.
func (p Protocol) String() string {
	switch p {
	case syscall.IPPROTO_TCP:
		return "TCP"
	case syscall.IPPROTO_UDP:
		return "UDP"
	case syscall.IPPROTO_SCTP:
		return "SCTP"
	default:
		return fmt.Sprintf("Protocol(%d)", p)
	}
}

// IP is a net.IP that returns its IPv6 address textual representation in square
// bracket notation "[textual-ipv6-address]". This is useful in "ip:port" output
// contexts, as it is common practise to then render IPv6 addresses as
// "[ipv6]:port" in order to be able to clearly see the port number, which
// otherwise could be mistaken for the final address group.
type IP net.IP

// String returns for IPv4 the IPv4 textual representation (for instance,
// "127.0.0.1") and for IPv6 the textual representation enclosed in square
// brackets, for instance: "[fe80::1]".
func (a IP) String() string {
	if ipv4 := net.IP(a).To4(); len(a) == net.IPv6len && ipv4 == nil {
		return "[" + net.IP(a).String() + "]"
	}
	return net.IP(a).String()
}
