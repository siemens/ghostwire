// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { Avatar, Card, CardContent, CardHeader, styled, Tooltip } from '@mui/material'

import AppProjectIcon from 'icons/groups/AppProject'
import ComposerProjectIcon from 'icons/groups/ComposerProject'

import { ContainerFlavors, GHOSTWIRE_LABEL_ROOT, Project, projectDescription } from 'models/gw'
import { lighten } from '@mui/material'


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

const AppTitle = styled("span")(({theme})=>({
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
                            ? <Avatar className="ieapp-icon" variant="rounded" src={iconData} />
                            : <Avatar variant="rounded"><ProjectIcon /></Avatar>
                        }
                    </Tooltip>
                }
            />
            <CardContent>
                {children}
            </CardContent>
        </ProjCard>
    )

}

export default ProjectCard
