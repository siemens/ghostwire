// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { useBranding } from '../branding'
import { useDynVars } from '../dynvars'
import type { SvgIconProps } from '@mui/material'
import SVG, { type Props as SVGProps } from 'react-inlinesvg'

/**
 * Renders the brand icon as an icon: this is either the build-time brand icon
 * (defaulting to the "Ghostwire" icon) or an optional brand icon override via
 * the dynamic variables passed to us by the Ghostwire service.
 */
export const BrandIcon = (props: SvgIconProps) => {

    const { brandicon } = useDynVars()
    const { icon: Icon } = useBranding()

    if (typeof brandicon === 'string' && brandicon !== '') {
        return <SVG
            className="MuiSvgIcon-root"
            uniquifyIDs={true}
            {...{
                fill: 'currentColor', // ensure in MUIv5 to use color for filling, unless explicitly overridden.
                ...(props as unknown as SVGProps),
            }}
            src={`data:image/svg+xml;utf-8,${encodeURIComponent(brandicon)}`}
        />
    }
    return <Icon {...props}/>
}
