// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import { Fragment } from 'react'

import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'
import CPUIcon from 'icons/CPUAffinity'
import { rgba } from 'utils/rgba'


const CPURangeList = styled('span')(({ theme }) => ({
    '&.nowrap, & .cpurange': {
        whiteSpace: 'nowrap',
    },
    '&.cpulist > .MuiSvgIcon-root': {
        color: rgba(theme.palette.text.primary, 0.1),
        verticalAlign: 'text-top',
        position: 'relative',
        top: '-0.15ex',
        marginRight: '0.2em',
    },
}))

export interface CPUListProps {
    /* list of CPU ranges */
    cpus: number[][] | null
    /* show/hide a CPU icon before the CPU ranges */
    showIcon?: boolean
    /* disallow line breaks after ranges (after the commas) */
    noWrap?: boolean
    /** optional tooltip override */
    tooltip?: string
    /** optional CSS class name(s). */
    className?: string
}

/**
 * The `CPUList` component renders a list of CPU ranges.
 */
export const CPUList = ({ cpus, showIcon, noWrap, tooltip, className }: CPUListProps) => {
    tooltip = tooltip || 'CPU list'
    return !!cpus && (
        <Tooltip title={tooltip}>
            <CPURangeList className={clsx('cpulist', noWrap && 'nowrap', className)}>
                {!!showIcon && <CPUIcon fontSize="small" />}
                {
                    cpus.map((cpurange, index) =>
                        <Fragment key={index}>
                            {index > 0 && ',\u200b'}
                            <span className="cpurange">
                                {cpurange[0] === cpurange[1] ? cpurange[0] : `${cpurange[0]}–${cpurange[1]}`}
                            </span>
                        </Fragment>)
                }
            </CPURangeList>
        </Tooltip>
    )
}

export default CPUList
