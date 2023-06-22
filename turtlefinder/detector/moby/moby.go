// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package moby

import (
	"context"
	"sort"
	"time"

	detect "github.com/siemens/ghostwire/v2/turtlefinder/detector"

	"github.com/docker/docker/client"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	mobyengine "github.com/thediveo/whalewatcher/engineclient/moby"
	"github.com/thediveo/whalewatcher/watcher"
	"github.com/thediveo/whalewatcher/watcher/moby"
)

// Register this Docker container (engine) discovery plugin. This statically
// ensures that the Detector interface is fully implemented.
func init() {
	plugger.Group[detect.Detector]().Register(
		&Detector{}, plugger.WithPlugin("dockerd"))
}

// Detector implements the detect.Detector interface; naming it the same as the
// interface makes the plugin symbol registration autodetect the correct alias
// name.
type Detector struct{}

// EngineNames returns the process name of the Docker/moby engine process.
func (d *Detector) EngineNames() []string {
	return []string{"dockerd"}
}

// NewWatcher returns a watcher for tracking alive Docker containers.
func (d *Detector) NewWatcher(ctx context.Context, pid model.PIDType, apis []string) watcher.Watcher {
	sort.Strings(apis) // in-place
	for _, apipathname := range apis {
		// As Docker's go client will accept any API pathname we throw at it and
		// throw up only when actually trying to communicate with the engine,
		// it's not sufficient to just create the watcher, we also need to check
		// that we actually can successfully talk with the daemon. Querying the
		// daemon's info sufficies and ensures that a partiular API path is
		// useful.
		log.Debugf("dialing Docker endpoint 'unix://%s'", apipathname)
		w, err := moby.New("unix://"+apipathname, nil, mobyengine.WithPID(int(pid)))
		if err == nil {
			ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
			_, err = w.Client().(*client.Client).Info(ctx)
			if ctxerr := ctx.Err(); ctxerr != nil {
				log.Debugf("Docker API Info call context hit deadline: %s", ctxerr.Error())
			}
			cancel()
			if err == nil {
				return w
			}
			w.Close()
		}
		log.Debugf("Docker API endpoint 'unix://%s' failed: %s", apipathname, err.Error())
	}
	log.Errorf("no working Docker API endpoint found.")
	return nil
}
