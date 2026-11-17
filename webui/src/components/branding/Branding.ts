// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'
import type { SvgIconProps } from '@mui/material'
import type { MDXContent } from 'mdx/types'

import GhostwireIcon from 'icons/Ghostwire'
import { useDynVars } from 'components/dynvars'

/**
 * A link to legal information, such as an imprint, privacy statement, or open
 * source software notices.
 */
export interface BrandingLink {
    /** link text. */
    label: string
    /** link URL. */
    href: string
}

/**
 * Branding of the app, as configured at build time. The brand name and icon
 * passed in by the Ghostwire service via the dynamic variables take precedence
 * over the build-time branding.
 */
export interface Branding {
    /** product name; shown in the app bar, title, about, and help. */
    name: string
    /** product icon; shown in the app bar, about, and help. */
    icon: React.ComponentType<SvgIconProps>
    /** URL of the favicon to use instead of the favicon from index.html. */
    favicon?: string
    /**
     * MDX content to render in the about view instead of the default about
     * content. It can use the `<Brand/>`, `<BrandIcon/>`, `<AppVersion/>`,
     * `<DiscoveryMetadata/>`, and `<CaptureMode/>` shortcodes.
     */
    about?: MDXContent
    /** legal links to show at the end of the about view. */
    links: BrandingLink[]
    /** web app version. */
    version: string
}

/** The default Ghostwire branding. */
export const defaultBranding: Branding = {
    name: 'Ghostwire',
    icon: GhostwireIcon,
    links: [],
    version: import.meta.env.VITE_REACT_APP_GIT_VERSION,
}

const BrandingContext = React.createContext<Branding>(defaultBranding)
BrandingContext.displayName = 'Branding'

/** Provides the branding to a (sub)tree of components. */
export const BrandingProvider = BrandingContext.Provider

/** Returns the branding of the app. */
export const useBranding = () => useContext(BrandingContext)

/**
 * Returns the brand name: either the brand name passed in by the Ghostwire
 * service or otherwise the build-time brand name.
 */
export const useBrandName = () => {
    const { brand } = useDynVars()
    const { name } = useBranding()
    return brand ? String(brand) : name
}
