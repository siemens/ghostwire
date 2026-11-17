// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Theme, ThemeOptions } from '@mui/material'

import { defaultBranding, type Branding } from 'components/branding'

/**
 * Configuration of the Ghostwire app, allowing downstream builds to change
 * the MUI themes and the branding. Any configuration elements left out default
 * to Ghostwire's defaults.
 */
export interface GhostwireAppConfig {
    /**
     * MUI theme options (or themes) for light and dark mode; these get deeply
     * merged into the default Ghostwire theme options.
     */
    themes?: {
        light?: ThemeOptions | Theme
        dark?: ThemeOptions | Theme
    }
    /** branding of the app. */
    branding?: Partial<Branding>
}

/**
 * Returns the branding of the configuration, with any missing branding
 * elements set to their defaults.
 */
export const resolveBranding = (config: GhostwireAppConfig): Branding => ({
    ...defaultBranding,
    ...Object.fromEntries(
        Object.entries(config.branding ?? {}).filter(([, value]) => value !== undefined)),
} as Branding)
