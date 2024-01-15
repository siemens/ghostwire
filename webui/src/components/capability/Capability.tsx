// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { dockerdefaultcaps } from 'utils/capabilities'
import { styled } from '@mui/material';


const DefaultCap = styled('span')(() => ({
}))

const NonDefaultCap = styled(DefaultCap)(({ theme }) => ({
    color: theme.palette.containee.elevated.main,
    fontWeight: 'bold',
}))


export interface CapabilityProps {
    /** name of capability (CAP_...). */
    name: string
    /** 
     * please ignore default/non-default and render in a default (plain) style.
     */
    plain?: boolean
}

/**
 * Renders a Linux (kernel) capability that is not in the set of Docker's
 * "standard" capabilities in a different ("high viz") style, so it becomes more
 * easily spottable.
 *
 * See also: [Docker Runtime privilege and Linux
 * capabilities](https://docs.docker.com/engine/reference/run/#runtime-privilege-and-linux-capabilities)
 */
export const Capability = ({ name, plain }: CapabilityProps) => {
    return (plain || !!dockerdefaultcaps[name])
        ? <DefaultCap>{name}</DefaultCap>
        : <NonDefaultCap>{name}</NonDefaultCap>
}

export default Capability
