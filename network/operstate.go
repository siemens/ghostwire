// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package network

import "fmt"

// OperState indicates the operational state of a network interface.
type OperState uint8

// Operational states of network interfaces. See also:
// https://www.kernel.org/doc/Documentation/networking/operstates.txt
const (
	Unknown        OperState = iota // operational state hasn't been specified by driver or userspace.
	NotPresent                      // unused in Linux kernel.
	Down                            // interface is unable to transfer data on "physical" level.
	LowerLayerDown                  // an interface onto which this interface is stacked is in Down state.
	Testing                         // unused in Linux kernel.
	Dormant                         // physical layer is up, but waiting for external event (Godot?).
	Up                              // interface is operational up and can be used.
)

// TerminalIcon returns a Unicode "icon" for the given operational state.
func (s OperState) TerminalIcon() string {
	return icons[s]
}

// Name returns the name of the operational state (in CAPITALS).
func (s OperState) Name() string {
	switch s {
	case Unknown:
		return "UNKNOWN"
	case NotPresent:
		return "NOTPRESENT"
	case Down:
		return "DOWN"
	case LowerLayerDown:
		return "LOWERLAYERDOWN"
	case Testing:
		return "TESTING"
	case Dormant:
		return "DORMANT"
	case Up:
		return "UP"
	default:
		return fmt.Sprintf("OperState(%d)", s)
	}
}

var icons = map[OperState]string{
	Unknown:        "◭", // unknown is not down, so it is considered to be operational.
	NotPresent:     "✖",
	Down:           "▽",
	LowerLayerDown: "⏬",
	Testing:        "*",
	Dormant:        "💤",
	Up:             "▲",
}
