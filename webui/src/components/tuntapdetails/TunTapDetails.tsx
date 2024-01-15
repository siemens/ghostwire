// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Process } from "components/process"
import { NetworkInterface } from "models/gw"

export interface TunTapDetailsProps {
    /** (VXLAN) network interface object. */
    nif: NetworkInterface
    /** CSS class name(s) */
    className?: string
}

/**
 * Renders additional properties of a TAP or TUN network interface, such as the
 * processes handling the traffic.
 */
export const TunTapDetails = ({ nif, className }: TunTapDetailsProps) => {
    const tuntap = nif.tuntapDetails

    if (!tuntap || !tuntap.processors.length) {
        return null
    }
    return (<div className={className || ''}>
        {tuntap.processors
            .sort((proc1, proc2) => proc1.pid - proc2.pid)
            .map((proc) =>
                <div key={`${nif.name}-${proc.pid}`} >
                    <Process
                        cmdline={proc.cmdline}
                        containee={proc.containee}
                        pid={proc.pid} />
                </div>)
        }
    </div>)
}
