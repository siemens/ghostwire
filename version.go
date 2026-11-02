// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

package gostwire

import (
	"runtime/debug"
	"strings"
)

const modulePath = "github.com/siemens/ghostwire/v2"

// SemVersion is the semantic version string of the ghostwire module. Release
// builds set it using “-ldflags -X”, otherwise it is taken from the Go build
// information.
var SemVersion string

// readBuildInfo is replaceable so tests can supply fake build information.
var readBuildInfo = debug.ReadBuildInfo

func init() {
	if SemVersion == "" {
		SemVersion = buildInfoVersion()
	}
}

// buildInfoVersion returns the ghostwire module version from the binary's build
// information, regardless of ghostwire being the main module or a dependency.
func buildInfoVersion() string {
	const devel = "(devel)"
	bi, ok := readBuildInfo()
	if !ok {
		return devel
	}
	for _, mod := range append([]*debug.Module{&bi.Main}, bi.Deps...) {
		if mod.Path != modulePath {
			continue
		}
		if mod.Replace != nil {
			mod = mod.Replace
		}
		if mod.Version == "" || mod.Version == devel {
			return devel
		}
		return strings.TrimPrefix(mod.Version, "v")
	}
	return devel
}
