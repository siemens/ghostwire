// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useRef } from 'react'
import { BrowserRouter as Router, Route, Routes, useMatch, Navigate } from 'react-router-dom'

import { basename } from 'utils/basename'

import { SnackbarProvider } from 'notistack'

import { Provider as StateProvider, useAtom } from 'jotai'

import {
    Badge,
    Box,
    createTheme,
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
} from '@mui/material';

import { gwDarkTheme, gwLightTheme } from './appstyles'

import { Discovery, useDiscovery } from 'components/discovery'
import Refresher from 'components/refresher'
import AppBarDrawer, { DrawerLinkItem } from 'components/appbardrawer'

import SettingsIcon from '@mui/icons-material/Settings'
import HelpIcon from '@mui/icons-material/Help'

import ScreenshotIcon from 'icons/Screenshot'
import WiringViewIcon from 'icons/views/Wiring'
import NetnsViewIcon from 'icons/views/Details'
import OpenHouseIcon from 'icons/views/OpenHouse'

import { About as AboutView } from 'views/about'
import { Help as HelpView } from 'views/help'
import { Settings as SettingsView, showEmptyNetnsAtom, themeAtom, THEME_DARK, THEME_USERPREF, filterPatternAtom, filterCaseSensitiveAtom, filterRegexpAtom } from 'views/settings'
import { Everything as EverythingView } from 'views/everything'
import { NetnsWiring } from 'views/netnswiring'
import { NetnsDetails as NetnsDetailsView } from 'views/netnsdetails'
import { ContaineeNavigator } from 'components/containeenavigator'
import { useScrollToHash } from 'hooks/scrolltohash'
import { scrollIdIntoView } from 'utils'
import { emptyNetns, NetworkNamespace } from 'models/gw'
import { Brand } from 'components/brand'
import { BrandIcon } from 'components/brandicon'

import { useDynVars } from 'components/dynvars'
import { ScreenShooter, useScreenShooterModal } from 'components/screenshooter'
import OpenHouse from 'views/openhouse/OpenHouse'
import { FilterInput, FilterPattern } from 'components/filterinput'


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

/**
 * The `LxGhostwireApp` component renders the general app layout without
 * thinking about providers for routing, themes, discovery, et cetera. So this
 * component deals with:
 * - app bar with title, number of namespaces badge, quick actions.
 * - drawer for navigating the different views and types of namespaces.
 * - scrollable content area.
 */
const GhostwireApp = () => {

    const [showEmptyNetns] = useAtom(showEmptyNetnsAtom)
    const [filterPattern, setFilterPattern] = useAtom(filterPatternAtom)
    const [filterCase, setFilterCase] = useAtom(filterCaseSensitiveAtom)
    const [filterRegexp, setFilterRegexp] = useAtom(filterRegexpAtom)

    // Determine the number of discovered network namespaces, as well as the
    // number of shown network namespaces: depending on filter settings and
    // system state, these numbers will differ when "lo"-only network namespaces
    // get filtered out (that is, hidden).
    const discovery = useDiscovery()
    const allnetns = Object.values(discovery.networkNamespaces)
    const IpStackCount = allnetns.length
    const visibleIpStackCount = allnetns
        .filter(netns => !emptyNetns(netns as NetworkNamespace) || showEmptyNetns)
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
    const isDetails = nmatch1 !== null || nmatch2 != null

    const listContainees = isWiring || isDetails

    const alsoSnapshotable = useMatch('/lochla') !== null

    // When to show the integrated snapshot functionality?
    const enableSnapshot = listContainees || alsoSnapshotable

    // What to capture, if any.
    const snapshotRef = useRef<HTMLDivElement | null>(null)

    const setModal = useScreenShooterModal()

    const onFilterChangeHandler = (fp: FilterPattern) => {
        if (filterPattern != fp.pattern) setFilterPattern(fp.pattern)
        if (filterCase != fp.isCaseSensitive) setFilterCase(fp.isCaseSensitive)
        if (filterRegexp != fp.isRegexp) setFilterRegexp(fp.isRegexp)
    }

    useScrollToHash(scrollIdIntoView)

    return (
        <Box width="100vw" height="100vh" display="flex" flexDirection="column">
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
                drawer={closeDrawer => <>
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
                                <Grid container direction="column">
                                    <Grid item>Containees</Grid>
                                    <Grid item>
                                        <FilterInput
                                            filterPattern={{
                                                pattern: filterPattern,
                                                isCaseSensitive: filterCase,
                                                isRegexp: filterRegexp,
                                            }}
                                            onChange={onFilterChangeHandler}
                                        />
                                    </Grid>
                                </Grid>
                            </ListSubheader>}
                            onClick={closeDrawer}
                        >
                            <ContaineeNavigator
                                allnetns={discovery.networkNamespaces}
                                filterEmpty={!showEmptyNetns}
                            />
                        </List>
                    </>}
                </>}
            />

            {/* main content area */}
            <Box m={0} flex={1} overflow="auto">
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
            </Box>
        </Box>
    );
}

// We need to wrap the application as otherwise we won't get a confirmer ...
// ouch. And since we're already at wrapping things, let's just wrap up all the
// other wrapping here, such as light/dark theme switching... *snicker*.
const ThemedApp = () => {

    const { brand } = useDynVars()
    const snapshotbasename = (brand || 'ghostwire').toString()
        .toLowerCase().replace(/\W/g, '')

    const prefersDarkMode = useMediaQuery('(prefers-color-scheme: dark)')
    const [theme] = useAtom(themeAtom)
    const themeMode = theme === THEME_USERPREF
        ? (prefersDarkMode ? 'dark' : 'light')
        : (theme === THEME_DARK ? 'dark' : 'light')

    const appTheme = React.useMemo(() => createTheme({
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
            mode: themeMode,
            primary: { main: '#009999' },
            secondary: { main: '#ffc400' },
        },
    }, themeMode === 'dark' ? gwDarkTheme : gwLightTheme), [themeMode])

    appTheme.palette.containee.privileged.contrastText =
        appTheme.palette.getContrastText(appTheme.palette.containee.privileged.main)

    return (
        <StyledEngineProvider injectFirst>
            <ThemeProvider theme={appTheme}>
                <CssBaseline />
                <SnackbarProvider maxSnack={3}>
                    <Discovery />
                    <Router basename={basename}>
                        <ScreenShooter basename={snapshotbasename}>
                            <GhostwireApp />
                        </ScreenShooter>
                    </Router>
                </SnackbarProvider>
            </ThemeProvider>
        </StyledEngineProvider>
    );
}

const App = () => (
    <StateProvider>
        <ThemedApp />
    </StateProvider>
)

export default App
