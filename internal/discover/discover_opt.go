// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

package discover

// DiscoverOpts provides information about the extent of a Ghostwire discovery.
type DiscoverOpts struct {
	ScanSockets bool
}

// DiscoveryOption represents a function able to set a particular discovery
// option state in [DiscoverOpts].
type DiscoveryOption func(*DiscoverOpts)

func WithStandardDiscovery() DiscoveryOption {
	return func(o *DiscoverOpts) {
		o.ScanSockets = true
	}
}

func WithSockets() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanSockets = true }
}

func WithoutSockets() DiscoveryOption {
	return func(o *DiscoverOpts) { o.ScanSockets = false }
}
