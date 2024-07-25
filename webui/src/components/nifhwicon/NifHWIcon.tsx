// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { SvgIconProps } from '@mui/material'

import HardwareNicIcon from 'icons/nifs/HardwareNic'
import HardwareNicPFIcon from 'icons/nifs/HardwareNicPF'
import HardwareNicVFIcon from 'icons/nifs/HardwareNicVF'
import { NetworkInterface, SRIOVRole } from 'models/gw/nif'

const nifSRIOVIcons = {
    [SRIOVRole.None]: HardwareNicIcon,
    [SRIOVRole.PF]: HardwareNicPFIcon,
    [SRIOVRole.VF]: HardwareNicVFIcon,
}

export interface NifHWIconProps extends SvgIconProps {
    /** network interface object describing a network interface in detail. */
    nif: NetworkInterface
}

export const NifHWIcon = ({ nif, ...props }: NifHWIconProps) => {
    const HWIcon = nifSRIOVIcons[nif.sriovrole || SRIOVRole.None]
    return <HWIcon {...props} />
}

export default NifHWIcon
