// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package engines

import (
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/metadata"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
)

func init() {
	plugger.Group[metadata.Metadata]().Register(
		Metadata, plugger.WithPlugin("engines"))
}

// EngineMeta contains meta-information about an individual container engine
// with active workload.
type EngineMeta struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Version string `json:"version"`
}

// Metadata returns metadata about the container engines currently with workload
// (that is, alive containers).
func Metadata(r gostwire.DiscoveryResult) map[string]interface{} {
	engines := map[*model.ContainerEngine]struct{}{}
	for _, cntr := range r.Lxkns.Containers {
		engines[cntr.Engine] = struct{}{}
	}
	enginesMeta := make([]EngineMeta, 0, len(engines))
	for engine := range engines {
		enginesMeta = append(enginesMeta, EngineMeta{
			ID:      engine.ID,
			Type:    engine.Type,
			Version: engine.Version,
		})
	}
	return map[string]interface{}{
		"container-engines": enginesMeta,
	}
}
