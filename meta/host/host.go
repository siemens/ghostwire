// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package host

import (
	"log/slog"
	"os"
	"sync"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/nonstd/xslog"
	"github.com/thediveo/osrelease"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/meta"
)

func init() {
	plugger.Group[meta.Producer]().Register(
		Metadata, plugger.WithPlugin("host"))
}

var once sync.Once
var hostMeta = meta.Data{}

// Metadata returns metadata describing certain aspects of the host the
// discovery was run on, such as its host name, OS version, ...
func Metadata(r gostwire.DiscoveryResult) meta.Data {
	once.Do(func() {
		hostMeta["hostname"] = getHostname(r)
		if osrelvars := getHostOsrelVars(); osrelvars != nil {
			hostMeta["osrel-name"] = osrelvars["NAME"]
			hostMeta["osrel-version"] = osrelvars["VERSION"]
		}
		if kv := getKernelVersion(); kv != "" {
			hostMeta["kernel-version"] = kv
		}
	})
	return hostMeta
}

// getHostrelVars fetches the host's os-release variables. It does so by reading
// from the filesystem view as manifested through the initial mount namespace
// (of PID 1). It returns a nil variables map on failure.
func getHostOsrelVars() map[string]string {
	hostfs, err := mountineer.New(model.NamespaceRef{"/proc/1/ns/mnt"}, nil)
	if err != nil {
		slog.Error("cannot access host mount namespace", xslog.Error(err))
		return nil
	}
	defer hostfs.Close()
	osrelPath, err := hostfs.Resolve("/etc/os-release")
	if err != nil {
		slog.Error("cannot resolve /etc/os-release-container host path", xslog.Error(err))
		return nil
	}
	vars, err := osrelease.NewFromNameErr(osrelPath)
	if err != nil {
		slog.Warn("cannot fetch OS release information", xslog.Error(err))
	}
	slog.Debug("found OS information", slog.Any("vars", vars))
	return vars
}

// getHostname picks up the host name from a discovery's process PID 1 DNS
// configuration information.
func getHostname(r gostwire.DiscoveryResult) string {
	// Find the process with PID 1 and then look into its DNS configuration.
	for _, netns := range r.Netns {
		for _, tenant := range netns.Tenants {
			if tenant.Process.PID == 1 {
				// use the currently active hostname, not necessarily
				// /etc/hostname
				return tenant.DNS.Hostname
			}
		}
	}
	return ""
}

// getKernelVersion reads the kernel version string from /proc/version in the
// current mount namespace.
func getKernelVersion() string {
	if v, _ := os.ReadFile("/proc/version"); v != nil {
		return string(v)
	}
	return ""
}
