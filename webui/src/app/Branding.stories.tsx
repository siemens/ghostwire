// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box, Button, createSvgIcon, List, ThemeProvider, Typography, type Theme } from '@mui/material'

import AppBarDrawer, { DrawerLinkItem } from '../components/appbardrawer'
import { Brand } from '../components/brand'
import { BrandIcon } from '../components/brandicon'
import { BrandingProvider } from '../components/branding'
import { About } from '../views/about'

import { resolveBranding, type GhostwireAppConfig } from './config'
import { createGhostwireTheme } from './themes'

const ExampleIcon = createSvgIcon(<path d="M12 2 2 22h20z" />, 'ExampleIcon')

const exampleConfig: Required<GhostwireAppConfig> = {
    themes: {
        light: { palette: { primary: { main: '#6a1b9a' }, background: { default: '#f3e5f5' } } },
        dark: { palette: { primary: { main: '#ce93d8' } } },
    },
    branding: {
        name: 'Example Netview',
        icon: ExampleIcon,
        links: [
            { label: 'Imprint', href: 'https://example.org/imprint' },
            { label: 'Open Source Software Notices', href: 'https://example.org/oss' },
        ],
        version: '9.8.7-example',
    },
}

/**
 * Renders the app bar and about view of the Ghostwire app, themed using
 * `createGhostwireTheme` with the specified app configuration's theme options
 * for the current light or dark mode.
 */
const Showcase = ({ config }: { config: GhostwireAppConfig }) => {
    const branding = resolveBranding(config)
    return (
        <BrandingProvider value={branding}>
            <ThemeProvider theme={(outer: Theme) =>
                createGhostwireTheme(outer.palette.mode, config.themes?.[outer.palette.mode])}
            >
                <Box sx={{ bgcolor: 'background.default' }}>
                    <AppBarDrawer
                        title={<><BrandIcon />&nbsp;<Typography variant="h6"><Brand /></Typography></>}
                        drawertitle={<Typography variant="h6" color="textSecondary"><Brand /></Typography>}
                        drawer={({ closeDrawer }) => (
                            <List onClick={closeDrawer}>
                                <DrawerLinkItem label="About" icon={<BrandIcon />} path="/" />
                            </List>
                        )}
                    />
                    <Box sx={{ m: 2 }}>
                        <Button variant="contained">primary</Button>&nbsp;
                        <Button variant="contained" color="secondary">secondary</Button>
                    </Box>
                    <About />
                </Box>
            </ThemeProvider>
        </BrandingProvider>
    )
}

const meta: Meta<typeof Showcase> = {
    title: 'Ghostwire/App Theming and Branding',
    component: Showcase,
    argTypes: {
        config: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Showcase>

export const Default: Story = {
    args: {
        config: {},
    },
}

/**
 * Non-default MUI theme options for light and dark mode, as well as a
 * different brand name, icon, version, and legal links.
 */
export const Rebranded: Story = {
    args: {
        config: exampleConfig,
    },
}
