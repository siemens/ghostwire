// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package portfwd

import (
	"github.com/thediveo/nufftables"
	"github.com/thediveo/nufftables/portfinder"
)

// Portwardings returns forwarded ports discovered from the table map of a
// single specific table family passed to it.
type PortForwardings func(
	tables nufftables.TableMap,
	family nufftables.TableFamily,
) []*portfinder.ForwardedPortRange
