// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { Key } from 'react'
import { TransitionGroup } from 'react-transition-group'

import { Collapse, styled } from '@mui/material'


const Surface = styled('div')(({ theme }) => ({
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'stretch',
    backgroundColor: theme.palette.background.default,

    '& > *': {
        margin: theme.spacing(1),
    }
}))


export interface CardTrayProps {
    /** children to render into a vertical list. */
    children?: React.ReactNode
    /** animate new children expanding? */
    animate?: boolean
}

/**
 * A `CardTray` renders its children (for instance, Material UI's `Card`
 * components) into a vertical list, optionally animating these children
 * (cards) to expand or collapse as they get added or removed. To provide
 * enough visual differentiation between the background and the cards, a
 * `CardTray` renders its own background using the current theme's default
 * background, instead of white or paper.
 *
 * @param children cards to render inside the card panel. 
 */
export const CardTray = ({ children, animate }: CardTrayProps) => {
    if (animate) {
        return (
            <TransitionGroup component={Surface}>
                {React.Children.map(children, child => <Collapse key={(child as {key: Key | null | undefined}).key} in>{child}</Collapse>)}
            </TransitionGroup>
        )
    }
    return <Surface>{children}</Surface>
}
