// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'
import { Process } from 'models/gw/process'
import { HistoryToggleOff } from '@mui/icons-material'
import { rgba } from 'utils/rgba'


const SchedInformation = styled('span')(({ theme }) => ({
    '& .policy': {
        //fontSize: '80%',
    },
    '& .normal,& .batch,& .idle': {
        color: theme.palette.sched.relaxed,
    },
    '& .fifo,& .rr': {
        color: theme.palette.sched.stressed,
    },
    '& .nice': {
        color: theme.palette.sched.nice,
    },
    '& .notnice': {
        color: theme.palette.sched.notnice,
    },
    '& .prio': {
        color: theme.palette.sched.prio,
    },
    '& > .MuiSvgIcon-root': {
        color: rgba(theme.palette.text.primary, 0.1),
        verticalAlign: 'text-top',
        position: 'relative',
        top: '-0.15ex',
        marginRight: '0.2em',
    },
}))

const schedulerPolicies: { [key: string]: string } = {
    '0': 'NORMAL',
    '1': 'FIFO',
    '2': 'RR',
    '3': 'BATCH',
    '5': 'IDLE',
    '6': 'DEADLINE',
}

const hasPriority = (process: Process) => {
    const policy = process.policy || 0
    return policy === 1 || policy === 2
}

const hasNice = (process: Process) => {
    const policy = process.policy || 0
    return policy === 0 || policy === 3
}

export interface SchedulerInfoProps {
    /** information about a discovered Linux OS process. */
    process: Process
    /** also schow (SCHED_) NORMAL? */
    showNormal?: boolean
}

export const SchedulerInfo = ({ process, showNormal }: SchedulerInfoProps) => {
    const schedpol = schedulerPolicies[process.policy || 0] || `policy #${process.policy}`
    const prio = process.priority || 0
    return <SchedInformation className="schedinfo">
        <HistoryToggleOff fontSize="small" />
        {(showNormal || !!process.policy) && <span className={clsx('policy', schedpol.toLowerCase())}>{schedpol}</span>}
        {hasPriority(process) && <span className={clsx(prio > 0 && 'prio')}>&nbsp;priority {prio}</span>}{
            hasNice(process) && !!process.nice &&
            <Tooltip title={process.nice >= 0 ? 'nice!' : 'not nice'}>
                <span className={process.nice >= 0 ? 'nice' : 'notnice'}>&nbsp;nice {process.nice}</span>
            </Tooltip>}
    </SchedInformation>
}

export default SchedulerInfo
