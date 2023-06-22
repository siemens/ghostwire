// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { NetworkInterface } from 'models/gw'
import { styled } from '@mui/material'


const Key = styled('span')(({ theme }) => ({
    color: theme.palette.text.secondary,
}))


export interface VxlanDetailsProps {
    /** (VXLAN) network interface object. */
    nif: NetworkInterface
    /** CSS class name(s) */
    className?: string
}

/**
 * Renders additional properties of a VXLAN network interface, such as the
 * assigned VXLAN ID.
 */
export const VxlanDetails = ({ nif, className }: VxlanDetailsProps) => {
    const vxlan = nif.vxlanDetails

    if (!vxlan) {
        return null
    }
    return (<div className={className || ''}>
        <div><Key>VXLAN ID:</Key> {vxlan.vid}</div>
        <div><Key>VXLAN arp proxy:</Key> {vxlan.arpProxy ? 'enabled' : 'disabled'}</div>
    </div>)
}
