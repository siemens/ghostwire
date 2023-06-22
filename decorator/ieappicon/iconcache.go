// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package ieappicon

import (
	"github.com/thediveo/lxkns/decorator/composer"
	"github.com/thediveo/lxkns/decorator/industrialedge"
	"github.com/thediveo/lxkns/model"
)

// ieAppProject represents information relevant to us for app icon decoration,
// including negative finds.
type ieAppProject struct {
	Name         string   // composer project name, derived from db's composerFilePath
	IconData     string   // icon in data: URL format
	ContainerIDs []string // known alive containers for this project
	Title        string   // from installed app's meta data
	Debuggable   string   // from installed app's meta data
}

// ieAppProjects caches the known IE App projects with some associated
// information.
type ieAppProjects map[string]ieAppProject

// icon returns the icon in data: URL form suitable for use in HTML img tags, et
// cetera. Returns a zero string if project isn't known or doesn't have any
// icon.
func (ap ieAppProjects) icon(projectname string) string {
	return ap[projectname].IconData
}

// title returns the title of the IE App identified by projectname.
func (ap ieAppProjects) title(projectname string) string {
	return ap[projectname].Title
}

// debuggable returns the debuggable flag of the IE App identified by
// projectname. For IE runtimes before 1.8 the flag is always "".
func (ap ieAppProjects) debuggable(projectname string) string {
	return ap[projectname].Debuggable
}

// adds a list of new IE App projects to the cache. The caller should have
// loaded the icon data (if any) beforehand.
func (ap ieAppProjects) add(news []ieAppProject) {
	for _, appProject := range news {
		ap[appProject.Name] = appProject
	}
}

// pruneAndUpdate prunes an AppProjects cache from projects for which no more
// alive containers exist and return newly found app projects. Already known
// projects are updated with respect to their workload container IDs, as long as
// there is still at least one container the same as before. Containers are
// identified by their IDs, not by their (functional) names.
func (ap ieAppProjects) pruneAndUpdate(engines []*model.ContainerEngine) []ieAppProject {
	// Phase I: build a map of current IE App projects with their alive
	// containers...
	projContainers := map[string][]string{} // ...map project name to alive container IDs
	for _, engine := range engines {
		for _, container := range engine.Containers {
			// We work only on Industrial Edge containers with an IE App
			// project, so make sure this container is one of them.
			if container.Flavor != industrialedge.IndustrialEdgeAppFlavor {
				continue
			}
			project := container.Group(composer.ComposerGroupType)
			if project == nil {
				continue
			}
			// Remember yet another alive container for a specific IE App
			// project.
			projContainers[project.Name] = append(projContainers[project.Name],
				container.ID)
		}
	}
	// Phase II: remove all projects from the cache for which we don't see any
	// alive containers anymore. Update the list of alive containers for the
	// other projects.
	for projectName, appProject := range ap {
		if containerIDs, ok := projContainers[projectName]; ok {
			// This project is still alive, so update its cache entry ... but
			// consider it to be alive only, if there is at least one container
			// still the same as before. Otherwise, consider this project to be
			// a new project: the project might have been updated and all
			// containers started anew.
			//
			// Caveat emptor: we might have hit the device data base at a time
			// where app project information was incomplete, so typically both
			// title and icon data are unknown (empty). In this case, do not
			// update the project's container list, but instead purge the
			// project, so it's device data base information will be fetched
			// anew ... and hopefully complete this time.
			if appProject.Title != "" && appProject.IconData != "" && !isNewProject(appProject.ContainerIDs, containerIDs) {
				appProject.ContainerIDs = containerIDs
				ap[projectName] = appProject
				continue
			}
		}
		// Prune this project from the cache.
		delete(ap, projectName)
	}
	// Phase III: determine newly found projects
	newProjects := []ieAppProject{}
	for projectName, containerIDs := range projContainers {
		if _, ok := ap[projectName]; ok {
			continue
		}
		newProjects = append(newProjects, ieAppProject{
			Name:         projectName,
			ContainerIDs: containerIDs,
		})
	}
	return newProjects
}

// isNewProject returns true if, and only if there are not a single matching
// container ID in the before and now lists of container IDs.
func isNewProject(befores, nows []string) bool {
	for _, before := range befores {
		for _, now := range nows {
			if before == now {
				return false
			}
		}
	}
	return true
}
