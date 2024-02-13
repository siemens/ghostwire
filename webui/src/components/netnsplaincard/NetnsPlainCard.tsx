// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { ChangeEvent, useState } from 'react'

import { Box, Checkbox, Fade, IconButton, styled, Tooltip } from '@mui/material'
import { Paper } from '@mui/material'
import FullscreenIcon from '@mui/icons-material/Fullscreen'

import { ContaineeBadge } from 'components/containeebadge'
import { PrimitiveContainee, containeeKey, netnsId, NetworkNamespace, orderNifByName, NetworkInterface, AddressFamilySet, containeesOfNetns, sortContaineesByName, Containee } from 'models/gw'
import { useContextualId } from 'components/idcontext'
import { NifBadge } from 'components/nifbadge'
import { NifNavigator } from 'components/nifnavigator'
import { RoutingEtcIcon } from 'icons/RoutingEtc'
import CaptureMultiIcon from 'icons/CaptureMulti'
import CaptureMultiOnIcon from 'icons/CaptureMultiOn'
import { TargetCapture } from 'components/targetcapture'
import { useDynVars } from 'components/dynvars'


// During development The (Capturing) Monolith can be enabled using
// VITE_APP_ENABLE_MONOLITH.
const forceMonolith = !!import.meta.env.VITE_APP_ENABLE_MONOLITH


const NetnsPaper = styled(Paper)(({ theme }) => ({
    // All information on a paper (card) for a network namespace gets the
    // usual inner padding between information and the paper edges.
    padding: theme.spacing(2),
}))

// Style the containees which are listed in form of badges along the top
// edge of any network namespace card.
const Containees = styled('div')(({ theme }) => ({
    position: 'relative',
    left: -theme.spacing(1),
    marginBottom: theme.spacing(2),
    display: 'flex',
    flexWrap: 'wrap',
    // Since CSS flex boxes don't offer "collapsible margins", we have to
    // emulate them by offsetting the margins by half the desired spacing
    // between the flex items (the containees). See also:
    // https://developer.mozilla.org/en-US/docs/Web/CSS/CSS_Flexible_Box_Layout/Mastering_Wrapping_of_Flex_Items
    margin: -theme.spacing(1),
    // For correctly spacing the list of containees, we assume that the
    // BoxedBadge components are immediate children of the containees
    // "container"; thus, we don't care about the specific element type, but
    // simply apply to the immediate children.
    '& > *': {
        flex: '0 0 auto',
        margin: theme.spacing(1),
        marginBottom: 0,
    }
}))

// A two-column grid is used for rendering network interfaces: bridges end
// up in the "left column", while any other network interface ends up in the
// "right column". However, we actually use *three* columns, where the first
// column is used to consume any available surplus space, the second column
// takes the role of the "left column", and the third column is the "right
// column".
const Nifs = styled('div')(({ theme }) => ({
    width: '100%',
    // Turns out that grids are quite versatile without the overhead of
    // tables, when the specialities of tables and their cells are note
    // required (or even causing issues).
    display: 'grid',
    // We use an empty first column which despite being empty swallows any
    // available surplus horizontal space. This ensures that especially
    // the network interface grid children won't be stretched horizontally
    // beyond what really is necessary to fit all network interfaces into
    // the third column, yet the column will make enough room to
    // accommodate the network interface badges.
    gridAutoColumns: '1fr auto auto',
    gridAutoRows: 'auto',
    columnGap: 0, // ...must be zero to avoid visible background gap
    rowGap: theme.spacing(1),
}))

const RoutingEtc = styled('div')(({ theme }) => ({
    gridColumn: '1/2',
    justifySelf: 'end',
    alignSelf: 'stretch',

    display: 'flex',
    alignItems: 'center',

    marginRight: theme.spacing(1),
    borderWidth: `1px`,
    borderStyle: `dotted`,
    borderColor: theme.palette.divider,
    borderRightWidth: 0,
    borderTopLeftRadius: theme.spacing(1),
    borderBottomLeftRadius: theme.spacing(1),
}))

// Either a single network interface or a "placeholder" for a list of bridge
// port network interfaces ("enslaved network interfaces").
const Nif = styled('div')(({ theme }) => ({
    gridColumn: '3/4',
    justifySelf: 'stretch',
    // Ensure that the "cell" in which the nif badge sits will correctly
    // stretch to fill the column height; this covers the situation where
    // there is only a single network interface controlled by a bridge and
    // the bridge badge needs multiple lines, such as due to an alias name
    // and not enough horizontal room to render the bridge badge on a single
    // line.
    alignSelf: 'stretch',
    '&.empty': {
        justifySelf: 'stretch',
        alignSelf: 'stretch',
        background: theme.palette.bridgepaper,
        borderTop: `1px solid ${theme.palette.divider}`,
        paddingTop: theme.spacing(1),
        borderBottom: `1px solid ${theme.palette.divider}`,
        paddingBottom: theme.spacing(1),
    },
    '&.bridged': {
        background: theme.palette.bridgepaper,
    },
}))

// Bridge network interfaces end up in the "left column", which actually is
// the second column, as the first column is solely for eating up surplus
// free horizontal space. 
const Bridge = styled(Nif)(({ theme }) => ({
    gridColumn: '2/3',
    justifySelf: 'stretch',
    alignSelf: 'stretch',
    borderTop: `1px solid ${theme.palette.divider}`,
    borderLeft: `1px solid ${theme.palette.divider}`,
    borderBottom: `1px solid ${theme.palette.divider}`,
    borderTopLeftRadius: theme.spacing(1),
    borderBottomLeftRadius: theme.spacing(1),
    paddingTop: theme.spacing(1),
    paddingLeft: theme.spacing(1),
    paddingBottom: theme.spacing(1),
    paddingRight: theme.spacing(2),
    background: theme.palette.bridgepaper,

    '& > span, & > span > span': {
        width: '100%',
    },

    '& .nifcaptureicon': {
        background: theme.palette.background.paper,
    },

    // Allow the nif badge itself to take on the paper background regardless of
    // whether the badge is inside a bridge "row" or not.
    '& > span > span, & > span > button': {
        background: theme.palette.background.paper,
    },
}))

const BridgePort = styled(Nif)(({ theme }) => ({
    //  Flex box containing a set of network interfaces enslaved to a bridge.
    //  This flex box is placed in the "right column" next to its bridge network
    //  interface.
    borderTop: `1px solid ${theme.palette.divider}`,
    paddingTop: theme.spacing(1),
    borderBottom: `1px solid ${theme.palette.divider}`,
    paddingBottom: theme.spacing(1),
    display: 'flex',
    flexDirection: 'column',
    flexWrap: 'nowrap',
    alignItems: 'stretch',
    background: theme.palette.bridgepaper,

    // place a "gutter" between adjacent port network interfaces of the same
    // bridge. As we're inside a column-oriented flex box, this can be done
    // easily using a top margin, no need for another grid. 
    '& > span + span': {
        marginTop: theme.spacing(1),
    },

    '& > span > button, & > span > span, & .nifcaptureicon': {
        background: theme.palette.background.paper,
    },
}))

const StretchedNif = styled(NifNavigator)(() => ({
    width: '100%',
}))

const CardButtonBox = styled(Box)(({ theme }) => ({
    float: 'right',
    // Move the right-floated button partly back into the padding, as the
    // button has a large "corona" and we thus get a more pleasing visual
    // alignment, especially with the tentant (boxed entity) badges along
    // the top of the network namespace card.
    position: 'relative',
    top: theme.spacing(-1),
    right: theme.spacing(-1),
}))

export interface NetnsPlainCardProps {
    /** 
     * network namespace object (with tons of information thanks to the
     * Ghostwire discovery engine), but only the information about the boxed
     * entities (containers, et cetera) and the network interfaces is used
     * and rendered.
     */
    netns: NetworkNamespace
    /**
     * callback triggering when the user clicks on the zoom button for this
     * network namespace card. The zoom button only gets shown when the onZoom
     * callback is set, otherwise it won't be rendered at all.
     */
    onNetnsZoom?: (netns: NetworkNamespace, fragment?: string) => void
    /**
     * callback triggering when the user clicks on the zoom button of one of
     * the containee cards.
     */
    onContaineeZoom?: (containee: PrimitiveContainee) => void
    /** 
     * callback triggering when the user chooses to navigate from one network
     * interface to another related network interface, such as a VETH peer,
     * MACVLAN (slave), MACVLAN master, et cetera.
     */
    onNavigation?: (nif: NetworkInterface) => void
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** CSS class name(s). */
    className?: string
}

/**
 * Component `NetnsPlainCard` renders the network interfaces of a network
 * interface on a card (a piece of paper), for use in the wiring view. It
 * shows the following details:
 * - contained entities, such as containers and stand-alone processes.
 * - (hierarchical) list of network interfaces: plain network interfaces, but
 *   without any details, such as addresses. Enslaved bridge network
 *   interfaces are shown as "under the bridge" ;)
 *
 * When compared to a `NetnsDetailCard`, this component doesn't render
 * additional details, such as routes, interface addresses, transport ports,
 * et cetera.
 *
 * This component does neither do zooming in nor navigation, but leaves that
 * up to its parent. Use the callbacks, Luke!
 */
export const NetnsPlainCard = ({ netns, onNavigation, onNetnsZoom, onContaineeZoom, families, className }: NetnsPlainCardProps) => {

    const netnsid = useContextualId(netnsId(netns))

    // Is capturing enabled in the UI?
    const { enableMonolith } = useDynVars()

    const [selectNifs, setSelectNifs] = useState(false)
    const [selectedNifs, setSelectedNifs] = useState<string[]>([])

    const nifsWithoutBridgePorts = Object.values(netns.nifs)
        .filter(nif => !nif.master || nif.master.kind !== 'bridge')
        .sort(orderNifByName)

    const handleRouting = () => {
        onNetnsZoom && onNetnsZoom(netns, 'routes')
    }

    // keep an up-to-date list of selected network interfaces in this network
    // namespace.
    const handleNifChange = (event: React.ChangeEvent<HTMLInputElement>, nif: NetworkInterface) => {
        const idx = selectedNifs.indexOf(nif.name)
        if (event.target.checked) {
            if (idx >= 0) {
                return
            }
            setSelectedNifs([...selectedNifs, nif.name])
            return
        }
        if (idx < 0) {
            return
        }
        const dup = [...selectedNifs]
        dup.splice(idx, 1)
        setSelectedNifs(dup)
    }

    // Render the individual network interfaces, with bridge network
    // interfaces "bridging" their port ("enslaved") network interfaces.
    const renderNifs = () => {
        let row: number = 0
        const nifs = nifsWithoutBridgePorts.map(nif => {
            row++
            if (nif.kind !== 'bridge') {
                // a "stand-alone" network interface which isn't a port of a
                // bridge...
                return <Nif
                    key={nif.name}
                    style={{ gridRow: `${row}/${row + 1}` }}
                >
                    <StretchedNif
                        nif={nif}
                        capture={!selectNifs}
                        nifCheck={selectNifs}
                        checked={selectedNifs.includes(nif.name)}
                        anchor
                        stretch
                        alignRight
                        onNavigation={onNavigation}
                        onChange={handleNifChange}
                        families={families}
                    />
                </Nif>
            }
            const slaves = nif.slaves || []
            // Note that a bridge nif badge won't ever be navigateable...
            const bridge = [<Bridge
                key={nif.name}
                style={{ gridRow: `${row}/${row + 1}` }}
            >
                <NifBadge
                    nif={nif}
                    capture={!selectNifs}
                    nifCheck={selectNifs}
                    checked={selectedNifs.includes(nif.name)}
                    anchor
                    families={families}
                    onChange={handleNifChange}
                />
            </Bridge>]
            // If the bridge has ports, then render them and place them into the
            // right column; make sure to classify the first and last port
            // network interface, so we can add some visual separation from
            // other network interfaces not belonging to the same bridge as we
            // do.
            if (slaves.length) {
                // to quote Curious Marc: "big ouchie" ... do sort the bridge
                // ports (slaves), so they don't swap places between refreshes;
                // above, we only sorted the overall list of network interfaces,
                // but not the individual relationship sets.
                const ports = [...slaves].sort(orderNifByName)
                return bridge.concat([
                    <BridgePort
                        key={`port/${nif.name}`}
                        className={"bridged"}
                        style={{ gridRow: `${row}/${row + 1}` }}
                    >
                        {ports.map(nif =>
                            <NifNavigator
                                key={nif.name}
                                nif={nif}
                                capture={!selectNifs}
                                nifCheck={selectNifs}
                                checked={selectedNifs.includes(nif.name)}
                                anchor
                                families={families}
                                stretch
                                alignRight
                                onNavigation={onNavigation}
                                onChange={handleNifChange}
                            />

                        )}
                    </BridgePort>])
            }
            // In case there are currently no bridge port, still render an
            // "empty" element, so we get the visual separation properly
            // applied.
            return bridge.concat([<Nif
                key={`noport/${nif.name}`}
                className="empty"
                style={{ gridRow: `${row}/${row + 1}` }}
            />])
        }).flat()
        nifs.push(
            <RoutingEtc key="routing" style={{ gridRow: `1/${row + 1}` }}>
                <Tooltip title="show routing">
                    <IconButton onClick={handleRouting} size="large">
                        <RoutingEtcIcon />
                    </IconButton>
                </Tooltip>
            </RoutingEtc>)
        return nifs
    }

    // bubble up the zoom navigation, adding the network namespace object to
    // zoom into.
    const handleZoom = () => {
        onNetnsZoom && onNetnsZoom(netns)
    }

    const handleMultiNic = (e: ChangeEvent<HTMLInputElement>) => {
        setSelectNifs(!!e.target.checked)
    }

    return (
        <NetnsPaper className={className}>
            <Box id={netnsid} sx={{ position: 'relative' }}>
                <CardButtonBox>
                    <Box sx={{ display: (enableMonolith || forceMonolith) ? 'inline-block' : 'none' }}>
                        <Fade in={selectNifs} timeout={500}>
                            <Tooltip title="start capture from selected network interfaces">
                                <span>
                                    <TargetCapture
                                        target={netns}
                                        targetNifs={selectedNifs}
                                        disabled={!selectNifs || !selectedNifs.length}
                                    />
                                </span>
                            </Tooltip>
                        </Fade>
                        <Tooltip
                            title={!selectNifs
                                ? "select network interfaces for capture"
                                : "exit selecting network interfaces for capture"}
                        >
                            <Checkbox
                                size="small"
                                checked={selectNifs}
                                icon={<CaptureMultiIcon />}
                                checkedIcon={<CaptureMultiOnIcon />}
                                onChange={handleMultiNic}
                            />
                        </Tooltip>
                    </Box>
                    {/* only render the zoom button when there's a callback for it ;) */}
                    {onNetnsZoom &&
                        <Tooltip title="show only this network namespace">
                            <IconButton onClick={handleZoom} size="small">
                                <FullscreenIcon />
                            </IconButton>
                        </Tooltip>}
                </CardButtonBox>

                {/* render the containees of this network namespace. */}
                <Containees>
                    {containeesOfNetns(netns)
                        .sort(sortContaineesByName)
                        .map(cntr =>
                            <ContaineeBadge
                                containee={cntr}
                                key={containeeKey(cntr)}
                                angled
                                button
                                capture
                                hideCapture={selectNifs}
                                endIcon={<FullscreenIcon />}
                                onClick={onContaineeZoom as (_: Containee) => void}
                            />)}
                </Containees>

                {/* and finally render all the network interfaces in this namespace */}
                <Nifs>{renderNifs()}</Nifs>
            </Box>
        </NetnsPaper >
    );
}

export default NetnsPlainCard
