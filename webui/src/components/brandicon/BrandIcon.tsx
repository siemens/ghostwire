// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import GhostwireIcon from 'icons/Ghostwire'
import { useDynVars } from 'components/dynvars'
import { SvgIconProps } from '@mui/material'
import SVG, { Props as SVGProps } from 'react-inlinesvg'

/**
 * Renders the brand icon as an icon: this is either the default "Ghostwire"
 * brand icon or an optional brand name override via the dynamic variables
 * passed to us by the Ghostwire service.
 */
export const BrandIcon = (props: SvgIconProps) => {

    const { brandicon } = useDynVars()

    if (!!brandicon) {
        return <SVG
            className="MuiSvgIcon-root"
            uniquifyIDs={true}
            src={`data:image/svg+xml;utf-8,${encodeURIComponent(brandicon)}`}
            {...{
                fill: 'currentColor', // ensure in MUIv5 to use color for filling, unless explicitly overridden.
                ...(props as unknown as SVGProps),
            }}
        />
    }
    return <GhostwireIcon {...props}/>
}
