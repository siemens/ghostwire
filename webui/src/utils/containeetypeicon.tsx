// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useMemo } from 'react'
import { type SvgIconProps } from '@mui/material'

import { type Containee } from 'models/gw'
import { ContaineeIcon } from './containeeicon'

export interface ContaineeTypeIconProps extends SvgIconProps {
    containee: Containee
    /** force generic IE runtime/app icons instead of app-specific icons. */
    forceStdIcon?: boolean
}

/**
 * Renders the type icon of the specified containee.
 */
export const ContaineeTypeIcon = ({ containee, forceStdIcon, ...props }: ContaineeTypeIconProps) => {
    const Icon = useMemo(() => ContaineeIcon(containee, forceStdIcon), [containee, forceStdIcon])
    // Icon is a static icon component, or an IE app icon memoized per containee.
    // eslint-disable-next-line react-hooks/static-components
    return <Icon {...props} />
}

export default ContaineeTypeIcon
