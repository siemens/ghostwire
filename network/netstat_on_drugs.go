// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import (
	"bufio"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"github.com/thediveo/lxkns/model"
	"golang.org/x/sys/unix"
)

// SocketState is a TCP's or UDP's socket state, such as listening, connected
// ("established"), et cetera.
type SocketState uint8

// TCP socket state codes are defined in
// https://elixir.bootlin.com/linux/v5.13.13/source/include/net/tcp_states.h
const (
	TCP_ESTABLISHED SocketState = iota + 1
	TCP_SYN_SENT
	TCP_SYN_RECV
	TCP_FIN_WAIT1
	TCP_FIN_WAIT2
	TCP_TIME_WAIT
	TCP_CLOSE
	TCP_CLOSE_WAIT
	TCP_LAST_ACK
	TCP_LISTEN
	TCP_CLOSING
	TCP_NEW_SYN_RECV
)

// String returns a clear-text socket status message.
func (s SocketState) String() string {
	if desc := socketStateDescriptions[s]; desc != "" {
		return desc
	}
	return fmt.Sprintf("SocketState(%d)", s)
}

var socketStateDescriptions = map[SocketState]string{
	TCP_ESTABLISHED:  "connected",
	TCP_SYN_SENT:     "SYN sent",
	TCP_SYN_RECV:     "SYN received",
	TCP_FIN_WAIT1:    "FIN WAIT1",
	TCP_FIN_WAIT2:    "FIN WAIT2",
	TCP_TIME_WAIT:    "time wait",
	TCP_CLOSE:        "closed",
	TCP_CLOSE_WAIT:   "close wait",
	TCP_LAST_ACK:     "last ACK",
	TCP_LISTEN:       "listening",
	TCP_CLOSING:      "closing",
	TCP_NEW_SYN_RECV: "new SYN received",
}

// UDP socket states are not defined explicitly, but instead the udp.c Linux
// kernel module reuses a few TCP socket states; for instance, see:
// https://elixir.bootlin.com/linux/v5.13.13/source/net/ipv4/udp.c#L786
const (
	UDP_ESTABLISHED = TCP_ESTABLISHED // UDP socket is connected
	UDP_LISTEN      = TCP_CLOSE       // sic! arguably semantically not 100% correct...
)

// SocketSimplifiedState indicates whether a TCP or UDP socket is listening or
// connected, and nothing else; this hides all the gory details of SocketState.
type SocketSimplifiedState uint8

// Gostwire's simplified socket states; please note that the concept of
// "connected" versus "listen" really doesn't map well onto UDP socket. What
// Gostwire considers to be a "listening" UDP socket is any open UDP socket that
// isn't connected to a specific remote socket. Connecting a UDP socket here
// means that the remote peer address has been set and the network stack filters
// out (throws away) any UDP packets reaching the socket but originating from
// sources other than the set remote peer address.
const (
	Unconnected = SocketSimplifiedState(0)               // undefined state
	Connected   = SocketSimplifiedState(TCP_ESTABLISHED) // both TCP and UDP sockets
	Listening   = SocketSimplifiedState(TCP_LISTEN)      // both TCP and UDP sockets
)

// String returns a clear-text socket simplified status message.
func (s SocketSimplifiedState) String() string {
	switch s {
	case Unconnected:
		return "unconnected"
	case Connected:
		return socketStateDescriptions[TCP_ESTABLISHED]
	case Listening:
		return socketStateDescriptions[TCP_LISTEN]
	default:
		return fmt.Sprintf("SocketSimplifiedState(%d)", s)
	}
}

// ProcessSocket describes the communication parameters of a network socket and
// which process is using it. Please note that the same socket can be used by
// multiple processes by sharing its file descriptor.
type ProcessSocket struct {
	Family          AddressFamily         // address family, such as unix.AD_INET6, ...
	Protocol        Protocol              // transport protocol, such as syscall.IPPROTO_TCP, ...
	LocalIP         net.IP                // local IP address; IPv4 addresses are in .To4() format.
	LocalPort       uint16                // local TCP/UDP port
	RemoteIP        net.IP                // remote IP address; IPv4 addresses are in .To4() format.
	RemotePort      uint16                // remote TCP/UDP port
	State           SocketState           // (detailed) socket state
	SimplifiedState SocketSimplifiedState // simplified state: either listening or connected (both TCP and UDP)
	PIDs            []model.PIDType       // processes using this socket
	Processes       []*model.Process      // processes using this socket
	IPv4Mapped      bool                  // IPv6 socket handling IPv4 traffic?
	Nifs            Interfaces            // network interfaces handling this traffic, based on address/routing data.
}

// ProcessSockets is a list of ProcessSocket elements, that optionally can be
// sorted in-place.
type ProcessSockets []ProcessSocket

// socketToProcessMap maps socket inode numbers to one or more processes using
// this socket, as discovered from the open file descriptors (courtesy of your
// proc file system) of the processes. Multiple processes might share the same
// file descriptor.
type socketToProcessMap map[uint64][]model.PIDType

// discoverSockets discovers the (Process)Sockets of a specific IP version and
// transport protocol in a network namespace referenced via one of the processes
// attached to the network namespace.
//
// Please note that the following ProcessSocket fields are not set, but instead
// need to be resolved by the caller: Nifs.
func discoverSockets(procroot string, pid model.PIDType, af int, proto int, sm socketToProcessMap) []ProcessSocket {
	path := fmt.Sprintf("%s/%d/net/", procroot, pid)
	switch proto {
	case syscall.IPPROTO_TCP:
		path += "tcp"
	case syscall.IPPROTO_UDP:
		path += "udp"
	default:
		panic(fmt.Sprintf("invalid transport-layer protocol %d", proto))
	}
	switch af {
	case unix.AF_INET:
	case unix.AF_INET6:
		path += "6"
	default:
		panic(fmt.Sprintf("invalid address family %d", af))
	}
	sox := []ProcessSocket{}
	// "path" is garantueed to be a procfs-based path, so no need to deploy the
	// mountineers... And no reason for gosec to go berserk: there is path where
	// user-controlled input can reach procroot and it is otherwise only used in
	// unit tests with a test-local fake procfs root.
	f, err := os.Open(path) // #nosec G304
	if err != nil {
		return sox
	}
	defer func() { _ = f.Close() }()
	scanner := bufio.NewScanner(f)
	// Skip the first "header" line.
	if !scanner.Scan() {
		return sox
	}
	for scanner.Scan() {
		if procsock := newProcessSocket(scanner.Text(), af, proto, sm); procsock.State != 0 {
			sox = append(sox, procsock)
		}
	}
	return sox
}

// newProcessSocket returns a ProcessSocket with the information from
// /proc/[PID]/net/[tcp|udp]6? filled in. Please note that ProcessSocket
// contains fields not set by newProcessSocket, such as the network interfaces
// handling the traffic.
func newProcessSocket(procnetline string, af int, proto int, sm socketToProcessMap) (procsock ProcessSocket) {
	fields := strings.Fields(procnetline)
	if len(fields) < 10 {
		return
	}
	procsock.LocalIP, procsock.LocalPort = decodeSockAddrPort(fields[1])
	procsock.RemoteIP, procsock.RemotePort = decodeSockAddrPort(fields[2])
	ino, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return
	}
	procsock.PIDs = sm[ino]
	state, err := strconv.ParseUint(fields[3], 16, 8)
	if err != nil {
		return
	}
	procsock.State = SocketState(state)
	switch proto {
	case syscall.IPPROTO_TCP:
		switch procsock.State {
		case TCP_LISTEN:
			procsock.SimplifiedState = Listening
		case TCP_ESTABLISHED, TCP_FIN_WAIT1, TCP_FIN_WAIT2:
			procsock.SimplifiedState = Connected
		case TCP_CLOSE, TCP_CLOSE_WAIT, TCP_CLOSING, TCP_LAST_ACK, TCP_NEW_SYN_RECV, TCP_SYN_RECV, TCP_SYN_SENT, TCP_TIME_WAIT:
			procsock.SimplifiedState = Unconnected
		default:
			procsock.State = 0
			return
		}
	case syscall.IPPROTO_UDP:
		switch procsock.State {
		case UDP_LISTEN:
			procsock.SimplifiedState = Listening
		case UDP_ESTABLISHED:
			procsock.SimplifiedState = Connected
		default:
			procsock.State = 0
			return
		}
	default:
		procsock.State = 0
		return
	}
	procsock.Family = AddressFamily(af)
	procsock.Protocol = Protocol(proto)
	return
}

// discoverAllSockInodes returns a map of the inodes-to-PID for all sockets that
// currently exist in the system.
func discoverAllSockInodes(procroot string) socketToProcessMap {
	procentries, err := ioutil.ReadDir(procroot)
	if err != nil {
		return nil
	}
	spm := socketToProcessMap{}
	for _, procentry := range procentries {
		// Get the process PID as a number.
		pid, err := strconv.ParseInt(procentry.Name(), 10, 32)
		if err != nil || pid <= 0 {
			continue
		}
		//
		fdbasepath := procroot + "/" + procentry.Name() + "/fd"
		fdentries, err := ioutil.ReadDir(fdbasepath)
		if err != nil {
			continue
		}
		for _, fdentry := range fdentries {
			stat, err := os.Stat(fdbasepath + "/" + fdentry.Name())
			if err != nil || stat.Mode()&os.ModeSocket == 0 {
				continue
			}
			sysstat, ok := stat.Sys().(*syscall.Stat_t)
			if !ok {
				continue
			}
			fdino := sysstat.Ino
			if pids, ok := spm[fdino]; ok {
				spm[fdino] = append(pids, model.PIDType(pid))
			} else {
				spm[fdino] = []model.PIDType{model.PIDType(pid)}
			}
		}
	}
	return spm
}

// decodeSockAddrPort returns the IP (v4/v6) address and port number encoded in
// the ACII hex string passed to it, taking the platform endianess into account.
func decodeSockAddrPort(addrport string) (net.IP, uint16) {
	fields := strings.Split(addrport, ":")
	if len(fields) != 2 {
		return nil, 0
	}
	// The port hex 2-digit number is always in big endian, regardless of the
	// endianess of the platform.
	port, err := strconv.ParseUint(fields[1], 16, 16)
	if err != nil {
		return nil, 0
	}
	// The IP address is in platform endianess as it nothing more than a hex
	// dump of the address field of a socket. And to make matters worse,
	// addresses are scrambled in form of 32bit words. Oh well.
	var ip net.IP
	if len(fields[0]) == 8 {
		// An IPv4 address is a single 32bit BE/LE, so we only need to swap
		// that.
		b, err := hex.DecodeString(fields[0])
		if err != nil {
			return nil, 0
		}
		if isLE {
			reverseUint32(b, 0)
		}
		ip = net.IP(b)
	} else {
		// Oh bulls! An IPv6 address is handlet as 4x32bit BE/LE, not as 128bit
		// BE/LE. That would have been too easy.
		b, err := hex.DecodeString(fields[0])
		if err != nil {
			return nil, 0
		}
		if isLE {
			reverseUint32(b, 0)
			reverseUint32(b, 4)
			reverseUint32(b, 8)
			reverseUint32(b, 12)
		}
		ip = net.IP(b)
	}

	return ip, uint16(port)
}

// reverseUint32 swaps an uint32 value in a byte buffer between little endian
// and big endian, or the other way round. We'll never know.
func reverseUint32(b []byte, idx int) {
	b[idx], b[idx+1], b[idx+2], b[idx+3] = b[idx+3], b[idx+2], b[idx+1], b[idx]
}

// Since the Linux kernel in part simply dumps certain data fields of a socket
// "as is" in form of a hex string, we need to know what endianess we're
// operating with.
var isLE bool

func init() {
	buf := [2]byte{}
	*(*uint16)(unsafe.Pointer(&buf[0])) = uint16(0xCAFE) // #nosec G103
	switch buf {
	case [2]byte{0xFE, 0xCA}:
		isLE = true
	case [2]byte{0xCA, 0xFE}:
		isLE = false
	default:
		panic("can not determine native endianness.")
	}
}

// Sort sorts a list of ProcessSocket elements:
// - local port,
// - transport protocol,
// - (reverse) connection state -- to show listening ports before connected ports,
// - IPv4 before IPv6,
// - local address,
// - remote address,
// - remote port.
func (p ProcessSockets) Sort() {
	sort.SliceStable(p, func(a, b int) bool {
		switch {
		case p[a].LocalPort < p[b].LocalPort:
			return true
		case p[a].LocalPort > p[b].LocalPort:
			return false
		}
		switch { // note: TCP < UDP
		case p[a].Protocol < p[b].Protocol:
			return true
		case p[a].Protocol > p[b].Protocol:
			return false
		}
		switch { // note: listening < connected
		case p[a].SimplifiedState > p[b].SimplifiedState:
			return true
		case p[a].SimplifiedState < p[b].SimplifiedState:
			return false
		}
		switch {
		case p[a].Family < p[b].Family:
			return true
		case p[a].Family > p[b].Family:
			return false
		}
		ipA := Exploded(p[a].LocalIP)
		ipB := Exploded(p[b].LocalIP)
		switch {
		case ipA < ipB:
			return true
		case ipA > ipB:
			return true
		}
		ipA = Exploded(p[a].RemoteIP)
		ipB = Exploded(p[b].RemoteIP)
		switch {
		case ipA < ipB:
			return true
		case ipA > ipB:
			return true
		}
		return p[a].RemotePort < p[b].RemotePort
	})
}
