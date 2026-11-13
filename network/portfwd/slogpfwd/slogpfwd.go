// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package slogpfwd

import (
	"log/slog"

	"github.com/thediveo/nufftables/portfinder"
)

// ForwardedPortAttrs returns a list of slog attributes describing the passed
// port forwarding in terms of protocol, IP address, port range, IP address
// forwarded to, as well as the port range forwarded to.
func ForwardedPortAttrs(fp *portfinder.ForwardedPortRange) []any {
	return []any{
		slog.String("protocol", fp.Protocol),
		slog.String("ip", fp.IP.String()),
		slog.Group("port",
			slog.Uint64("min", uint64(fp.PortMin)),
			slog.Uint64("max", uint64(fp.PortMax)),
		),
		slog.Group("forward",
			slog.String("ip", fp.ForwardIP.String()),
			slog.Group("port",
				slog.Uint64("min", uint64(fp.ForwardPortMin+fp.PortMin)),
				slog.Uint64("max", uint64(fp.ForwardPortMin+fp.PortMax)),
			),
		),
	}
}
