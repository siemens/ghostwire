// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/siemens/ieddata"
	"github.com/thediveo/go-plugger/v3"
	"github.com/thediveo/lxkns/decorator"
	"github.com/thediveo/lxkns/decorator/industrialedge"
	"github.com/thediveo/lxkns/log"
	"github.com/thediveo/lxkns/model"
	"github.com/thediveo/procfsroot"
	"github.com/thediveo/whalewatcher/engineclient/moby"
)

// IconLabel is the name of a label storing a container's icon in form of a
// data: URL, if any.
const IconLabel = "gostwire/icon"

// IEAppTitleLabel is the name of a label storing a container's IE App title.
const IEAppTitleLabel = "gostwire/ieapp/title"

// IEAppDebuggableLabel is new as of IE runtime/core 1.8 and is set to the value
// of the isDebuggingEnabled column of an App if not "0" or missing.
// IEAppDebuggableLabel is an integer in stringified form.
const IEAppDebuggableLabel = "gostwire/ieapp/debuggable"

// IEAppDiscoveryLabel is the name of a discovery options label to enable (or
// disable) the ieappicon plugin for a specific discovery. If a label of this
// name is present in the options of a discovery and its value isn't "off" then
// this decorator will be used in that particular discovery.
const IEAppDiscoveryLabel = "ieappicons"

// IEAppDiscoveryOff is the "magic" value of an IEAppDiscoveryLabel to
// explicitly turn off this plugin for a specific discovery. Please note that
// this plugin defaults to being inactive for any discovery unless the
// IEAppDiscoveryLabel is present in the discovery options and its value isn't
// IEAppDiscoveryOff. The rationale here is that it is still possible to disable
// this plugin on a discovery, even if the magic label is present.
const IEAppDiscoveryOff = "off"

// EnablePlugin enables this plugin; by default, it is enabled but still needs
// to be triggered during discovery by passing a discovery label with the name
// "ieappicons" and value other than "off".
var EnablePlugin = true

// Name of the IE edge core runtime database storing information about the
// installed apps.
const platformboxdbName = "platformbox.db"

var appIcons = ieAppProjects{} // App project icons cache
var appIconsMu sync.Mutex      // TODO: improve locking

// Register this (lxkns!) Decorator plugin.
func init() {
	plugger.Group[decorator.Decorate]().Register(
		Decorate, plugger.WithPlugin("industrialedge-appicon"))
}

// Decorate all IE App containers with their app project icons; the icons are
// stored in the container labels in form of data: URLs of base64-encoded
// png-type images.
func Decorate(engines []*model.ContainerEngine, labels map[string]string) {
	// Skip this (plugin) decorator if it either has been disabled globally or
	// not requested explicitly.
	if label, ok := labels[IEAppDiscoveryLabel]; !ok || label == IEAppDiscoveryOff || !EnablePlugin {
		log.Debugf("skipping ieappicon decorator because it hasn't been requested in this discovery")
		return
	}
	appIconsMu.Lock()
	news := appIcons.pruneAndUpdate(engines)
	if len(news) > 0 {
		// Try to load IE App project icons, where available; we unconditionally
		// add projects without icons as negative hits as to avoid further
		// lookups on subsequent discoveries.
		loadProjectIcons(engines, news)
		appIcons.add(news)
	}
	appIconsMu.Unlock()
	// Go over all IE app-related containers and set their icons, if any.
	for _, engine := range engines {
		for _, container := range engine.Containers {
			if container.Flavor != industrialedge.IndustrialEdgeAppFlavor {
				continue
			}
			projectName := container.Labels[moby.ComposerProjectLabel]
			if projectName == "" {
				continue
			}
			if iconData := appIcons.icon(projectName); iconData != "" {
				container.Labels[IconLabel] = iconData
			}
			if title := appIcons.title(projectName); title != "" {
				container.Labels[IEAppTitleLabel] = title
			}
			if debuggable := appIcons.debuggable(projectName); debuggable != "" && debuggable != "0" {
				container.Labels[IEAppDebuggableLabel] = debuggable
			}
		}
	}
}

// loadProjectIcons loads the icon (data) for the specified IE App projects. It
// needs the engines information in order to locate the IED core runtime
// container in order to access its installed applications database.
func loadProjectIcons(engines []*model.ContainerEngine, projects []ieAppProject) {
	log.Debugf("discovering IE App icons")
	// First, locate the Edge Core and fetch its platformbox database, as that
	// tells us more details about the currently installed IE Apps.
	edgeCorePID := edgeCoreContainerPID(engines)
	if edgeCorePID == 0 {
		return
	}
	db, err := ieddata.OpenInPID(platformboxdbName, edgeCorePID)
	if err != nil {
		log.Errorf("cannot access IED IE App data base: %s", err.Error())
		return
	}
	defer db.Close()
	apps, err := db.Apps()
	if err != nil {
		log.Errorf("cannot discover installed IE Apps: %s", err.Error())
		return
	}
	// Next, build an index that maps (composer) project names to IE App
	// information found in the platformbox database.
	appIndex := map[string]*ieddata.App{}
	for idx, app := range apps {
		// The app's repository name is a single filesystem-compatible name
		// (without any slashes, dashes, ...) that also doubles as a path
		// element of the docker-compose.yaml file somewhere in the host's file
		// system after an app has been installed.
		if app.RepositoryName == "" {
			continue
		}
		appIndex[app.RepositoryName] = &apps[idx]
	}
	// Finally, update the projects specified by the caller with their
	// associated icons, if found.
	for idx := range projects {
		app, ok := appIndex[projects[idx].Name]
		if !ok {
			continue
		}
		iconData := loadAppIcon(app, edgeCorePID)
		projects[idx].IconData = iconData
		if iconData != "" {
			log.Debugf("found icon for IE App project '%s'", projects[idx].Name)
		}
		projects[idx].Title = app.Title
		projects[idx].Debuggable = strconv.Itoa(app.IsDebuggingEnabled)
	}
}

// loadAppIcon loads the icon of the specified IE App and returns it in data:
// URL format, if available; otherwise, returns a zero string.
func loadAppIcon(app *ieddata.App, edgeCorePID model.PIDType) string {
	// The icon URLs are locking something like
	// file://.../device/edge/BoxCache/app/<ID>/<truncatedname>
	iconUrl, err := url.Parse(app.IconPath)
	if err != nil {
		log.Errorf("IE App with invalid icon path '%s': %s", app.IconPath, err.Error())
		return ""
	}
	fields := strings.Split(iconUrl.Path, "/")
	if len(fields) < 4 {
		log.Errorf("IE App with not enough segments in icon path '%s'", app.IconPath)
		return ""
	}
	// Regenerate the icon path in a way that we can access it inside the
	// Edge core runtime container.
	iconPath := "/data/BoxCache/app/" + fields[len(fields)-2] + "/" + fields[len(fields)-1]
	root := fmt.Sprintf("/proc/%d/root/", edgeCorePID)
	path, err := procfsroot.EvalSymlinks(iconPath, root, procfsroot.EvalFullPath)
	if err != nil {
		log.Errorf("IE App with core-relative icon path '%s': %s", iconPath, err.Error())
		return ""
	}
	iconBytes, err := ioutil.ReadFile(root + path)
	if err != nil {
		log.Errorf("cannot read IE App icon file '%s': %s", root+path, err.Error())
		return ""
	}
	mimetype := http.DetectContentType(iconBytes)
	switch mimetype {
	case "image/png":
		break
	default:
		log.Errorf("invalid mime type '%s' for IE App icon '%s'",
			mimetype, root+path)
		return ""
	}
	log.Debugf("loaded IE APP icon (type '%s') from '%s'", mimetype, root+path)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(iconBytes)
}

// edgeCoreContainerPID returns the PID of the IED core runtime container, if
// present. Otherwise returns a zero PID.
func edgeCoreContainerPID(engines []*model.ContainerEngine) model.PIDType {
	for _, engine := range engines {
		if engine.Type != moby.Type {
			continue
		}
		for _, container := range engine.Containers {
			if container.Name == ieddata.EdgeIotCoreContainerName {
				return container.PID
			}
		}
	}
	return 0
}
