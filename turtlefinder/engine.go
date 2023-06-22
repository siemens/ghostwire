// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"context"
	"time"

	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/whalewatcher/watcher"
)

// Engine watches a container engine for signs of container life, using a
// supplied "whale watcher". Engine objects then can be queried for a list of
// currently alive (running/paused) containers they manage.
type Engine struct {
	watcher.Watcher               // engine watcher (doubles as engine adapter).
	ID              string        // engine ID.
	Version         string        // engine version.
	Done            chan struct{} // closed when watch is done/has terminated.
}

// NewEngine returns a new Engine given the specified watcher. The Engine is
// already "warming up" and has started watching (using the given context).
func NewEngine(ctx context.Context, watch watcher.Watcher) *Engine {
	idctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	e := &Engine{
		Watcher: watch,
		ID:      watch.ID(idctx),
		Version: watch.Version(idctx),
		Done:    make(chan struct{}, 1), // might never be picked up in some situations
	}
	cancel() // ensure to quickly release cancel, silence linter
	log.Infof("watching %s container engine (PID %d) with ID '%s', version '%s'",
		watch.Type(), watch.PID(), e.ID, e.Version)
	go func() {
		err := e.Watcher.Watch(ctx)
		log.Infof("stopped watching container engine (PID %d), reason: %s",
			watch.PID(), err.Error())
		close(e.Done)
		e.Close()
	}()
	return e
}

// Containers returns the alive containers managed by this engine, using the
// associated watcher.
func (e *Engine) Containers(ctx context.Context) []*model.Container {
	eng := &model.ContainerEngine{
		ID:      e.ID,
		Type:    e.Watcher.Type(),
		Version: e.Version,
		API:     e.Watcher.API(),
		PID:     model.PIDType(e.Watcher.PID()),
	}
	// Adapt the whalewatcher container model to the lxkns container model,
	// where the latter takes container engines and groups into account of its
	// information model. We only need to set the container engine, as groups
	// will be handled separately by the various (lxkns) decorators.
	for _, projname := range append(e.Watcher.Portfolio().Names(), "") {
		project := e.Watcher.Portfolio().Project(projname)
		if project == nil {
			continue
		}
		for _, container := range project.Containers() {
			// Ouch! Make sure to clone the Labels map and not simply pass it
			// directly on to our ontainer objects. Otherwise decorators adding
			// labels would modify the labels shared through the underlying
			// container label source. So, clone the labels (top-level only) and
			// then happy decorating.
			clonedLabels := model.Labels{}
			for k, v := range container.Labels {
				clonedLabels[k] = v
			}
			cntr := &model.Container{
				ID:     container.ID,
				Name:   container.Name,
				Type:   eng.Type,
				Flavor: eng.Type,
				PID:    model.PIDType(container.PID),
				Paused: container.Paused,
				Labels: clonedLabels,
				Engine: eng,
			}
			eng.AddContainer(cntr)
		}
	}
	return eng.Containers
}

// IsAlive returns true as long as the engine watcher is operational and hasn't
// permanently failed/terminated.
func (e *Engine) IsAlive() bool {
	select {
	case <-e.Done:
		return false
	default:
		// nothing to see, move on!
	}
	return true
}
