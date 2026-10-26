// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'
import { type Project } from 'models/gw'

export const ProjectContext = React.createContext<Project | undefined>(undefined)

/**
 * Returns the project of the enclosing `ProjectCard`, if any; this allows
 * descendants to skip rendering project information already shown by the card.
 */
export const useEnclosingProject = () => useContext(ProjectContext)
