// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package turtlefinder

import (
	"github.com/thediveo/lxkns/model"
)

// GostwireContainerPrefixLabelName defines the label name for attaching prefix
// information about the engine-hierarchy to containers. Discovery client can
// use these container labels to find out the hierarchy of containers. For
// instance, if container "A" is managed by a container engine hosted inside
// container "B", then container "A" is labelled with prefix "B".
const GostwireContainerPrefixLabelName = "gostwire/container/prefix"

// PrefixSeparator is the separator used in hierarchical prefixes.
const PrefixSeparator = "/"

// stackedEngine temporarily stores additional details about a container engine
// while we figure out if and how engines have been stacked, or rather, put into
// each other.
type stackedEngine struct {
	EncloserName      string           // name derived from enclosing container, if any, otherwise "".
	EncloserEnginePID model.PIDType    // PID of engine PID managing the enclosing container, if any, otherwise 0.
	Prefix            string           // hierarchical prefix for this engine, or "".
	Parent            *stackedEngine   // parent container engine, if any.
	Children          []*stackedEngine // child container engines, if any.
}

// Add a stackedEngine as the child of this stackedEngine, at the same time also
// setting this stackedEngine to be the parent of the added stackedEngine.
func (e *stackedEngine) Add(child *stackedEngine) {
	child.Parent = e
	e.Children = append(e.Children, child)
}

// stackEngines discovers the hierarchical relationships (if any) between
// container engines, that is, when one engine is running inside a container
// managed by another container engine.
func stackEngines(containers []*model.Container, engines []*Engine, proctable model.ProcessTable) {
	// Let's build an index for mapping the PIDs of the containers' initial
	// processes to their containers.
	containersByPID := map[model.PIDType]*model.Container{}
	for _, container := range containers {
		containersByPID[container.PID] = container
	}
	// Index the list of engines we were told, in order to quickly look up the
	// additional information we need to associate with the engines during the
	// stacking process. The index key is an engine's PID, as this is the only
	// correct link between a model.ContainerEngine and a turtlefinder.Engine.
	// We also use this chance to see if an engine is inside a container and in
	// which one in particular.
	stackedEngines := map[model.PIDType]*stackedEngine{}
	for _, engine := range engines {
		// Climb up the process tree until we either hit a container PID or we
		// fall off the ... root? Okay, another +1 on the eternal counter of
		// really bad metaphors.
		var (
			name           string
			outerEnginePID model.PIDType
			container      *model.Container
		)
		proc := proctable[model.PIDType(engine.PID())]
		for proc != nil {
			var ok bool
			if container, ok = containersByPID[proc.PID]; ok {
				name = container.Name
				outerEnginePID = container.Engine.PID
				break
			}
			// rinse and repeat until container PID hit or falling off root.
			proc = proc.Parent
		}
		stackedEngines[model.PIDType(engine.PID())] = &stackedEngine{
			EncloserName:      name,
			EncloserEnginePID: outerEnginePID,
		}
	}
	// Now that we know which engines are containerized, set these engines to be
	// children of the container engines managing the engine-enclosing
	// containers. Hopefully, we end up with some hierarchy. While it is not
	// strictly necessary to explicitly build this engine hierarchy, it helps
	// with detecting sibling engines in the same context, such as side-by-side
	// engines in the host or in some container.
	var nullEngine = &stackedEngine{} // acts as "fake" root
	for _, engine := range stackedEngines {
		if pid := engine.EncloserEnginePID; pid != 0 {
			if parentEngine := stackedEngines[pid]; parentEngine != nil {
				parentEngine.Add(engine)
				continue
			}
		}
		nullEngine.Add(engine)
	}
	// Next, we can finally determine the engine prefixes ("turtle paths") based
	// on the discovered engine hierarchy. Looks like a recursion allergy ;)
	for _, engine := range stackedEngines {
		prefix := ""
		eng := engine
		for eng != nil {
			if eng.EncloserName != "" {
				if prefix == "" {
					prefix = eng.EncloserName
				} else {
					prefix = eng.EncloserName + PrefixSeparator + prefix
				}
			}
			eng = eng.Parent
		}
		engine.Prefix = prefix
	}
	// Finally distribute the per-engine prefixes to the individual containers;
	// the prefixes are attached as Gostwire-specific container labels.
	var engine *model.ContainerEngine
	var cachedEnginePrefix string
	for _, container := range containers {
		if container.Engine != engine {
			engine = container.Engine
			if steng, ok := stackedEngines[engine.PID]; ok {
				cachedEnginePrefix = steng.Prefix
			} else {
				cachedEnginePrefix = ""
			}
		}
		container.Labels[GostwireContainerPrefixLabelName] = cachedEnginePrefix
	}
}
