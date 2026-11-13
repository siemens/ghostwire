// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { SvgIcon, type SvgIconProps } from '@mui/material'
import { type Container, GHOSTWIRE_LABEL_ROOT } from 'models/gw'

/**
 * Returns the IE app icon for the container if it is an IE app container as
 * well as an associated icon.
 *
 * @param container container object.
 */
export const IEAppProjectIcon = (containee: Container) => {
    const iconData = containee.labels[GHOSTWIRE_LABEL_ROOT+'icon']
    return iconData
        ? (props: SvgIconProps) => <SvgIcon {...props}><image href={iconData} width="24px" height="24px" /></SvgIcon>
        : null
}
