// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'
import clsx from 'clsx'

import { IconButton, Collapse, Divider, Typography, styled } from '@mui/material'
import ExpandLess from '@mui/icons-material/ExpandLess'
import ExpandMore from '@mui/icons-material/ExpandMore'


const CardSect = styled('div')(({ theme }) => ({
    // keep the divider fully width'ed under any right-floating zoom button,
    // et cetera.
    clear: 'both',

    '& + &': {
        marginTop: theme.spacing(2),
    }
}))


export interface CardSectionProps {
    /** section caption/header/title. */
    caption: string
    /** 
     * indicates whether section is collapsible and its initial state; defaults
     * to `never` if left undefined.
     */
    collapsible?: 'never' | 'collapse' | 'expand'
    /** CSS class name(s) for the div element enclosing the children. */
    className?: string
    /** optional fragment id. */
    fragment?: string
    /** children components to render as part of this section */
    children: React.ReactNode
}

/**
 * Component `CardSection` renders a section divider with a caption immediately
 * below it, as well as the children as the contents of the section.
 */
export const CardSection = ({ caption, collapsible, className, children, fragment }: CardSectionProps) => {
    collapsible = collapsible || 'never'
    const [expanded, setExpanded] = useState(collapsible === 'expand')

    const handleExpandClick = () => {
        setExpanded(!expanded)
    }

    return (
        <CardSect id={fragment}>
            <Divider />
            {(collapsible !== 'never') &&
                <IconButton
                    edge="start"
                    size="small"
                    onClick={handleExpandClick}
                >
                    {expanded ? <ExpandLess /> : <ExpandMore />}
                </IconButton>
            }
            <Typography variant="caption" color="textSecondary">{caption}</Typography>
            {(collapsible === 'never'
                && <div className={clsx(className, 'section')}>
                    {children}
                </div>)
                || <Collapse
                    className={clsx(className, 'section')}
                    in={expanded}
                    mountOnEnter={true}
                    timeout="auto"
                >
                    {children}
                </Collapse>
            }
        </CardSect>
    )
}
