// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package engines

import (
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/meta"
)

func init() {
	plugger.Group[meta.Producer]().Register(
		Metadata, plugger.WithPlugin("engines"))
}

// EngineMeta contains meta-information about an individual container engine
// with active workload.
type EngineMeta struct {
	ID      string        `json:"id"`
	Type    string        `json:"type"`
	Version string        `json:"version"`
	PID     model.PIDType `json:"pid"`
}

// Metadata returns metadata about the container engines, regardless of
// currently with or without workload (that is, alive containers).
func Metadata(r gostwire.DiscoveryResult) meta.Data {
	enginesMeta := make([]EngineMeta, 0, len(r.Engines))
	for _, engine := range r.Engines {
		enginesMeta = append(enginesMeta, EngineMeta{
			ID:      engine.ID,
			Type:    engine.Type,
			Version: engine.Version,
			PID:     engine.PID,
		})
	}
	return meta.Data{
		"container-engines": enginesMeta,
	}
}
