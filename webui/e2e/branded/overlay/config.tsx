// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

// An example app configuration overlay with non-default themes and branding,
// built using VITE_APP_CONFIG=e2e/branded/overlay/config.tsx.

import { createSvgIcon } from '@mui/material'

import type { GhostwireAppConfig } from 'app'

import About from './About.mdx'
import favicon from './favicon.svg?url'

const ExampleIcon = createSvgIcon(
    <path data-testid="example-brand-icon" d="M12 2 2 22h20z" />,
    'ExampleIcon')

const config: GhostwireAppConfig = {
    themes: {
        light: {
            palette: {
                primary: { main: '#6a1b9a' },
                background: { default: '#f3e5f5' },
            },
        },
        dark: {
            palette: {
                primary: { main: '#ce93d8' },
            },
        },
    },
    branding: {
        name: 'Example Netview',
        icon: ExampleIcon,
        favicon: favicon,
        about: About,
        links: [
            { label: 'Imprint', href: 'https://example.org/imprint' },
            { label: 'Open Source Software Notices', href: 'https://example.org/oss' },
        ],
        version: '9.8.7-example',
    },
}

export default config
