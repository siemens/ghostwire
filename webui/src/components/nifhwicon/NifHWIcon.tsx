// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT


import { type SvgIconProps } from '@mui/material'

import HardwareNicIcon from 'icons/nifs/HardwareNic'
import HardwareNicPFIcon from 'icons/nifs/HardwareNicPF'
import HardwareNicVFIcon from 'icons/nifs/HardwareNicVF'
import { type NetworkInterface, SRIOVRole } from 'models/gw/nif'

const nifSRIOVIcons = {
    [SRIOVRole.None]: HardwareNicIcon,
    [SRIOVRole.PF]: HardwareNicPFIcon,
    [SRIOVRole.VF]: HardwareNicVFIcon,
}

/**
 * Properties of the `NifHWIcon` component; additionally accepts all `SvgIcon`
 * properties, which are passed on to the rendered icon.
 */
export interface NifHWIconProps extends SvgIconProps {
    /** network interface object describing a network interface in detail. */
    nif: NetworkInterface
}

/**
 * Renders the hardware icon of a (physical) network interface, depending on
 * its SR-IOV role: a plain hardware NIC, an SR-IOV physical function (PF), or
 * an SR-IOV virtual function (VF).
 */
export const NifHWIcon = ({ nif, ...props }: NifHWIconProps) => {
    const HWIcon = nifSRIOVIcons[nif.sriovrole || SRIOVRole.None]
    return <HWIcon {...props} />
}

export default NifHWIcon
