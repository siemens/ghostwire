// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Box } from '@mui/material'

import { Brand } from 'components/brand'
import { BrandIcon } from 'components/brandicon'
import { GwMarkdown } from 'components/gwmarkdown'
import { useDiscovery } from "components/discovery"


// As the "about" markdown text is very short, we don't import it lazily, but
// directly and thus synchronously (for a suitable definition of "synchronous").
import AboutMDX from './About.mdx'
import { useDynVars } from 'components/dynvars'

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

/*
 * Render version and help information about this web application.
 */
export const About = () => (
    // As we want automatic scrollbars, we need to wrap the rendered markdown
    // into its own flexible box.
    <Box m={2} flex={1} overflow="auto" >
        <GwMarkdown mdx={AboutMDX} shortcodes={{ Brand, BrandIcon, DiscoveryMetadata, CaptureMode }} />
    </Box >
)

export default About
