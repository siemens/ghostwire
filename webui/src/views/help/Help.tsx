// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { ReactNode } from 'react'

import { Provider } from 'jotai'

import { Box, Card, useTheme } from '@mui/material'

import { Brand } from 'components/brand'
import { BrandIcon } from 'components/brandicon'
import { HelpViewer, HelpViewerChapter } from 'components/helpviewer'
import { GwMarkdown } from 'components/gwmarkdown'
import { MuiMarkdownProps } from 'components/muimarkdown'

import { containeesCutoffAtom, nifsCutoffAtom, portsCutoffAtom, routesCutoffAtom, showEmptyNetnsAtom, showIpv4Atom, showIpv6Atom, showLoopbackAtom, showMACAtom, showNamespaceIdsAtom, showSandboxesAtom } from 'views/settings'
import { useHydrateAtoms } from 'jotai/utils'


/**
 * Convenience wrapper for lazily importing a help chapter MDX module.
 * 
 * @param name name (without .mdx extension and without any path) of a chapter
 * .mdx file; chapter files are located in the chapters/ subdirectory.
 */
const ch = (name: string) => React.lazy(() => import(`./chapters/${name}.mdx`))

// DynamicVars represents an object "map" of dynamic variable passed into an
// application by the server at load time (as opposed to static REACT_APP_
// variables which are fixed at build time).
type DynamicVars = { [key: string]: unknown }

// Dynamic variables from the server get passed in via dynamically served
// "index.html" that sets the dynvars element of the window object.
declare global {
    interface Window {
        dynvars: DynamicVars
    }
}

const chapters: HelpViewerChapter[] = [
    { title: (window.dynvars && window.dynvars.brand as string) || 'Ghostwire', chapter: ch('Ghostwire'), slug: 'gw' },
    { title: 'Discovery/Refresh', chapter: ch('Refresh'), slug: 'refresh' },
    { title: 'IP Stacks Galore!', chapter: ch('Badge'), slug: 'badge' },
    { title: 'Technical Features', chapter: ch('Technical'), slug: 'tech' },
    { title: 'Navigation Drawer', chapter: ch('Drawer'), slug: 'drawer' },
    { title: 'Wiring View', chapter: ch('Wiring'), slug: 'wiring' },
    { title: 'Open & Forwarding Host Ports', chapter: ch('Lochla'), slug: 'lochla' },
    { title: 'Details View', chapter: ch('Details'), slug: 'details' },
    { title: 'Network Interfaces', chapter: ch('Nifs'), slug: 'nifs' },
    { title: 'Containees', chapter: ch('Containees'), slug: 'containees' },
    { title: 'Live Capture', chapter: ch('Capture'), slug: 'capture' },
    { title: 'Settings', chapter: ch('Settings'), slug: 'settings' },
]

const markdowner = (props: MuiMarkdownProps) => (<GwMarkdown {...props} />)

interface ExampleProps {
    children: ReactNode
    p: string | number
    card: object
}

/**
 * Shortcode component rendering a Mui card with the specified children inside
 * it. The card has a theme-based margin as well as internal padding.
 */
const Example = ({ children, p, card, ...otherprops }: ExampleProps) => {
    return (
        <Box m={2} {...otherprops}>
            <Card {...card}>
                <Box p={p || 0}>
                    {children}
                </Box>
            </Card>
        </Box>
    )
}

/**
 * Renders a "fake" application bar for use in help examples.
 */
const FakeAppBar = ({ children }: { children: ReactNode }) => {
    const theme = useTheme()
    return (
        <Example p={2} card={{
            square: true,
            style: {
                maxWidth: '15em',
                color: theme.palette.primary.contrastText,
                background: theme.palette.primary.main,
            },
        }}>
            ☰&nbsp;&nbsp;
            {children}
        </Example>
    )
}

const HydrateAtoms = ({ initialValues, children }) => {
    useHydrateAtoms(initialValues)
    return children
}

const shortcodes = { Brand, BrandIcon, Example, FakeAppBar }

/**
 * Render the detailed help (chapters).
 *
 * Note: the help viewer already handles overflow/scrolling itself (it has to).
 * It also does proper themed margins.
 *
 * Note 2: To ensure deterministic rendering with respect to user configurable
 * settings, the help viewer supplies its own jotai provider with well-defined
 * settings.
 */
export const Help = () => {
    return <Provider>
        <HydrateAtoms initialValues={[
            [showLoopbackAtom, false],
            [showEmptyNetnsAtom, false],
            [showMACAtom, true],
            [showIpv4Atom, true],
            [showIpv6Atom, true],
            [containeesCutoffAtom, 100],
            [portsCutoffAtom, 100],
            [routesCutoffAtom, 100],
            [nifsCutoffAtom, 100],
            [showSandboxesAtom, false],
            [showNamespaceIdsAtom, true],
        ]}>
            <HelpViewer
                chapters={chapters}
                baseroute="/help"
                markdowner={markdowner}
                shortcodes={shortcodes}
                style={{ overflow: 'visible' }}
            />
        </HydrateAtoms>
    </Provider>
}

export default Help
