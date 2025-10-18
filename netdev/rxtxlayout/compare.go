// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package rxtxlayout

import "strings"

// CompareNetdevsByName sorts Netdev objects by their names, with the exception
// of putting “lo” always first (so it sticks out like a sore toe).
func CompareNetdevsByName(a, b *Netdev) int {
	isLoA := a.Name == "lo"
	isLoB := b.Name == "lo"
	if isLoA != isLoB {
		switch {
		case isLoA:
			return -1
		default:
			return 1
		}
	}
	return strings.Compare(a.Name, b.Name)
}

// CompareIRQsByID sorts IRQ objects by their IDs in increasing order.
func CompareIRQsByID(a, b *IRQ) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return 0
	}
}

// CompareNAPIsByID sorts NAPI objects by their IDs in increasing order.
func CompareNAPIsByID(a, b *NAPI) int {
	switch {
	case a.ID < b.ID:
		return -1
	case a.ID > b.ID:
		return 1
	default:
		return 0
	}
}
