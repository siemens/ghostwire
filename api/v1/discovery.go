// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package v1

import (
	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/thediveo/lxkns/model"
)

// DiscoveryResult represents a Ghostwire v1 discovery result, consisting of
// some metadata, the discovered network namespaces and PID namespaces.
type DiscoveryResult struct {
	Metadata          Metadata          `json:"metadata"`
	NetworkNamespaces networkNamespaces `json:"network-namespaces"`
	PIDNamespaces     pidNamespaces     `json:"pid-namespaces"`
}

// NewDiscoveryResult returns a JSON marshallable DiscoveryResult for the given
// gostwire.DiscoveryResult.
func NewDiscoveryResult(result gostwire.DiscoveryResult) DiscoveryResult {
	return DiscoveryResult{
		Metadata:          NewMetadata(result),
		NetworkNamespaces: newNetworkNamespace(result.Netns),
		PIDNamespaces:     pidNamespaces(result.Lxkns.Namespaces[model.PIDNS]),
	}
}
