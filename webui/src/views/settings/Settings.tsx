// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useAtom, useSetAtom } from 'jotai'

import {
    Box,
    Card,
    Divider,
    Grid,
    List,
    ListItem,
    ListItemText,
    MenuItem,
    Select,
    type SelectChangeEvent,
    Stack,
    styled,
    Switch as Toggle,
    Typography,
} from '@mui/material';
import { AddressFamily } from 'models/gw'
import { CutoffSelector } from 'components/cutoffselector'
import {
    THEME_DARK, THEME_LIGHT, THEME_USERPREF, themeAtom,
    containeesCutoffAtom, forwardedPortsCutoffAtom, neighborhoodCutoffAtom,
    nifsCutoffAtom, portsCutoffAtom, routesCutoffAtom,
    showEmptyNetnsAtom, showIEAppIconsAtom, showIpFamiliesAtom, showIpv4Atom,
    showIpv6Atom, showLoopbackAtom, showMACAtom, showMultiBroadcastRoutesAtom,
    showNamespaceIdsAtom, showSandboxesAtom,
} from './atoms'

const cutOffEm = 12

const SettingsGrid = styled(Grid)(({ theme }) => ({
    width: `calc(100% - calc(${theme.spacing(2)} * 2))`,
    margin: theme.spacing(2),

    '& .MuiCard-root + .MuiTypography-subtitle1': {
        marginTop: theme.spacing(4),
    },

    '& .MuiListItem-root': {
        gap: theme.spacing(2),
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
        <Box sx={{ m: 1, overflow: 'auto' }}>
            <SettingsGrid
                container
                direction="row"
                sx={{ justifyContent: 'center' }}
            >
                <Stack style={{ minWidth: '35em', maxWidth: '60em' }}>
                    <Typography variant="subtitle1">Appearance</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Theme" />
                                <Select size="small" value={theme} onChange={handleThemeChange}>
                                    <MenuItem value={THEME_USERPREF}>user preference</MenuItem>
                                    <MenuItem value={THEME_LIGHT}>light</MenuItem>
                                    <MenuItem value={THEME_DARK}>dark</MenuItem>
                                </Select>
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Display Filters Network Layer</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Show IP address families" />
                                <Select size="small" value={showIpFamilies.toString()} onChange={handleFamilyChange}>
                                    <MenuItem value={[AddressFamily.IPv4].toString()}>IPv4 only</MenuItem>
                                    <MenuItem value={[AddressFamily.IPv4, AddressFamily.IPv6].toString()}>IPv4 and IPv6</MenuItem>
                                    <MenuItem value={[AddressFamily.IPv6].toString()}>IPv6 only</MenuItem>
                                </Select>
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Show multicast and broadcast routes" />
                                <Toggle
                                    checked={showMultiBroadcastRoutes}
                                    onChange={() => setShowMultiBroadcastRoutes(!showMultiBroadcastRoutes)}
                                    color="primary"
                                />
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Display Filters Data Link Layer</Typography>
                    <Card>
                        <List>

                            <ListItem>
                                <ListItemText primary='Show non-container "isolated" network namespaces with only "lo"' />
                                <Toggle
                                    checked={showEmptyNetns}
                                    onChange={() => setShowEmptyNetns(!showEmptyNetns)}
                                    color="primary"
                                />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary='Show "lo" interfaces' />
                                <Toggle
                                    checked={showLoopbacks}
                                    onChange={() => setShowLoopbacks(!showLoopbacks)}
                                    color="primary"
                                />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary='Show MAC addresses unless 00:00:00:00:00:00' />
                                <Toggle
                                    checked={showMAC}
                                    onChange={() => setShowMAC(!showMAC)}
                                    color="primary"
                                />
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Containee Details</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Show &quot;sandbox&quot; containers" />
                                <Toggle
                                    checked={showSandboxes}
                                    onChange={() => setShowSandboxes(!showSandboxes)}
                                    color="primary"
                                />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Show namespace identifiers" />
                                <Toggle
                                    checked={showNamespaceIds}
                                    onChange={() => setShowNamespaceIds(!showNamespaceIds)}
                                    color="primary"
                                />
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Detail Sections Default Expansion</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Expand containees section" />
                                <CutoffSelector atom={containeesCutoffAtom} element="containee" em={cutOffEm} />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand neighborhood services section" />
                                <CutoffSelector atom={neighborhoodCutoffAtom} element="service" em={cutOffEm} />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand forwarded transport ports section" />
                                <CutoffSelector atom={forwardedPortsCutoffAtom} element="port" em={cutOffEm} />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand transport ports section" />
                                <CutoffSelector atom={portsCutoffAtom} element="port" em={cutOffEm} />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand routes section" />
                                <CutoffSelector atom={routesCutoffAtom} element="route" em={cutOffEm} />
                            </ListItem>
                            <Divider />
                            <ListItem>
                                <ListItemText primary="Expand network interfaces section" />
                                <CutoffSelector atom={nifsCutoffAtom} element="interface" em={cutOffEm} />
                            </ListItem>
                        </List>
                    </Card>

                    <Typography variant="subtitle1">Siemens Industrial Edge</Typography>
                    <Card>
                        <List>
                            <ListItem>
                                <ListItemText primary="Experimental: Show IE Apps icons and titles" />
                                <Toggle
                                    checked={showIEAppIcons}
                                    onChange={() => setShowIEAppIcons(!showIEAppIcons)}
                                    color="primary"
                                />
                            </ListItem>
                        </List>
                    </Card>
                </Stack>
            </SettingsGrid>
        </Box>
    );
}

export default Settings
