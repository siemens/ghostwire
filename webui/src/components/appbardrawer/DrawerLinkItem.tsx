// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { useLocation, Link } from "react-router-dom"

import { Avatar, ListItem, ListItemAvatar, ListItemIcon, Typography } from '@mui/material'

export interface DrawerLinkItemProps {
    /** 
     * drawer item icon, which automatically will be enclosed in ListItemIcon
     * components.
     */
    icon?: React.ReactNode
    /** render the drawer item icon as an avatar. */
    avatar?: boolean
    /** label of the drawer item. */
    label: React.ReactNode | string
    /** route path to activate when the user clicks on this drawer item. */
    path: string
}

/**
 * `DrawerLinkItem` renders an individual item inside an
 * [`AppBarDrawer`](#appbardrawer) and links it to a specific route path. It
 * is a convenience component that simplifies describing the drawer items with
 * their icons and route paths.
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
export const DrawerLinkItem = ({ icon, avatar, label, path }: DrawerLinkItemProps) => {

    const location = useLocation()
    const selected = location.pathname === path

    return (
        <ListItem
            button
            component={Link}
            to={path}
            selected={selected}
        >
            {(avatar && icon &&
                <ListItemAvatar><Avatar>{icon}</Avatar></ListItemAvatar>
            ) || (
                icon && <ListItemIcon>{icon}</ListItemIcon>
            )}
            <Typography>{label}</Typography>
        </ListItem>
    )
}
