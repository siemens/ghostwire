// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package host

import (
	"os"
	"sync"

	"github.com/siemens/ghostwire/v2/metadata"

	gostwire "github.com/siemens/ghostwire/v2"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/osrelease"
)

func init() {
	plugger.Group[metadata.Metadata]().Register(
		Metadata, plugger.WithPlugin("host"))
}

var once sync.Once
var hostMeta = map[string]interface{}{}

// Metadata returns metadata describing certain aspects of the host the
// discovery was run on, such as its host name, OS version, ...
func Metadata(r gostwire.DiscoveryResult) map[string]interface{} {
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
		log.Errorf("cannot access host mount namespace: %s", err.Error())
		return nil
	}
	defer hostfs.Close()
	osrelPath, err := hostfs.Resolve("/etc/os-release")
	if err != nil {
		log.Errorf("cannot resolve /etc/os-release-container host path: %s", err.Error())
		return nil
	}
	vars, err := osrelease.NewFromNameErr(osrelPath)
	if err != nil {
		log.Warnf("cannot fetch OS release information, reason: %s", err.Error())
	}
	log.Debugf("OS information...")
	for key, value := range vars {
		log.Debugf("  %s: %s", key, value)
	}
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
