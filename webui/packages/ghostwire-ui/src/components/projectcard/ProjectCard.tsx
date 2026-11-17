// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { alpha, Avatar, Card, CardContent, CardHeader, getLuminance, getOverlayAlpha, styled, type Theme, Tooltip, useTheme } from '@mui/material'

import AppProjectIcon from '../../icons/groups/AppProject'
import ComposerProjectIcon from '../../icons/groups/ComposerProject'

import { ContainerFlavors, GHOSTWIRE_LABEL_ROOT, type Project, projectDescription } from '../../models/gw'
import { lighten } from '@mui/material'
import { useImageLuminance } from '../../hooks/imageluminance'

import { ProjectContext } from './context'


// Minimum contrast ratio between an app icon and its avatar background; see
// also WCAG 2.x "non-text contrast".
const MIN_ICON_CONTRAST = 3

/**
 * Returns true in dark mode if the specified icon luminance doesn't
 * sufficiently contrast with the theme's default background.
 */
const needsLighterIconBackground = (theme: Theme, iconLuminance: number | undefined) => {
    if (theme.palette.mode !== 'dark' || iconLuminance === undefined) {
        return false
    }
    const bgLuminance = getLuminance(theme.palette.background.default)
    const contrast = (Math.max(iconLuminance, bgLuminance) + 0.05)
        / (Math.min(iconLuminance, bgLuminance) + 0.05)
    return contrast < MIN_ICON_CONTRAST
}

// Background overlay of an elevation 1 Paper in dark mode.
const paperOverlay = `linear-gradient(${alpha('#fff', getOverlayAlpha(1))}, ${alpha('#fff', getOverlayAlpha(1))})`


const ProjCard = styled(Card)(({ theme }) => ({
    backgroundColor: theme.palette.background.default,
    borderStyle: 'dashed',
    borderWidth: 2,

    '& .MuiAvatar-root': {
        backgroundColor: lighten('#1993ef', 0.6),
    },
    '&.flavor-ie-app .MuiAvatar-root': {
        backgroundColor: lighten('#009999', 0.6),
    },
    '&.flavor-ie-app .MuiAvatar-root.ieapp-icon': {
        borderColor: theme.palette.divider,
        borderWidth: 1,
        borderStyle: "solid",
        backgroundColor: theme.palette.background.default,
    },
    '&.flavor-ie-app .MuiAvatar-root.ieapp-icon.low-contrast': {
        backgroundColor: theme.palette.background.paper,
        backgroundImage: paperOverlay,
    },
    '& > .MuiCardContent-root': {
        paddingTop: 0,
        paddingBottom: theme.spacing(2),
        paddingLeft: theme.spacing(2),
        paddingRight: theme.spacing(2),
    },
    '& > .MuiCardContent-root > *': {
        margin: `${theme.spacing(2)} 0`,
    },
    '& > .MuiCardContent-root > div:first-of-type': {
        marginTop: 0,
    },
    '& > .MuiCardContent-root > div:last-of-type': {
        marginBottom: 0,
    },
}))

const AppTitle = styled("span")(()=>({
    fontStyle: "italic",
    paddingRight: "0.1em",
}))


export interface ProjectCardProps {
    /** the project. */
    project: Project
    /** children to render within the content pane. */
    children: React.ReactNode
}

/**
 * `Project` renders a (Docker) composer project group.
 */
const ProjectCard = ({ project, children }: ProjectCardProps) => {

    const isIEApp = project.flavor === ContainerFlavors.IEAPP
    const ProjectIcon = isIEApp ? AppProjectIcon : ComposerProjectIcon
    const iconData = (isIEApp && project.containers[0].labels[GHOSTWIRE_LABEL_ROOT+'icon']) || null
    const appTitle = project.containers[0].labels[GHOSTWIRE_LABEL_ROOT+'ieapp/title']
    const theme = useTheme()
    const lowContrast = needsLighterIconBackground(theme, useImageLuminance(iconData))

    return (
        <ProjCard
            variant="outlined"
            className={clsx(project.flavor !== '' && `flavor-${project.flavor}`)}
        >
            <CardHeader
                title={<>{project.name}{appTitle ? <> (<AppTitle>{appTitle}</AppTitle>)</> : ""}</>}
                avatar={
                    <Tooltip title={projectDescription(project)}>
                        {(isIEApp && iconData)
                            ? <Avatar
                                className={clsx('ieapp-icon', lowContrast && 'low-contrast')}
                                variant="rounded"
                                src={iconData}
                            />
                            : <Avatar variant="rounded"><ProjectIcon /></Avatar>
                        }
                    </Tooltip>
                }
            />
            <CardContent>
                <ProjectContext.Provider value={project}>
                    {children}
                </ProjectContext.Provider>
            </CardContent>
        </ProjCard>
    )

}

export default ProjectCard
