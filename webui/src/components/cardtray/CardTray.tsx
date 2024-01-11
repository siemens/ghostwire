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
export const CardTray = ({ animate, ...otherprops }: React.PropsWithChildren<CardTrayProps>) => {
    if (animate) {
        const c = React.Children.map(
            otherprops?.children, 
            child => <Collapse key={(child as {key: Key}).key} in>{child}</Collapse>)
        return (
            <TransitionGroup component={Surface}>
                {c || undefined}
            </TransitionGroup>
        )
    }
    return <Surface>{otherprops?.children}</Surface>
}
