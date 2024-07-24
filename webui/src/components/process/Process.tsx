// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { styled } from '@mui/material'
import { ProcessDetails } from "components/procdetails"
import { PrimitiveContainee, containeeDisplayName, isContainer } from "models/gw"
import { ContaineeIcon } from "utils/containeeicon"

const Details = styled('span')(({ theme }) => ({
    whiteSpace: 'nowrap',

    '& .MuiSvgIcon-root': {
        marginRight: '0.1em',
        verticalAlign: 'baseline',
        position: 'relative',
        top: '0.2ex',
        color: theme.palette.text.disabled,
    },
}))

const DetailsOfProcess = styled('span')(() => ({
    display: 'inline-block',
    whiteSpace: 'nowrap',
}))


export interface ProcessProps {
    /** command line of the process using the port. */
    cmdline: string[]
    /** the containee this process using the port belongs to. */
    containee: PrimitiveContainee
    /** PID of process using the port. */
    pid: number
}

/**
 * Renders a process, together with its group (pod, ...) and container
 * information, if any.
 */
export const Process = ({ cmdline, containee, pid }: ProcessProps) => {
    const info: (string | JSX.Element | (string | JSX.Element)[])[] = []

    // Good gracious! That took a long time to figure out that this seemingly
    // function is a source of non-unique keys, grmpf. Adding keys to each and
    // every array item finally silences the warnings.

    if (containee) {
        if (isContainer(containee) && containee.pod) {
            // This is a "pot'ed" container...
            const CI = ContaineeIcon(containee)
            info.push([<CI key="pod" fontSize="inherit" />, containee.pod.name])
        }
        // Add the container details...
        const CI = ContaineeIcon(containee)
        info.push([<CI key="containee" fontSize="inherit" />, containeeDisplayName(containee)])
    }
    // And finally: the process details ... cmdline and PID.
    info.push(
        <DetailsOfProcess key="process"><ProcessDetails cmdline={cmdline} pid={pid} /></DetailsOfProcess>
    )
    // Finally return the detail elements, separated by commas; and no, we can't
    // use Array.join() here, as we face JSX elements.
    return <Details>
        {info.reduce((list, element, index) => {
            if (index) {
                list.push(<span key={index}> · </span>)
            }
            list.push(element)
            return list
        }, [] as (string | JSX.Element | (string | JSX.Element)[])[]).flat()}
    </Details>

} 