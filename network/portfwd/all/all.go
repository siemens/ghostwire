// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package all

import (
	_ "github.com/siemens/ghostwire/v2/network/portfwd/docker"    // activate port fwd detection for Docker
	_ "github.com/siemens/ghostwire/v2/network/portfwd/kubeproxy" // activate port fwd detection for kube-proxy
)
