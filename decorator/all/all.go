// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package all

import (
	_ "github.com/siemens/ghostwire/v2/decorator/dockernet"   // activate Docker-managed network alias name decoration.
	_ "github.com/siemens/ghostwire/v2/decorator/dockerproxy" // activate nerdctl-managed CNI network alias name decoration.
	_ "github.com/siemens/ghostwire/v2/decorator/ieappicon"   // include (on-demand) IE App icon decoration.
	_ "github.com/siemens/ghostwire/v2/decorator/nerdctlnet"  // activate nerdctl-managed CNI network alias name decoration.
	_ "github.com/siemens/ghostwire/v2/decorator/podmannet"   // activate podman-managed network alias name decoration.
)
