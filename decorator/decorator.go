// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package decorator

import (
	"context"

	"github.com/siemens/ghostwire/v2/network"

	"github.com/thediveo/lxkns/model"
)

// Decorate is called after the core discovery of network namespaces as well as
// network interfaces configuration and relations has finished, so decorators
// can post-process the information model. As at least some decorators might
// need to run engine queries any decoration post-processing can be time-boxed
// by passing a suitable Context.
type Decorate func(
	ctx context.Context,
	allnetns network.NetworkNamespaces,
	allprocs model.ProcessTable,
	engines []*model.ContainerEngine,
)
