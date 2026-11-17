// (c) Siemens AG 2023, 2026
//
// SPDX-License-Identifier: MIT

import React, { useCallback, useEffect, useMemo, useRef } from 'react'
import { BrowserRouter as Router, Route, Routes, useMatch, Navigate, useLocation } from 'react-router-dom'

import { ErrorBoundary } from 'react-error-boundary'

import { basename } from '../utils/basename'

import { SnackbarProvider } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import {
    Badge,
    Box,
    CssBaseline,
    Divider,
    IconButton,
    List,
    ListSubheader,
    ThemeProvider,
    StyledEngineProvider,
    Tooltip,
    Typography,
    useMediaQuery,
    Grid,
    ListItem,
} from '@mui/material';

import { Discovery, useDiscovery } from '../components/discovery'
import Refresher from '../components/refresher'
import AppBarDrawer, { DrawerLinkItem } from '../components/appbardrawer'

import SettingsIcon from '@mui/icons-material/Settings'
import HelpIcon from '@mui/icons-material/Help'

import ScreenshotIcon from '../icons/Screenshot'
import WiringViewIcon from '../icons/views/Wiring'
import NetnsViewIcon from '../icons/views/Details'
import OpenHouseIcon from '../icons/views/OpenHouse'

import { About as AboutView } from '../views/about'
import { Help as HelpView } from '../views/help'
import { Settings as SettingsView, showEmptyNetnsAtom, themeAtom, THEME_DARK, THEME_USERPREF, filterPatternAtom, filterCaseSensitiveAtom, filterRegexpAtom } from '../views/settings'
import { Everything as EverythingView } from '../views/everything'
import { NetnsWiring } from '../views/netnswiring'
import { NetnsDetails as NetnsDetailsView } from '../views/netnsdetails'
import { ContaineeNavigator } from '../components/containeenavigator'
import { useScrollToHash } from '../hooks/scrolltohash'
import { scrollIdIntoView } from '../utils'
import { emptyNonContainerNetns, type NetworkNamespace } from '../models/gw'
import { Brand } from '../components/brand'
import { BrandIcon } from '../components/brandicon'

import { BrandingProvider, useBranding, useBrandName } from '../components/branding'
import { ScreenShooter, useScreenShooterModal } from '../components/screenshooter'
import OpenHouse from '../views/openhouse/OpenHouse'
import { FilterInput, type FilterPattern } from '../components/filterinput'
import { createGhostwireTheme } from './themes'
import { resolveBranding, type GhostwireAppConfig } from './config'
import { AppErrorFallback, ErrorNotice } from '../components/errorfallback'
import { UnhandledErrors } from '../components/unhandlederrors'


const SettingsViewIcon = SettingsIcon
const AboutViewIcon = BrandIcon

const refresherIntervals = [
    { interval: null },
    { interval: 1000 },
    { interval: 5 * 1000 },
    { interval: 30 * 1000 },
    { interval: 60 * 1000 },
    { interval: 5 * 60 * 1000 },
]

interface GwDrawerProps {
    closeDrawer: () => void
    focusRef: React.RefObject<HTMLDivElement | null>
}

// Must stay at module level: defining it inside another component would remount
// it on every parent render, losing the filter input focus.
const GwDrawer = ({ closeDrawer, focusRef }: GwDrawerProps) => {

    const [showEmptyNetns] = useAtom(showEmptyNetnsAtom)
    const [filterPattern, setFilterPattern] = useAtom(filterPatternAtom)
    const [filterCase, setFilterCase] = useAtom(filterCaseSensitiveAtom)
    const [filterRegexp, setFilterRegexp] = useAtom(filterRegexpAtom)

    const discovery = useDiscovery()

    const wmatch1 = useMatch('/w')
    const wmatch2 = useMatch('/w/:slug')
    const nmatch1 = useMatch('/n')
    const nmatch2 = useMatch('/n/:slug')
    const listContainees = wmatch1 !== null || wmatch2 !== null
        || nmatch1 !== null || nmatch2 !== null
    const canFilter = wmatch1 !== null || nmatch1 !== null

    const onFilterChangeHandler = useCallback((fp: FilterPattern) => {
        setFilterPattern(fp.pattern)
        setFilterCase(fp.isCaseSensitive)
        setFilterRegexp(fp.isRegexp)
    }, [setFilterPattern, setFilterCase, setFilterRegexp])

    return <>
        <List onClick={closeDrawer}>
            <DrawerLinkItem
                key="wiring"
                icon={<WiringViewIcon />}
                label="Wiring"
                path="/w"
            />
            <DrawerLinkItem
                key="openhouse"
                icon={<OpenHouseIcon />}
                label="Open & Forwarding Host Ports"
                path="/lochla"
            />
            <DrawerLinkItem
                key="details"
                icon={<NetnsViewIcon />}
                label="Network Namespace Details"
                path="/n"
            />
            <DrawerLinkItem
                key="settings"
                icon={<SettingsViewIcon />}
                label="Settings"
                path="/settings"
            />
            <DrawerLinkItem
                key="help"
                icon={<HelpIcon />}
                label="Help"
                path="/help/gw"
            />
            <DrawerLinkItem
                key="about"
                icon={<AboutViewIcon />}
                label="About"
                path="/about"
            />
        </List>

        {/* show containee navigation only in wiring and detail views */}
        {listContainees && <>
            <Divider />
            <List
                subheader={<ListSubheader onClick={(event) => {
                    event.stopPropagation()
                    event.preventDefault()
                }}>
                    <Grid container sx={{ flexDirection: "column" }}>
                        <Grid>Containees</Grid>
                        {canFilter &&
                            <Grid>
                                <FilterInput
                                    focusRef={focusRef}
                                    filterPattern={{
                                        pattern: filterPattern,
                                        isCaseSensitive: filterCase,
                                        isRegexp: filterRegexp,
                                    }}
                                    onChange={onFilterChangeHandler}
                                    onEnter={closeDrawer}
                                />
                            </Grid>
                        }
                    </Grid>
                </ListSubheader>}
                onClick={closeDrawer}
            >
                <ErrorBoundary
                    resetKeys={[discovery.networkNamespaces]}
                    fallbackRender={({ error }) =>
                        <ListItem>
                            <ErrorNotice sx={{ flex: 1 }} title="containee list unavailable" error={error} />
                        </ListItem>}
                >
                    <ContaineeNavigator
                        allnetns={discovery.networkNamespaces}
                        filterEmpty={!showEmptyNetns}
                    />
                </ErrorBoundary>
            </List>
        </>}
    </>
}

/**
 * The `GhostwireApp` component renders the general app layout without thinking
 * about providers for routing, themes, discovery, et cetera. So this component
 * deals with:
 * - app bar with title, number of namespaces badge, quick actions.
 * - drawer for navigating the different views and types of namespaces.
 * - scrollable content area.
 */
const GhostwireApp = () => {

    const [showEmptyNetns] = useAtom(showEmptyNetnsAtom)

    // Determine the number of discovered network namespaces, as well as the
    // number of shown network namespaces: depending on filter settings and
    // system state, these numbers will differ when "lo"-only network namespaces
    // get filtered out (that is, hidden).
    const discovery = useDiscovery()
    const allnetns = Object.values(discovery.networkNamespaces)
    const IpStackCount = allnetns.length
    const visibleIpStackCount = allnetns
        .filter(netns => !emptyNonContainerNetns(netns as NetworkNamespace) || showEmptyNetns)
        .length
    const badgeContent = visibleIpStackCount.toString()
        + (visibleIpStackCount !== IpStackCount ? '+' : '')

    // Show containees only for specific views/routes: wiring and details,
    // either all or for a single namespace only. This allows navigating between
    // individual detail views.
    const wmatch1 = useMatch('/w')
    const wmatch2 = useMatch('/w/:slug')
    const isWiring = wmatch1 !== null || wmatch2 !== null

    const nmatch1 = useMatch('/n')
    const nmatch2 = useMatch('/n/:slug')
    const isDetails = nmatch1 !== null || nmatch2 !== null

    const listContainees = isWiring || isDetails

    const alsoSnapshotable = useMatch('/lochla') !== null

    // When to show the integrated snapshot functionality?
    const enableSnapshot = listContainees || alsoSnapshotable

    // What to capture, if any.
    const snapshotRef = useRef<HTMLDivElement | null>(null)

    const setModal = useScreenShooterModal()

    useScrollToHash(scrollIdIntoView)

    const { pathname } = useLocation()

    return (
        <Box
            sx={{
                width: '100vw',
                height: '100vh',
                display: 'flex',
                flexDirection: 'column',
            }}
        >
            <AppBarDrawer
                drawerwidth={360}

                title={<>
                    <Typography variant="h6"><Brand /></Typography>&nbsp;
                    <Badge badgeContent={badgeContent} invisible={!IpStackCount} color="secondary">
                        <AboutViewIcon />
                    </Badge>
                </>}

                tools={() => <>
                    {enableSnapshot && allnetns.length > 0 && <Tooltip title="download PNG screenshot">
                        <IconButton
                            color="inherit"
                            onClick={() => setModal && setModal(snapshotRef.current)}
                            size="large">
                            <ScreenshotIcon color="inherit" />
                        </IconButton>
                    </Tooltip>}
                    <Refresher intervals={refresherIntervals} />
                </>}

                drawertitle={() => <>
                    <Typography variant="h6" style={{ flexGrow: 1 }} color="textSecondary" component="span">
                        <Brand />
                    </Typography>
                </>}
                drawer={GwDrawer}
            />

            {/* main content area */}
            <Box component="main" sx={{ m: 0, flex: 1, overflow: "auto", display: "flex", flexDirection: "column" }}>
                {/* a crashing view must not take down the app bar and drawer;
                    retry automatically after navigating or new discovery data. */}
                <ErrorBoundary
                    resetKeys={[pathname, discovery]}
                    fallbackRender={({ error, resetErrorBoundary }) =>
                        <ErrorNotice
                            sx={{ m: 1 }}
                            title="This view couldn't be shown"
                            error={error}
                            onRetry={resetErrorBoundary}
                        />}
                >
                    <Routes>
                        <Route path="/w/:slug" element={<NetnsDetailsView ref={snapshotRef} />} />
                        <Route path="/w" element={<NetnsWiring ref={snapshotRef} />} />
                        <Route path="/n/:slug" element={<NetnsDetailsView ref={snapshotRef} />} />
                        <Route path="/n" element={<EverythingView ref={snapshotRef} />} />
                        <Route path="/lochla" element={<OpenHouse ref={snapshotRef} />} />
                        <Route path="/settings" element={<SettingsView />} />
                        <Route path="/about" element={<AboutView />} />
                        <Route path="/help/*" element={<HelpView />} />
                        <Route path="/" element={<Navigate replace to="/w" />} />
                    </Routes>
                </ErrorBoundary>
            </Box>
        </Box>
    );
}

// We need to wrap the application as otherwise we won't get a confirmer ...
// ouch. And since we're already at wrapping things, let's just wrap up all the
// other wrapping here, such as light/dark theme switching... *snicker*.
const ThemedApp = ({ themes }: Pick<GhostwireAppConfig, 'themes'>) => {

    const brand = useBrandName()
    const snapshotbasename = brand.toLowerCase().replace(/\W/g, '')

    useEffect(() => { document.title = brand }, [brand])

    const { favicon } = useBranding()
    useEffect(() => {
        if (!favicon) {
            return
        }
        let link = document.querySelector<HTMLLinkElement>('link[rel~="icon"]')
        if (!link) {
            link = document.createElement('link')
            link.rel = 'icon'
            document.head.appendChild(link)
        }
        link.removeAttribute('type')
        link.href = favicon
    }, [favicon])

    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
    const [theme] = useAtom(themeAtom)
    const themeMode = theme === THEME_USERPREF
        ? (prefersDarkMode ? 'dark' : 'light')
        : (theme === THEME_DARK ? 'dark' : 'light')

    const appTheme = useMemo(
        () => createGhostwireTheme(themeMode, themes?.[themeMode]),
        [themeMode, themes])

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={appTheme}>
                <CssBaseline />
                <SnackbarProvider maxSnack={3}>
                    <UnhandledErrors />
                    <ErrorBoundary FallbackComponent={AppErrorFallback}>
                        <Discovery />
                        <ScreenShooter basename={snapshotbasename}>
                            <GhostwireApp />
                        </ScreenShooter>
                    </ErrorBoundary>
                </SnackbarProvider>
            </ThemeProvider>
        </StyledEngineProvider>
    )
}

export interface AppProps {
    /** optional app configuration; defaults to Ghostwire's themes and branding. */
    config?: GhostwireAppConfig
}

// Finally, the exported App component wraps the themed app component into a
// Jotai state provider, to keep state provision and app theme switching
// separated. And we also place the router high up here, as well as the
// branding.
const App = ({ config = {} }: AppProps) => {
    const branding = useMemo(() => resolveBranding(config), [config])
    return (
        <BrandingProvider value={branding}>
            <StateProvider>
                <Router basename={basename}>
                    <ThemedApp themes={config.themes} />
                </Router>
            </StateProvider>
        </BrandingProvider>
    )
}

export default App
