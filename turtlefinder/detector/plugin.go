// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package detector

import (
	"context"

	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// Detector allows specialized container engine detector plugins to interface
// with the generic engine discovery mechanism.
type Detector interface {
	// EngineNames returns one or more process name(s) of a specific type of
	// container engine.
	EngineNames() []string

	// NewWatcher returns a watcher for tracking alive containers of the
	// container engine accessible by at least one of the specified API
	// pathnames.
	NewWatcher(ctx context.Context, pid model.PIDType, apis []string) watcher.Watcher
}
