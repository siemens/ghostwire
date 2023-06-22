// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	"time"

	"github.com/siemens/ghostwire/v2/metadata"
	_ "github.com/siemens/ghostwire/v2/metadata/all"

	gostwire "github.com/siemens/ghostwire/v2"
)

// CreatorID specifies the metadata creator ID.
const CreatorID = "gostwire"

// Metadata holds meta information about a discovery, such as the creator used
// ("Dr. Livingstone, I presume") and when the discovery was done.
type Metadata map[string]interface{}

// NewMetadata returns new and properly filled-in discovery meta data. It
// invokes the registered metadata plugins to augment the baseline metadata with
// additional tidbits of information.
func NewMetadata(result gostwire.DiscoveryResult) Metadata {
	basemd := Metadata{
		"creator":            "Gostwire Linux virtual networking topology and configuration discovery engine",
		"creator-id":         CreatorID,
		"creator-version":    gostwire.SemVersion,
		"creation-timestamp": time.Now().UTC(),
	}
	md, err := metadata.Augment(result, basemd)
	if err != nil {
		return basemd
	}
	return md
}
