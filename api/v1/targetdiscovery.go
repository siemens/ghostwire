// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import gostwire "github.com/siemens/ghostwire/v2"

// TargetDiscoveryResult contains information about the capture targets for JSON
// marshalling.
type TargetDiscoveryResult struct {
	Metadata   Metadata       `json:"metadata"`
	Containers captureTargets `json:"containers"`
}

// NewTargetDiscoveryResult returns a new TargetDiscoveryResult for the
// specified network namespace discovery results, to be marshalled into JSON.
func NewTargetDiscoveryResult(result gostwire.DiscoveryResult) TargetDiscoveryResult {
	return TargetDiscoveryResult{
		Metadata:   NewMetadata(result),
		Containers: newCaptureTargets(result),
	}
}
