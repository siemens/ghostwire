// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { atom, useAtom, useSetAtom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

import {
    Box,
    Card,
    Divider,
    Grid,
    List,
    ListItem,
    ListItemSecondaryAction,
    ListItemText,
    MenuItem,
    Select,
    SelectChangeEvent,
    styled,
    Switch as Toggle,
    Typography,
} from '@mui/material';
import { AddressFamily } from 'models/gw'
import { CutoffSelector, NEVER } from 'components/cutoffselector'


const themeKey = 'ghostwire.theme'

const showLoopbackKey = 'ghostwire.showlo'
const showEmptyNetnsKey = 'ghostwire.showemptynetns'
const showMACKey = 'ghostwire.showmac'
const showIpv4Key = 'ghostwire.ipv4'
const showIpv6Key = 'ghostwire.ipv6'

const showSandboxesKey = 'ghostwire.showsandboxes'
const showNamespaceIds = 'ghostwire.shownsids'

const cutoffContaineesKey = 'ghostwire.cutoff.containees'
const cutoffNeighborhoodKey = 'ghostwire.cutoff.neighborhood'
const cutoffPortsKey = 'ghostwire.cutoff.ports'
const cutoffForwardedPortsKey = 'ghostwire.cutoff.forwardedports'
const cutoffRoutesKey = 'ghostwire.cutoff.routes'
const cutoffNifsKey = 'ghostwire.cutoff.nifs'

const snapshotDensityKey = 'ghostwire.snapshot.density'

const showIEAppIconsKey = 'ghostwire.showieappicons'

const showMultiBroadcastRoutesKey = 'ghostwire.showmultibroadcastroutes'

const filterPatternKey = 'ghostwire.filter.pattern'
const filterCaseSensitiveKey = 'ghostwire.filter.case'
const filterRegexpKey = 'ghostwire.filter.regexp'


export const THEME_USERPREF = 0
export const THEME_LIGHT = 1
export const THEME_DARK = -1
export const themeAtom = atomWithStorage(themeKey, THEME_USERPREF)

export const showLoopbackAtom = atomWithStorage(showLoopbackKey, false)
export const showEmptyNetnsAtom = atomWithStorage(showEmptyNetnsKey, false)
export const showMACAtom = atomWithStorage(showMACKey, false)

// The "IP families" setting actually consists of separate IPv4 and IPv6 boolean
// settings.
export const showIpv4Atom = atomWithStorage(showIpv4Key, true)
export const showIpv6Atom = atomWithStorage(showIpv6Key, false) // this default is PAINFUL!
export const showIpFamiliesAtom = atom((get) => {
    const families = get(showIpv4Atom) ? [AddressFamily.IPv4] : []
    get(showIpv6Atom) && families.push(AddressFamily.IPv6)
    return families
})

export const containeesCutoffAtom = atomWithStorage(cutoffContaineesKey, 3)
export const neighborhoodCutoffAtom = atomWithStorage(cutoffNeighborhoodKey, 10)
export const portsCutoffAtom = atomWithStorage(cutoffPortsKey, 10)
export const forwardedPortsCutoffAtom = atomWithStorage(cutoffForwardedPortsKey, 10)
export const routesCutoffAtom = atomWithStorage(cutoffRoutesKey, 10)
export const nifsCutoffAtom = atomWithStorage(cutoffNifsKey, NEVER)

export const showSandboxesAtom = atomWithStorage(showSandboxesKey, false)

export const showNamespaceIdsAtom = atomWithStorage(showNamespaceIds, false)

export const snapshotDensityAtom = atomWithStorage(snapshotDensityKey, 1)

export const showIEAppIconsAtom = atomWithStorage(showIEAppIconsKey, false)

export const showMultiBroadcastRoutesAtom = atomWithStorage(showMultiBroadcastRoutesKey, false)

export const filterPatternAtom = atomWithStorage(filterPatternKey, '')
export const filterCaseSensitiveAtom = atomWithStorage(filterCaseSensitiveKey, false)
export const filterRegexpAtom = atomWithStorage(filterRegexpKey, false)

const cutOffEm = 12

const SettingsGrid = styled(Grid)(({ theme }) => ({
    width: `calc(100% - calc(${theme.spacing(2)} * 2))`,
    margin: theme.spacing(2),

    '& .MuiCard-root + .MuiTypography-subtitle1': {
        marginTop: theme.spacing(4),
    },
}))


/**
 * Renders the "settings" page (view) of the Ghostwire client browser app.
 */
export const Settings = () => {
    // Tons of settings to play around with...
    const [theme, setTheme] = useAtom(themeAtom)
    const [showLoopbacks, setShowLoopbacks] = useAtom(showLoopbackAtom)
    const [showEmptyNetns, setShowEmptyNetns] = useAtom(showEmptyNetnsAtom)
    const [showMAC, setShowMAC] = useAtom(showMACAtom)
    const [showIpFamilies] = useAtom(showIpFamiliesAtom)
    const setIpv4 = useSetAtom(showIpv4Atom)
    const setIpv6 = useSetAtom(showIpv6Atom)
    const [showSandboxes, setShowSandboxes] = useAtom(showSandboxesAtom)
    const [showNamespaceIds, setShowNamespaceIds] = useAtom(showNamespaceIdsAtom)
    const [showIEAppIcons, setShowIEAppIcons] = useAtom(showIEAppIconsAtom)
    const [showMultiBroadcastRoutes, setShowMultiBroadcastRoutes] = useAtom(showMultiBroadcastRoutesAtom)

    // When the user changes the selection of the IP address families to show,
    // then we first need to translate the selection value from its string
    // representation of an array into an array with address family enum
    // values. Then we can flip the settings depending on what we can find in
    // that array.
    const handleFamilyChange = (event: SelectChangeEvent<string>) => {
        const families = event.target.value.split(',').map(fam => parseInt(fam, 10))
        setIpv4(families.includes(AddressFamily.IPv4))
        setIpv6(families.includes(AddressFamily.IPv6))
    }

    const handleThemeChange = (event: SelectChangeEvent<number>) => {
        setTheme(event.target.value as number)
    }

    return (
        <Box m={1} overflow="auto">
            <SettingsGrid
                container
                direction="row"
                justifyContent="center"
            >
                <Grid
                    container
                    direction="column"
                    style={{ minWidth: '35em', maxWidth: '60em' }}
                >
                    <Typography variant="subtitle1">Appearance</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Theme" />
                                <ListItemSecondaryAction>
                                    <Select size="small" value={theme} onChange={handleThemeChange}>
                                        <MenuItem value={THEME_USERPREF}>user preference</MenuItem>
                                        <MenuItem value={THEME_LIGHT}>light</MenuItem>
                                        <MenuItem value={THEME_DARK}>dark</MenuItem>
                                    </Select>
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Display Filters Network Layer</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Show IP address families" />
                                <ListItemSecondaryAction>
                                    <Select size="small" value={showIpFamilies.toString()} onChange={handleFamilyChange}>
                                        <MenuItem value={[AddressFamily.IPv4].toString()}>IPv4 only</MenuItem>
                                        <MenuItem value={[AddressFamily.IPv4, AddressFamily.IPv6].toString()}>IPv4 and IPv6</MenuItem>
                                        <MenuItem value={[AddressFamily.IPv6].toString()}>IPv6 only</MenuItem>
                                    </Select>
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Show multicast and broadcast routes" />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showMultiBroadcastRoutes}
                                        onChange={() => setShowMultiBroadcastRoutes(!showMultiBroadcastRoutes)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Display Filters Data Link Layer</Typography>
                    <Card>
                        <List>

                            <ListItem>
                                <ListItemText primary='Show "isolated" network namespaces with only "lo"' />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showEmptyNetns}
                                        onChange={() => setShowEmptyNetns(!showEmptyNetns)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary='Show "lo" interfaces' />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showLoopbacks}
                                        onChange={() => setShowLoopbacks(!showLoopbacks)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary='Show MAC addresses unless 00:00:00:00:00:00' />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showMAC}
                                        onChange={() => setShowMAC(!showMAC)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Containee Details</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Show &quot;sandbox&quot; containers" />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showSandboxes}
                                        onChange={() => setShowSandboxes(!showSandboxes)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Show namespace identifiers" />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showNamespaceIds}
                                        onChange={() => setShowNamespaceIds(!showNamespaceIds)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Detail Sections Default Expansion</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Expand containees section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={containeesCutoffAtom} element="containee" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand neighborhood services section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={neighborhoodCutoffAtom} element="service" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand forwarded transport ports section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={forwardedPortsCutoffAtom} element="port" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand transport ports section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={portsCutoffAtom} element="port" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand routes section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={routesCutoffAtom} element="route" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand network interfaces section" />
                                <ListItemSecondaryAction>
                                    <CutoffSelector atom={nifsCutoffAtom} element="interface" em={cutOffEm} />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Siemens Industrial Edge</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Experimental: Show IE Apps icons and titles" />
                                <ListItemSecondaryAction>
                                    <Toggle
                                        checked={showIEAppIcons}
                                        onChange={() => setShowIEAppIcons(!showIEAppIcons)}
                                        color="primary"
                                    />
                                </ListItemSecondaryAction>
                            </ListItem>
                        </List>
                    </Card>
                </Grid>
            </SettingsGrid>
        </Box>
    );
}

export default Settings
