// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package containerd

import (
	"context"
	"sort"
	"strings"
	"time"

	detect "github.com/siemens/ghostwire/v2/turtlefinder/detector"

	cdclient "github.com/containerd/containerd"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	cdengine "github.com/thediveo/whalewatcher/engineclient/containerd"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/containerd"
)

// Register this Docker container (engine) discovery plugin. This statically
// ensures that the Detector interface is fully implemented.
func init() {
	plugger.Group[detect.Detector]().Register(
		&Detector{}, plugger.WithPlugin("containerd"))
}

// Detector implements the detect.Detector interface; naming it the same as the
// interface makes the plugin symbol registration autodetect the correct alias
// name.
type Detector struct{}

// EngineNames returns the process name of the containerd engine process.
func (d *Detector) EngineNames() []string {
	return []string{"containerd"}
}

// NewWatcher returns a watcher for tracking alive containerd containers.
func (d *Detector) NewWatcher(ctx context.Context, pid model.PIDType, apis []string) watcher.Watcher {
	sort.Strings(apis) // in-place
	for _, apipathname := range apis {
		if strings.HasSuffix(apipathname, ".ttrpc") {
			continue
		}
		// As containerd's go client will accept more or less any API pathname
		// we throw at it and throw up only when actually trying to communicate
		// with the engine and only after some time, it's not sufficient to just
		// create the watcher, we also need to check that we actually can
		// successfully talk with the daemon. Querying the daemon's version
		// information sufficies and ensures that a partiular API path is
		// useful.
		log.Debugf("dialing containerd endpoint '%s'", apipathname)
		w, err := containerd.New(apipathname, nil, cdengine.WithPID(int(pid)))
		if err == nil {
			ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			_, err := w.Client().(*cdclient.Client).Version(ctx)
			if err := ctx.Err(); err != nil {
				log.Debugf("containerd API Info call context hit deadline: %s", err.Error())
			}
			cancel()
			if err == nil {
				return w
			}
			w.Close()
		}
		log.Debugf("containerd API endpoint '%s' failed: %s", apipathname, err.Error())
	}
	log.Errorf("no working containerd API endpoint found.")
	return nil
}
