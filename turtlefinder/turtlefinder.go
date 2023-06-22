// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"context"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/siemens/ghostwire/v2/turtlefinder/detector"
	_ "github.com/siemens/ghostwire/v2/turtlefinder/detector/all" // pull in engine detector plugins

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/containerizer"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/procfsroot"
	"github.com/thediveo/whalewatcher/watcher"
)

// Overseer gives access to information about container engines currently monitored.
type Overseer interface {
	Engines() []*model.ContainerEngine
}

// Contexter supplies a TurtleFinder with a suitable context for long-running
// container engine workload watching.
type Contexter func() context.Context

// TurtleFinder implements the lxkns Containerizer interface to discover alive
// containers from one or more container engines. It can be safely used from
// multiple goroutines.
//
// On demand, a TurtleFinder scans a process list for signs of container engines
// and then tries to contact the potential engines in order to watch their
// containers.
type TurtleFinder struct {
	contexter     Contexter      // contexts for workload watching.
	engineplugins []engineplugin // static list of engine plugins.

	mux     sync.Mutex                // protects the following fields.
	engines map[model.PIDType]*Engine // engines by PID; may be failed.
}

// TurtleFinder implements the lxkns Containerizer interface.
var _ containerizer.Containerizer = (*TurtleFinder)(nil)

// engineplugin represents the process names of a container engine discovery
// plugin, as well as the plugin's Discover function.
type engineplugin struct {
	names      []string          // process names of interest
	detector   detector.Detector // detector plugin interface
	pluginname string            // for housekeeping and logging
}

// engineprocess represents an individual container engine process and the
// container engine discovery plugin responsible for it.
type engineprocess struct {
	proc   *model.Process
	engine *engineplugin
}

// New returns a TurtleFinder object for further use. The supplied contexter is
// called whenever a new container engine has been found and its workload is to
// be watched: the contexter then has to return a suitable (long-running)
// context, it proferably has control over in order to properly shut down the
// background goroutine resources (indirectly) used by a TurtleFinder.
func New(contexter Contexter) *TurtleFinder {
	f := &TurtleFinder{
		contexter: contexter,
		engines:   map[model.PIDType]*Engine{},
	}
	// Query the available turtle finder plugins for the names of processes to
	// look for, in order to later optimize searching the processes; as we're
	// working only with a static set of plugins we only need to query the basic
	// information once.
	namegivers := plugger.Group[detector.Detector]().PluginsSymbols()
	engineplugins := make([]engineplugin, 0, len(namegivers))
	for _, namegiver := range namegivers {
		engineplugins = append(engineplugins, engineplugin{
			names:      namegiver.S.EngineNames(),
			detector:   namegiver.S,
			pluginname: namegiver.Plugin,
		})
	}
	f.engineplugins = engineplugins
	log.Infof("available engine detector plugins: %s",
		strings.Join(plugger.Group[detector.Detector]().Plugins(), ", "))
	return f
}

// Containers returns the current container state of (alive) containers from all
// discovered container engines.
func (f *TurtleFinder) Containers(
	ctx context.Context, procs model.ProcessTable, pidmap model.PIDMapper,
) []*model.Container {
	// Do some quick housekeeping first: remove engines whose processes have
	// vanished.
	if !f.prune(procs) {
		return nil // sorry, we're closed.
	}
	// Then look for new engine processes.
	f.update(ctx, procs)
	// Now query the available engines for containers that are alive...
	f.mux.Lock()
	engines := make([]*Engine, 0, len(f.engines))
	for _, engine := range f.engines {
		// create copies of the engine objects in order to not trash the
		// original engine objects.
		engine := *engine
		engines = append(engines, &engine)
	}
	f.mux.Unlock()
	// Feel the heat and query the engines in parallel; to collect the results
	// we use a buffered channel of the size equal the number of engines to
	// query.
	//
	// TODO: bounded worker model
	log.Infof("consulting %d container engines ... in parallel", len(engines))
	enginecontainers := make(chan []*model.Container, len(engines))
	var wg sync.WaitGroup
	wg.Add(len(engines))
	for _, engine := range engines {
		go func(engine *Engine) {
			defer wg.Done()
			containers := engine.Containers(ctx)
			enginecontainers <- containers
		}(engine)
	}
	// Wait for all query workers to complete and push their results into the
	// buffered channel; only then we close the channel and then pull off the
	// buffered results.
	wg.Wait()
	close(enginecontainers)
	containers := []*model.Container{}
	for conts := range enginecontainers {
		containers = append(containers, conts...)
	}
	// Fill in the engine hierarchy, if necessary: note that we can't use this
	// without knowing the containers and especially their names.
	stackEngines(containers, engines, procs)

	return containers
}

// Close closes all resources associated with this turtle finder. This is an
// asynchronous process. Make sure to also cancel or have already cancelled the
// context
func (f *TurtleFinder) Close() {
	f.mux.Lock()
	defer f.mux.Unlock()
	for _, engine := range f.engines {
		engine.Close()
	}
	f.engines = nil
}

// Engines returns information about the container engines currently being
// monitored.
func (f *TurtleFinder) Engines() []*model.ContainerEngine {
	f.mux.Lock()
	defer f.mux.Unlock()
	engines := make([]*model.ContainerEngine, 0, len(f.engines))
	for _, engine := range f.engines {
		select {
		case <-engine.Done:
			continue // already Done, so ignore this engine.
		default:
			// not Done, so let's move on and add it to the list of available
			// engines.
		}
		engines = append(engines, &model.ContainerEngine{
			ID:      engine.ID,
			Type:    engine.Type(),
			Version: engine.Version,
			API:     engine.API(),
			PID:     model.PIDType(engine.PID()),
		})
	}
	return engines
}

// EngineCount returns the number of container engines currently under watch.
// Callers might want to use the Engines method instead as EngineCount bases on
// it (because we don't store an explicit engine count anywhere).
func (f *TurtleFinder) EngineCount() int {
	f.mux.Lock()
	defer f.mux.Unlock()
	return len(f.engines)
}

// prune any terminated watchers, either because the watcher terminated itself
// or we can't find the associated engine process anymore.
func (f *TurtleFinder) prune(procs model.ProcessTable) bool {
	f.mux.Lock()
	defer f.mux.Unlock()
	if f.engines == nil {
		return false
	}
	for pid, engine := range f.engines {
		if procs[pid] == nil && !engine.IsAlive() {
			delete(f.engines, pid)
			engine.Close() // ...if not already done so.
		}
	}
	return true
}

// update our knowledge about container engines if necessary, given the current
// process table and by asking engine discovery plugins for any signs of engine
// life.
func (f *TurtleFinder) update(ctx context.Context, procs model.ProcessTable) {
	// Look for potential signs of engine life, based on process names...
	engineprocs := []engineprocess{}
NextProcess:
	for _, proc := range procs {
		procname := proc.Name
		for engidx, engine := range f.engineplugins {
			for _, enginename := range engine.names {
				if procname == enginename {
					engineprocs = append(engineprocs, engineprocess{
						proc:   proc,
						engine: &f.engineplugins[engidx], // ...we really don't want to address the loop variable here
					})
					continue NextProcess
				}
			}
		}
	}
	// Next, throw out all engine processes we already know of and keep only the
	// new ones to look into them further. This way we keep the lock as short as
	// possible.
	newengineprocs := make([]engineprocess, 0, len(engineprocs))
	f.mux.Lock()
	for _, engineproc := range engineprocs {
		// Is this an engine PID we already know and watch?
		if _, ok := f.engines[engineproc.proc.PID]; ok {
			continue
		}
		newengineprocs = append(newengineprocs, engineproc)
	}
	f.mux.Unlock()
	if len(newengineprocs) == 0 {
		return
	}
	// Finally look into each new engine process: try to figure out its
	// potential API socket endpoint pathname and then try to contact the engine
	// via this (these) pathname(s)... Again, we aggressively go parallel in
	// contacting new engines.
	var wg sync.WaitGroup
	wg.Add(len(newengineprocs))
	for _, engineproc := range newengineprocs {
		go func(engineproc engineprocess) {
			defer wg.Done()
			log.Debugf("scanning new potential engine process %s (%d) for API endpoints...",
				engineproc.proc.Name, engineproc.proc.PID)
			// Does this process have any listening unix sockets that might act as
			// API endpoints?
			apisox := discoverAPISockets(engineproc.proc.PID)
			if apisox == nil {
				log.Debugf("process %d no API endpoint found", engineproc.proc.PID)
				return
			}
			// Translate the API pathnames so that we can access them from our
			// namespace via procfs wormholes; to make this reliably work we need to
			// evaluate paths for symbolic links...
			for idx, apipath := range apisox {
				root := "/proc/" + strconv.FormatUint(uint64(engineproc.proc.PID), 10) +
					"/root"
				if p, err := procfsroot.EvalSymlinks(apipath, root, procfsroot.EvalFullPath); err == nil {
					apisox[idx] = root + p
				} else {
					log.Warnf("invalid API endpoint at %s", apipath)
					apisox[idx] = ""
				}
			}
			// Ask the contexter to give us a long-living engine workload
			// watching context; just using the background context (or even a
			// request's context) will be a bad idea as it doesn't give the
			// users of a Turtlefinder the means to properly spin down workload
			// watchers when retiring a Turtlefinder.
			enginectx := f.contexter()
			if w := engineproc.engine.detector.NewWatcher(enginectx, engineproc.proc.PID, apisox); w != nil {
				// We've got a new watcher!
				startWatch(enginectx, w)
				eng := NewEngine(enginectx, w)
				f.mux.Lock()
				f.engines[engineproc.proc.PID] = eng
				f.mux.Unlock()
			}
		}(engineproc)
	}
	wg.Wait()
}

// startWatch starts the watch and then shortly waits for a watcher to
// synchronize and then watches in the background (spinning off a separate go
// routine) the watcher synchronizing to its engine state, logging begin and end
// as informational messages.
func startWatch(ctx context.Context, w watcher.Watcher) {
	log.Infof("beginning synchronization to %s engine (PID %d) at API %s",
		w.Type(), w.PID(), w.API())
	// Start the watch including the initial synchronization...
	errch := make(chan error, 1)
	go func() {
		errch <- w.Watch(ctx)
		close(errch)
	}()
	// Wait in the background for the synchronization to complete and then
	// report the engine ID.
	go func() {
		<-w.Ready()
		// Getting the engine ID should be carried out swiftly, so we timebox
		// it.
		idctx, cancel := context.WithTimeout(ctx, 2*time.Second)
		log.Infof("synchronized to %s engine (PID %d) with ID %s",
			w.Type(), w.PID(), w.ID(idctx))
		cancel() // ensure to quickly release cancel
	}()
	// Give the watcher a (short) chance to get in sync, but do not hang around
	// for too long...
	//
	// Oh, well: time.After is kind of hard to use without small leaks.
	// Now, a 5s timer will be GC'ed after 5s anyway, but let's do it
	// properly for once and all, to get the proper habit. For more
	// background information, please see, for instance:
	// https://www.arangodb.com/2020/09/a-story-of-a-memory-leak-in-go-how-to-properly-use-time-after/
	wecker := time.NewTimer(2 * time.Second)
	select {
	case <-w.Ready():
		if !wecker.Stop() { // drain the timer, if necessary.
			<-wecker.C
		}
	case <-wecker.C:
		log.Warnf("%s engine (PID %d) not yet synchronized ... continuing in background",
			w.Type(), w.PID())
	}
}
