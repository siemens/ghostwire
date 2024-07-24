// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { styled, Tooltip } from '@mui/material'
import CPUIcon from 'icons/CPUAffinity'


const CPURangeList = styled('span')(() => ({
    '&.cpulist > .MuiSvgIcon-root': {
        verticalAlign: 'text-top',
        position: 'relative',
        top: '0.1ex',
        marginRight: '0.2em',
    },
}))

export interface CPUListProps {
    /* list of CPU ranges */
    cpus: number[][] | null
    /* show/hide a CPU icon before the CPU ranges */
    showIcon?: boolean
    /* allow line breaks after range (after the comma) */
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
    const sep = noWrap ? ',' : ',\u200b'
    tooltip = tooltip || 'CPU list'
    return !!cpus && (
        <Tooltip title={tooltip}>
            <CPURangeList className={clsx('cpulist', className)}>
                {!!showIcon && <CPUIcon fontSize="inherit" />}
                {
                    cpus.map((cpurange, index) => {
                        if (cpurange[0] === cpurange[1]) {
                            return <span key={index}>{index > 0 && sep}{cpurange[0]}</span>
                        }
                        return <span key={index}>{index > 0 && sep}{cpurange[0]}–{cpurange[1]}</span>
                    })
                }
            </CPURangeList>
        </Tooltip>
    )
}

export default CPUList
