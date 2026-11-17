// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import { createTheme, type PaletteMode, type Theme, type ThemeOptions } from '@mui/material'
import { mergeWith } from 'lodash'

import { gwDarkTheme, gwLightTheme } from '../styles/themes'

// Deeply merges the sources into a new object, with arrays getting replaced
// instead of merged element-wise, like MUI's own deepmerge.
const deepmerge = <T extends object>(...sources: object[]): T => mergeWith(
    {}, ...sources,
    (_: unknown, src: unknown) => Array.isArray(src) ? src : undefined)

/**
 * Returns the default Ghostwire theme options for the specified palette mode.
 */
export const ghostwireThemeOptions = (mode: PaletteMode): ThemeOptions => deepmerge<ThemeOptions>(
    {
        components: {
            MuiSelect: {
                defaultProps: {
                    variant: 'standard', // MUI v4 default.
                },
            },
            MuiCssBaseline: {
                styleOverrides: {
                    body: {
                        fontSize: '0.875rem', // ...go back to typography body2 font size as in MUI v4.
                        lineHeight: 1.43,
                        letterSpacing: '0.01071em',
                    },
                },
            },
        },
        palette: {
            mode: mode,
            primary: { main: '#009999' },
            secondary: { main: '#ffc400' },
        },
    },
    mode === 'dark' ? gwDarkTheme : gwLightTheme,
)

/**
 * Creates a MUI theme for the specified palette mode from the default
 * Ghostwire theme options, deep-merged with the optional theme options (or
 * theme) overrides. As the Ghostwire-specific palette elements are always
 * present, overrides only need to specify what they want to change.
 */
export const createGhostwireTheme = (mode: PaletteMode, overrides?: ThemeOptions | Theme): Theme => {
    const theme = createTheme(deepmerge<ThemeOptions>(ghostwireThemeOptions(mode), overrides ?? {}))
    const privileged = theme.palette.containee.privileged
    return createTheme(theme, {
        palette: {
            containee: {
                privileged: {
                    ...privileged,
                    contrastText: theme.palette.getContrastText(privileged.main)
                }
            }
        }
    })
}
