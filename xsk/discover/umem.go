// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

import (
	"github.com/vishvananda/netlink"
)

// Umem provides umem configuration and statistics information for a particular
// umem. As umem can be shared between XDP sockets [bound to the same netdev and
// queue], Umem objects describe the umem-specific discovery information that
// really doesn't belong to the XDP sockets themselves.
//
// [bound to the same netdev and queue]: https://www.kernel.org/doc/html/v6.0/networking/af_xdp.html#umem
type Umem struct {
	*netlink.XDPDiagUmem
	FillRingEntries       uint32
	CompletionRingEntries uint32
	FillRingEmpty         uint64
	XSKs                  []XSK // XDP sockets sharing the same umem between them.
}

// UmemsByID indexes umems by their umem IDs.
type UmemsByID map[uint32]*Umem

// ExtractUmems extracts umem configuration and statistics from the discovered
// XSKs.
func ExtractUmems(xsks NetnsXSKs) UmemsByID {
	umems := UmemsByID{}
	for _, netnsXsks := range xsks {
		for _, xsk := range netnsXsks {
			// In case of XSKs sharing their umem, each XSK diagnosis
			// information contains the same umem information, so we create the
			// extracted umem diagnosis information (only) when we encounter a
			// umem ID for the first time.
			umem := umems[xsk.Diag.XDPInfo.Umem.ID]
			if umem == nil {
				umem = &Umem{
					XDPDiagUmem:           xsk.Diag.XDPInfo.Umem,
					FillRingEntries:       xsk.Diag.XDPInfo.UmemFillRingEntries,
					CompletionRingEntries: xsk.Diag.XDPInfo.UmemCompletionRingEntries,
					FillRingEmpty:         xsk.Diag.XDPInfo.Stats.FillRingEmpty,
				}
				umems[xsk.Diag.XDPInfo.Umem.ID] = umem
			}
			umem.XSKs = append(umem.XSKs, xsk)
		}
	}
	return umems
}
