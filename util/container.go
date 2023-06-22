// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package util

import (
	"sort"
	"strings"

	"github.com/siemens/ghostwire/v2/network"

	"github.com/thediveo/lxkns/decorator/kuhbernetes"
	"github.com/thediveo/lxkns/model"
)

// FindContainer returns the container matching the specified name and type.
func FindContainer(allnetns network.NetworkNamespaces, name string, typ string) *model.Container {
	for _, netns := range allnetns {
		for _, tenant := range netns.Tenants {
			container := tenant.Process.Container
			if container != nil && container.Name == name && container.Type == typ {
				return container
			}
		}
	}
	return nil
}

// AllContainerNamesWithGroups lists the names of all containers and optionally
// includes the k8s container name as well as any group names a container
// belongs to.
func AllContainerNamesWithGroups(containers []*model.Container) []string {
	names := make([]string, 0, len(containers))
	for _, container := range containers {
		name := container.Name
		if cntrname := container.Labels[kuhbernetes.PodContainerNameLabel]; cntrname != "" {
			name += "(" + cntrname + ")"
		}
		if len(container.Groups) > 0 {
			groups := make([]string, 0, len(container.Groups))
			for _, group := range container.Groups {
				groups = append(groups, group.Name)
			}
			name += "[" + strings.Join(groups, ",") + "]"
		}
		names = append(names, name)
	}
	return names
}

// AllPods returns the names of all k8s pods the specified containers belong to
// (if any).
func AllPods(containers []*model.Container) []string {
	podIndex := map[string]struct{}{}
	names := []string{}
	for _, container := range containers {
		names = append(names, container.Name)
		for _, group := range container.Groups {
			if group.Type != kuhbernetes.PodGroupType {
				continue
			}
			podIndex[group.Name] = struct{}{}
		}
	}
	pods := make([]string, 0, len(podIndex))
	for podname := range podIndex {
		pods = append(pods, podname)
	}
	sort.Strings(pods)
	return pods
}
