// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { styled } from "@mui/material"
import ProcessIcon from 'icons/Process'

const Details = styled('span')(() => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
}))

const Cmdline = styled('span')(() => ({
    maxWidth: '16em',
    // "overflow: hidden" needs either a block or inline-block, but in our
    // case we need an inline-block.
    display: 'inline-block',
    // Now, "overflow: hidden" will cause the alignment to switch from
    // baseline to bottom; but we need it to align with the same top as the
    // following text, so it's vertical alignment to the top for us in this
    // situation. Oh, well...
    verticalAlign: 'top',
    // Clip the command if it grows too long, and simply put in an ellipsis.
    whiteSpace: 'nowrap',
    overflow: 'hidden',
    textOverflow: 'ellipsis',
}))

const PID = styled('span')(() => ({
    // Keep the same alignment as for the (potentially clipped) command
    // information, as otherwise the rendered outcome will just suck.
    display: 'inline-block',
    verticalAlign: 'top',
}))

/** Return last path component of first command line element. */
const command = (cmdline: string[]) => {
    const name = cmdline[0].split('/')
    return name[name.length - 1] + ' ' + cmdline.slice(1).join(' ')
}

export interface ProcessDetailsProps {
    cmdline: string[]
    pid: number
}

export const ProcessDetails = ({ cmdline, pid }: ProcessDetailsProps) => (
    <Details>
        <ProcessIcon fontSize="inherit" />
        <Cmdline>{command(cmdline)}</Cmdline>
        <PID>&nbsp;({pid})</PID>
    </Details>
)
