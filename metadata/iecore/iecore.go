// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package iecore

import (
	"sync"

	gostwire "github.com/siemens/ghostwire/v2"
	"github.com/siemens/ghostwire/v2/metadata"
	"github.com/siemens/ieddata"

	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator/industrialedge"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/lxkns/ops/mountineer"
	"github.com/thediveo/osrelease"
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

func init() {
	plugger.Group[metadata.Metadata]().Register(
		Metadata, plugger.WithPlugin("iecore"))
}

var once = &sync.Once{} // We need to query this only once; no hot core update (yet)
var coreMeta = map[string]interface{}{}

// Metadata returns metadata describing certain aspects of the host the
// discovery was run on.
func Metadata(r gostwire.DiscoveryResult) map[string]interface{} {
	once.Do(func() {
		coreMeta = gatherMetadata(r)
	})
	return coreMeta
}

// gatherMetadata about the Industrial Edge core/runtime, if present.
func gatherMetadata(r gostwire.DiscoveryResult) map[string]interface{} {
	// Is an Industrial Edge core runtime container present...?
	core := findEdgeCoreContainer(r)
	if core == nil {
		return nil
	}
	// Get the sem version of the edge core...
	osrel := readEdgeCoreContainerOsrelease(core)
	iedata := map[string]string{
		"semversion": osrel["VERSION_ID"],
	}
	if devInfo := deviceInfo(core); devInfo != nil {
		if devname := devInfo["deviceName"]; devname != "" {
			iedata["device-name"] = devname
		}
		if devmode := devInfo["developerMode"]; devmode != "" {
			iedata["developer-mode"] = devmode
		}
	}
	// Finally return the metadata as its own individual "industrial-edge"
	// JSON object.
	return map[string]interface{}{
		"industrial-edge": iedata,
	}
}

// deviceInfo returns the device information (key-value pairs) managed by an
// Industrial Edge core/runtime, if present. Otherwise, returns nil.
func deviceInfo(cc *model.Container) map[string]string {
	if cc.Process == nil {
		return nil
	}
	db, err := ieddata.OpenInPID(ieddata.PlatformBoxDb, cc.Process.PID)
	if err != nil {
		return nil
	}
	defer db.Close() // three cheers to fdooze!
	kv, _ := db.DeviceInfo()
	return kv
}

// readEdgeCoreContainerOsrelease reads /etc/os-release-container inside the
// Industrial Edge core/runtime container, returning the information found as a
// map with OS-release variable names found as keys, together with their values.
// Any broken "-e VERSION_ID" keys will automatically be copied over into
// "VERSION_ID" if the latter isn't present.
func readEdgeCoreContainerOsrelease(cc *model.Container) map[string]string {
	if cc.Process == nil || cc.Process.Namespaces[model.MountNS] == nil {
		log.Errorf("no mount namespace information available for edge-iot-core")
		return nil
	}
	corefs, err := mountineer.New(cc.Process.Namespaces[model.MountNS].Ref(), nil)
	if err != nil {
		log.Errorf("cannot access mount namespace of edge-iot-core: %s", err.Error())
		return nil
	}
	defer corefs.Close()
	// Read OS release information inside the edge core container, from its
	// current mount namespace.
	osrelPath, err := corefs.Resolve("/etc/os-release-container")
	if err != nil {
		log.Errorf("cannot resolve /etc/os-release-container path in edge-iot-core: %s", err.Error())
	}
	osrel := osrelease.NewFromName(osrelPath)
	// Fix broken VERSION_ID entries...
	if versionID, ok := osrel["-e VERSION_ID"]; ok && osrel["VERSION_ID"] == "" {
		osrel["VERSION_ID"] = versionID
	}
	return osrel
}

// findEdgeCoreContainer locates and returns the Industrial Edge core/runtime
// container, if present. Otherwise, returns nil.
func findEdgeCoreContainer(r gostwire.DiscoveryResult) *model.Container {
	for _, container := range r.Lxkns.Containers {
		if container.Type == moby.Type &&
			container.Flavor == industrialedge.IndustrialEdgeRuntimeFlavor {
			return container
		}
	}
	return nil
}
