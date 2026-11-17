// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { createSvgIcon } from '@mui/material'

import { createGhostwireApp } from 'ghostwire-ui'

import '@fontsource/roboto/400.css'
import '@fontsource/roboto/500.css'
import '@fontsource/roboto/700.css'
import '@fontsource/roboto-mono/400.css'

const ExampleIcon = createSvgIcon(<path d="M12 2 2 22h20z" />, 'ExampleIcon')

const App = createGhostwireApp({
    themes: {
        light: { palette: { primary: { main: '#2e7d32' } } },
        dark: { palette: { primary: { main: '#81c784' } } },
    },
    branding: {
        name: 'Example Netview',
        icon: ExampleIcon,
        links: [
            { label: 'Open Source Software Notices', href: 'https://example.org/oss' },
        ],
    },
})

createRoot(document.getElementById('root')!).render(
    <StrictMode><App /></StrictMode>
)
