// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package meta

import (
	"encoding/json"
	"log/slog"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/nonstd/xslog"

	gostwire "github.com/siemens/ghostwire/v2"
)

// Data is JSON-marshallable key-value data, that can optionally be
// hierarchically nested.
type Data = map[string]any

// Producer is a plugin function that returns (additional) Metadata,
// optionally basing on a discovery result. The returned metadata must be in
// form of JSON marshallable data (especially struct) and the top-level map
// string indices, when marshalling into JSON, will become field names inside
// the metadata toplevel discovery result element.
type Producer func(gostwire.DiscoveryResult) Data

// Augment augments the passed metadata with additional plugin-supplied metadata
// and returns the final result in form of a string-indexed map, ready to be
// used in JSON marshalling, et cetera.
func Augment(result gostwire.DiscoveryResult, metadata any) (Data, error) {
	slog.Debug("metadata discovery started...")
	data, err := asData(metadata)
	if err != nil {
		return nil, err
	}
	for _, producer := range plugger.Group[Producer]().PluginsSymbols() {
		moredata := producer.S(result)
		if moredata == nil {
			continue
		}
		// ...merges metadata returned by plugin
		slog.Debug("merging metadata from plugin",
			slog.String("plugin", producer.Plugin),
			slog.Any("metadata", moredata))
		m, err := asData(moredata)
		if err != nil {
			slog.Warn("cannot merge metadata from plugin",
				slog.String("plugin", producer.Plugin),
				xslog.Error(err))
		}
		deepMerge(m, data)
	}
	slog.Debug("metadata discovery finished")
	return data, nil
}

// asData deeply converts a struct into a map.
func asData(x any) (Data, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	r := Data{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// deepMerge deeply merges map a into map b, modifying b.
func deepMerge(a, into Data) {
	for keyA, valA := range a {
		// If the values in a and in this map for the specified key are both
		// maps, then recursively merge.
		if valA, ok := valA.(Data); ok {
			if valB, ok := into[keyA].(Data); ok {
				deepMerge(valA, valB)
				continue
			}
		}
		// Otherwise, simply overwrite any existing value in the destination map
		// for the key with the value from a.
		into[keyA] = valA
	}
}
