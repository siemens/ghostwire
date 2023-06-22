// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

package mobydig

import "github.com/thediveo/lxkns/model"

// Containers contains a bunch of containers.
type Containers []*model.Container

// Contains returns true if container c is contained in the set of containers.
// Equality of containers is simplified to require only the same type, name, and
// ID.
func (cs Containers) Contains(c *model.Container) bool {
	for _, cntr := range cs {
		if c.Name == cntr.Name && c.ID == cntr.ID && c.Type == cntr.Type {
			return true
		}
	}
	return false
}

// Shares returns true if this set of containers and the c container set contain
// at least one same container. Here, "same" means same type, name, and ID.
func (cs Containers) Shares(c []*model.Container) bool {
	for _, cntr := range c {
		if cs.Contains(cntr) {
			return true
		}
	}
	return false
}

// Merge containers from another set of containers with this container set and
// return a new merged set without any container duplicates.
func (cs Containers) Merge(c []*model.Container) Containers {
	cntrs := cs[:]
	for _, bcntr := range c {
		if cntrs.Contains(bcntr) {
			continue
		}
		cntrs = append(cntrs, bcntr)
	}
	return cntrs
}
