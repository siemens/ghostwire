// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { lazy } from 'react'

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box, createTheme, Paper, ThemeProvider, Typography } from '@mui/material'

import intro from '../helpviewer/01-intro.mdx'
import foobar from '../helpviewer/02-foobar.mdx'

import { MuiMarkdown } from './MuiMarkdown'

const meta: Meta<typeof MuiMarkdown> = {
    title: 'Universal/MuiMarkdown',
    component: MuiMarkdown,
    argTypes: {
        mdx: { control: false },
        shortcodes: { control: false },
        fallback: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof MuiMarkdown>

/**
 * In this story we (synchronously) import markdown "compiled" into
 * [MDX](https://mdxjs.com) and feed that into our `MuiMarkdown` component.
 */
export const Short: Story = {
    args: {
        mdx: intro,
    },
}

export const Long: Story = {
    args: {
        mdx: foobar,
    },
}

const LazyFoobar = lazy(() => import('../helpviewer/02-foobar.mdx'))

/**
 * Lazy loading of potentially larger markdown/MDX documents is also supported
 * and works as follows:
 *
 * ```tsx
 * const LazyExampleMDX = React.lazy(() => import('./example.mdx'))
 * ```
 */
export const Lazy: Story = {
    args: {
        mdx: (props) => <LazyFoobar {...props} />,
    },
}

const themes = [
    createTheme({ palette: { mode: 'light' } }),
    createTheme({ palette: { mode: 'dark' } }),
]

/** This component has basic support for light and dark theming. */
export const LightAndDarkThemes: Story = {
    render: () => <>
        {themes.map((theme, idx) => (
            <ThemeProvider key={idx} theme={theme}>
                <Box sx={{ p: 1, bgcolor: 'background.default', color: 'text.primary' }}>
                    <Typography>with {theme.palette.mode} theme:</Typography>
                    <Paper variant="outlined" sx={{ p: 1 }}>
                        <MuiMarkdown mdx={intro} />
                    </Paper>
                </Box>
            </ThemeProvider>
        ))}
    </>,
}
