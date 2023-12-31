// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

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

const DetailsOfProcess = styled('span')(({ theme }) => ({
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
    let info = []

    // Good gracious! That took a long time to figure out that this seemingly
    // function is a source of non-unique keys, grmpf. Adding keys to each and
    // every array item finally silences the warnings.

    if (!!containee) {
        if (isContainer(containee) && containee.pod) {
            // This is a "pot'ed" container...
            info.push([ContaineeIcon(containee.pod)({ key: 'pod', fontSize: 'inherit' }), containee.pod.name])
        }
        // Add the container details...
        info.push([ContaineeIcon(containee)({ key: 'containee', fontSize: 'inherit' }), containeeDisplayName(containee)])
    }
    // And finally: the process details ... cmdline and PID.
    info.push(
        <DetailsOfProcess><ProcessDetails key="process" cmdline={cmdline} pid={pid} /></DetailsOfProcess>
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
        }, []).flat()}
    </Details>

} 