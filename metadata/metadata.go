// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package metadata

import (
	"encoding/json"

	gostwire "github.com/siemens/ghostwire/v2"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
)

// Metadata is a plugin function that returns (additional) metadata, optionally
// basing on a discovery result. The returned metadata must be in form of JSON
// marshallable data (especially struct) and the top-level map string indices,
// when marshalling into JSON, will become field names inside the metadata
// toplevel discovery result element.
type Metadata func(gostwire.DiscoveryResult) map[string]interface{}

// Augment augments the passed metadata with additional plugin-supplied metadata
// and returns the final result in form of a string-indexed map, ready to be
// used in JSON marshalling, et cetera.
func Augment(result gostwire.DiscoveryResult, metadata interface{}) (map[string]interface{}, error) {
	log.Debugf("metadata discovery started...")
	augmented, err := toMap(metadata)
	if err != nil {
		return nil, err
	}
	for _, metadata := range plugger.Group[Metadata]().PluginsSymbols() {
		if md := metadata.S(result); md != nil {
			// ...merges metadata returned by plugin
			log.Debugf("merging metadata from plugin '%s': %v", metadata.Plugin, md)
			m, err := toMap(md)
			if err != nil {
				log.Warnf("cannot merge metadata from plugin '%s': %s", metadata.Plugin, err.Error())
			}
			deepMapMerge(m, augmented)
		}
	}
	log.Debugf("metadata discovery finished")
	return augmented, nil
}

// toMap deeply converts a struct into a map.
func toMap(x interface{}) (map[string]interface{}, error) {
	b, err := json.Marshal(x)
	if err != nil {
		return nil, err
	}
	r := map[string]interface{}{}
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, err
	}
	return r, nil
}

// deepMapMerge deeply merges map a into map b, modifying b.
func deepMapMerge(a, b map[string]interface{}) {
	for keyA, valA := range a {
		// If the values in a and b for the key found in a are both maps, then
		// recursively merge.
		if valA, ok := valA.(map[string]interface{}); ok {
			if valB, ok := b[keyA].(map[string]interface{}); ok {
				deepMapMerge(valA, valB)
				continue
			}
		}
		// Otherwise, simply overwrite any existing value in b for the key with
		// the value from a.
		b[keyA] = valA
	}
}
