// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT


import { Box, Link } from '@mui/material'

import { Brand } from '../../components/brand'
import { BrandIcon } from '../../components/brandicon'
import { useBranding } from '../../components/branding'
import { GwMarkdown } from '../../components/gwmarkdown'
import { useDiscovery } from "../../components/discovery"


// As the "about" markdown text is very short, we don't import it lazily, but
// directly and thus synchronously (for a suitable definition of "synchronous").
import AboutMDX from './About.mdx'
import { useDynVars } from '../../components/dynvars'

// MDX "shortcode" component rendering the discovery data creator identifier and
// version, if available. 
const DiscoveryMetadata = () => {
    const discovery = useDiscovery()
    return (<>
        {discovery.metadata
            ? `${discovery.metadata["creator-id"]} ${discovery.metadata['creator-version']}`
            : <>(<em>please refresh discovery data</em>)</>}
    </>)
}

const CaptureMode = () => {
    const { enableMonolith } = useDynVars()
    return <>{enableMonolith ? 'enabled ' : 'service not available'}</>
}

// MDX "shortcode" component rendering the web app version.
const AppVersion = () => <span>{useBranding().version}</span>

/*
 * Render version and help information about this web application, using
 * either the default or the branding's about content, followed by any legal
 * links of the branding.
 */
export const About = () => {
    const { about, links } = useBranding()
    return (
        // As we want automatic scrollbars, we need to wrap the rendered markdown
        // into its own flexible box.
        <Box sx={{ m: 2, flex: 1, overflow: 'auto' }}>
            <GwMarkdown mdx={about ?? AboutMDX} shortcodes={{ Brand, BrandIcon, AppVersion, DiscoveryMetadata, CaptureMode }} />
            {links.length > 0 &&
                <Box component="ul" aria-label="legal information" sx={{ typography: 'body1' }}>
                    {links.map(({ label, href }) =>
                        <li key={href}>
                            <Link href={href} target="_blank" rel="noopener noreferrer">{label}</Link>
                        </li>)}
                </Box>}
        </Box >
    )
}

export default About
