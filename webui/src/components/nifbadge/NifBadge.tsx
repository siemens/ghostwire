// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import clsx from 'clsx'

import { Button, styled } from '@mui/material'
import HearingIcon from '@mui/icons-material/Hearing'

import { DormantIcon, DownIcon, LowerLayerDownIcon, UpIcon } from 'icons/operstates'
import { BridgeIcon, BridgeInternalIcon, DummyIcon, HardwareNicIcon, HardwareNicPFIcon, HardwareNicVFIcon, MacvlanIcon, MacvlanMasterIcon, NicIcon, OverlayIcon, TapIcon, TunIcon, VethIcon } from 'icons/nifs'

import { AddressFamily, AddressFamilySet, GHOSTWIRE_LABEL_ROOT, NetworkInterface, nifId, orderAddresses, SRIOVRole } from 'models/gw'
import { OperationalState } from 'models/gw'
import { useContextualId } from 'components/idcontext'
import { TooltipWrapper } from 'utils/tooltipwrapper'
import { relationClassName } from 'utils/relclassname'
import { rgba } from 'utils/rgba'
import { TargetCapture } from 'components/targetcapture'


// The outer span holding together an optional "hardware" NIC icon as well
// as the network interface badge itself. In order to be able to stretch the
// badge to take up the available space, but without overflowing or pushing
// an optional hardware icon aside, we use an inline flex box. The badge
// flex child then is allowed to stretch if the corresponding component
// property "stretch" has been set.
const Badge = styled('span')(({ theme }) => ({
    display: 'inline-flex',
    flexDirection: 'row',
    alignItems: 'center',
    //whiteSpace: 'nowrap',
    // ...ensure that we DON'T inherit any hanging indentations! :(
    padding: '0',
    textIndent: '0',

    '& .nifcaptureicon + $nifbadge': {
        marginLeft: theme.spacing(0.5),
    },
    '& $nifbadge + .nifcaptureicon': {
        marginLeft: theme.spacing(0.5),
    },
}))

// (Optional) "hardware" NIC icon preceeding the network interface badge,
// if a network interface has been marked as "physical". Whatever
// "physical" is worth in our virtualized world...
const HWNif = styled('span')(({ theme }) => ({
    display: 'inline-block',
    paddingRight: '0.2em',
    color: theme.palette.text.secondary,
    '& > .MuiSvgIcon-root': {
        verticalAlign: 'middle',
    },
}))

const Promiscuous = styled('span')(({ theme }) => ({
    display: 'inline-block',
    paddingRight: '0.2em',
    color: theme.palette.text.secondary,
    '& > .MuiSvgIcon-root': {
        verticalAlign: 'middle',
    }
}))

const Capture = styled(TargetCapture)(({ theme }) => ({
    marginLeft: '0.2em',
    '&.alignright': {
        marginLeft: 0,
        marginRight: '0.2em',
    }
}))

const Nif = styled(Button)(({ theme }) => ({
    // General+basis badge styling...
    display: 'inline-block',
    paddingTop: '0.1ex',
    paddingBottom: '0.1ex',
    paddingLeft: '0.3em',
    paddingRight: '0.3em',
    borderColor: theme.palette.mode === 'light' ? 'rgba(0, 0, 0, 0.23)' : 'rgba(255, 255, 255, 0.23)',
    borderWidth: '1px',
    borderStyle: 'solid',

    // same alignment as a MUI button in case we render NifBadges as
    // non-references as well as references in the same block.
    verticalAlign: 'middle',

    // line height of Material UI buttons...
    lineHeight: '1.75',

    // Nothing more annoying than a bunch of veths with their proportional
    // font names not aligning, resulting in jittery UX.
    fontFamily: 'Roboto Mono',

    // Style the nif name and alias differently for better
    // differentiation...
    '& .name': { fontWeight: 'bold' },
    '& .alias': { fontStyle: 'italic' },

    // Ensure to use the proper button appearance when linking to a
    // network interface.
    '&.reference': {
        borderRadius: theme.shape.borderRadius,
        fontWeight: 'normal',
        textTransform: 'none',
        color: theme.palette.text.primary, // ensure to keep the primary text color, instead of button color.
    },

    '&.stretch': {
        width: '100%',
    },
    '&.alignright': {
        textAlign: 'right',
    },

    [`&.${OperationalState.Unknown.toLowerCase()}`]: {
        borderColor: theme.palette.mode === 'light' ?
            rgba(theme.palette.operstate.unknown, 0.23) : 'rgba(155, 255, 155, 0.23)',
    },
    [`&.${OperationalState.Up.toLowerCase()}`]: {
        borderColor: theme.palette.mode === 'light' ?
            rgba(theme.palette.operstate.up, 0.23) : 'rgba(155, 255, 155, 0.23)',
    },
    [`&.${OperationalState.Down.toLowerCase()}`]: {
        borderColor: theme.palette.mode === 'light' ?
            rgba(theme.palette.operstate.down, 0.23) : 'rgba(255, 116, 116, 0.23)',
    },
    [`&.${OperationalState.LowerLayerDown.toLowerCase()}`]: {
        borderColor: theme.palette.mode === 'light' ?
            rgba(theme.palette.operstate.lowerlayerdown, 0.23) : 'rgba(255, 116, 116, 0.23)',
    },
    [`&.${OperationalState.Dormant.toLowerCase()}`]: {
        borderColor: theme.palette.mode === 'light' ?
            rgba(theme.palette.operstate.dormant, 0.23) : 'rgba(37, 90, 223, 0.23)',
    },

    // Reduce opacity of the type and operstate icons so that they become
    // slightly a less prominent size compared to the following text.
    '& .MuiSvgIcon-root': {
        verticalAlign: 'middle',
        fillOpacity: '0.6',
    },
    '& .MuiSvgIcon-root.hwnif': {
    },

    '& .nif-alias-proxy-redirect': {
        // TODO: highlight
    },
}))

// The operational state indicator appears to either the right or left of
// a network interface badge and signals the operational state of the
// interface. Different states are signalled by different icons which we
// additionally color according to state to foster better understanding of
// the indicators.
const OperstateIndicator = styled('span')(({ theme }) => ({
    verticalAlign: 'middle',
    marginRight: '0.1em',
    [`&.${OperationalState.Unknown.toLowerCase()}`]: { color: theme.palette.operstate.unknown },
    [`&.${OperationalState.Up.toLowerCase()}`]: { color: theme.palette.operstate.up },
    [`&.${OperationalState.Down.toLowerCase()}`]: { color: theme.palette.operstate.down },
    [`&.${OperationalState.LowerLayerDown.toLowerCase()}`]: { color: theme.palette.operstate.lowerlayerdown },
    [`&.${OperationalState.Dormant.toLowerCase()}`]: { color: theme.palette.operstate.dormant },
}))

const nifSRIOVIcons = {
    [SRIOVRole.None]: HardwareNicIcon,
    [SRIOVRole.PF]: HardwareNicPFIcon,
    [SRIOVRole.VF]: HardwareNicVFIcon,
}

// Known network interface type icons, indexed by the kind property of network
// interface objects (and directly taken from what Linux' RTNETLINK tells us).
const nifTypeIcons = {
    'bridge': BridgeIcon,
    'dummy': DummyIcon,
    'macvlan': MacvlanIcon,
    'tap': TapIcon,
    'tun': TunIcon,
    'veth': VethIcon,
    'vxlan': OverlayIcon,
}

const nifIcon = (nif: NetworkInterface) => {
    if (GHOSTWIRE_LABEL_ROOT + 'bridge/internal' in nif.labels) {
        return BridgeInternalIcon
    }
    return (nif.tuntapDetails && nifTypeIcons[nif.tuntapDetails.mode]) || 
        (nif.macvlans && MacvlanMasterIcon) ||
        nifTypeIcons[nif.kind] || NicIcon
}

const operStateIcons = {
    [OperationalState.Unknown]: UpIcon,
    [OperationalState.Dormant]: DormantIcon,
    [OperationalState.Down]: DownIcon,
    [OperationalState.LowerLayerDown]: LowerLayerDownIcon,
    [OperationalState.Up]: UpIcon,
}

const nifKindTips = {
    'hw': '(virtual) hardware',
    'pf': 'SR-IOV PF hardware',
    'vf': 'SR-IOV VF hardware',
    
    'lo': 'loopback',
    
    'bridge': 'virtual bridge',
    'dummy': 'all packets swallowing dummy',
    'macvlan': 'MACVLAN',
    'tap': 'layer 2 TAP',
    'tun': 'layer 3 TUNnel',
    'veth': 'virtual peer-to-peer Ethernet',
    'vlan': 'virtual LAN',
    'vxlan': 'VXLAN overlay',
}

export interface NifBadgeProps {
    /** network interface object describing a network interface in detail. */
    nif: NetworkInterface
    /** CSS class name(s). */
    className?: string
    /** CSS properties, for instance, to allow grid placement. */
    style?: React.CSSProperties
    /** put an ID to this badge in any case (for placing the wiring). */
    anchor?: boolean
    /** 
     * render the badge as a clickable button, instead of just a passive
     * badge.
     */
    button?: boolean
    /** if true, disable the tooltip of this network interface badge. */
    notooltip?: boolean
    /** optionally show a capture button? */
    capture?: boolean
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*) as part of the tooltip. If left undefined, then it
     * defaults to showing both IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** 
     * if true, stretch the network interface badge to fill available
     * horizontal space. 
     */
    stretch?: boolean
    /** right align button contents when stretched. */
    alignRight?: boolean
    /** 
     * An element to place to the right of the badge contents; can be used to
     * show a navigation icon, et cetera. 
     */
    endIcon?: React.ReactNode
    /** 
     * optional callback handler: when set, the callback will be fired when
     * the user clicks on the badge.
     */
    onClick?: (event: React.MouseEvent<HTMLElement>, nif: NetworkInterface) => void
}

/**
 * The `NifBadge` component renders a single network interface "badge",
 * including the operational status of the interface, as well as an
 * interface-type specific icon.
 *
 * A "hardware" NIC icon is shown for "physical" network interfaces. Of
 * course, in our modern world, "physical" hardware might be as virtual as a
 * VETH interface. Thus, "hardware" refers to an interface having an
 * associated "hardware" driver, as opposed to the other kinds of virtual
 * network interfaces, such as bridges and VETHs.
 *
 * When a network interface operates in promiscuous mode, then an ears icon
 * gets shown.
 *
 * Supported kinds of network interfaces:
 * - **generic**: that is, `kind===""` or `kind==="lo"`, results in a generic
 *   "Ethernet" plug icon being shown.
 *   - **hardware**: an additionally "NIC" (network interface card) is
 *     rendered next to the interface badge.
 *   - **`macvlan`** master: the interface type icon changes from the generic
 *       "Ethernet" plug type with a single cable to one with two cables.
 * - **`veth`** peer: the interface type icon shows a pair of plugs.
 * - **`macvlan`**: the type icon shows a down-pointing plug.
 *
 * `NifBadge`s automatically get (additional) CSS class names reflecting their
 * relationships with other network interfaces (making use of the
 * `relationClassName()` utility function). Simply put, these interface
 * relation CSS classes are generated from the network namespace IDs and
 * interface names related. To ensure predictability, the lexicographically
 * earlier combination of namespace ID and interface name will always be put
 * first in the generated class name. Relation class names are generated only
 * in these situations:
 * - "physical" network interface (which has a relation to the external
 *   "world").
 * - veth pair.
 * - MACVLAN network interface and its MACVLAN master.
 * - VXLAN.
 * - TODO: TAPTUN.
 */
export const NifBadge = ({
    nif,
    button,
    anchor,
    className,
    style,
    notooltip,
    capture,
    families,
    stretch,
    alignRight,
    endIcon,
    onClick
}: NifBadgeProps) => {
    // We might later need the contextual base DOM element ID for constructing
    // the DOM element identifiers of related network interfaces in order to
    // be able to scroll them into view.
    const domIdBase = useContextualId('')
    const nifDomId = domIdBase + nifId(nif)

    const name = <span className="name">{nif.name}</span>
    const alias = (nif.alias && nif.alias !== "") && <> (~<span className="alias">{nif.alias}</span>)</>
    const vid = (nif.vlanDetails) && <> VID&nbsp;{nif.vlanDetails.vid}</>

    const NifIcon = nifIcon(nif)
    const OperstateIcon = operStateIcons[nif.operstate]

    const content = <>
        <NifIcon />
        <OperstateIndicator
            as={OperstateIcon}
            className={nif.operstate.toLowerCase()}
        />
        {name}{alias}{vid}
    </>

    families = families || [AddressFamily.IPv4, AddressFamily.IPv6]

    const tooltipBase = (nifKindTips[
        (nif.kind === 'tuntap' && nif.tuntapDetails.mode)
        || nif.kind
        || (nif.sriovrole === SRIOVRole.PF && 'pf')
        || (nif.sriovrole === SRIOVRole.VF && 'vf')
        || (nif.isPhysical && 'hw')
        || (nif.name === 'lo' && 'lo')
    ])
    const tooltipInternalNetwork = (nif.labels[GHOSTWIRE_LABEL_ROOT + 'bridge/internal'] === 'True') ? 'internal ' : ''
    const withMacvlans = (tooltipBase && nif.macvlans && ' with additional MACVLAN interface(s)') || ''
    const withOverlays = (tooltipBase && nif.overlays && ' with VXLAN overlay(s)') || ''
    const vxlanDetails = (nif.vxlanDetails && ` ID ${nif.vxlanDetails.vid}`) || ''
    const nifAddrs = nif.addresses
        .filter(addr => addr.family !== AddressFamily.MAC)
        .filter(addr => families.includes(addr.family))
        .sort(orderAddresses)
        .map(addr => <span key={addr.address}>{addr.address}<br /></span>)
    const tooltip = <>
        {(tooltipBase ? tooltipInternalNetwork + tooltipBase + ' network interface' : '')}{vxlanDetails}{withOverlays}{withOverlays && withMacvlans ? ' and' : ''}{withMacvlans}<br />
        {nifAddrs}
    </>

    const stretchBadgeClass = stretch ? 'stretch' : ''
    const alignRightClass = alignRight ? 'alignright' : ''
    const aliasClass = (nif.alias && nif.alias !== "") ? `nif-alias-${nif.alias}` : ''

    // In order to support the hover functionality in wiring views, where the
    // use can hover over any part of a wire and connected network interfaces
    // and all connected parts will light up, we tack on one or more CSS
    // classes based on the particular relations our interface is part of.
    let relationClasses: string[] = []
    if (nif.macvlan) { // ...we're one of the MACVLANs.
        relationClasses.push(relationClassName(domIdBase, nif, nif.macvlan))
    }
    if (nif.macvlans) { // ...we're the master of some MACVLANs.
        relationClasses.push(...nif.macvlans.map(macvlan => relationClassName(domIdBase, nif, macvlan)))
    }
    if (nif.overlays) { // ...and have some overlays.
        relationClasses.push(...nif.overlays.map(vxlan => relationClassName(domIdBase, nif, vxlan)))
    }
    if (nif.peer) { // ...peer-to-peer virtual Ethernet cable relationship.
        relationClasses.push(relationClassName(domIdBase, nif, nif.peer))
    }
    if (nif.underlay) { // ...we're the VXLAN overlay, because we have an underlay :)
        relationClasses.push(relationClassName(domIdBase, nif, nif.underlay))
    }
    if (nif.isPhysical) { // ..."external" relation.
        switch (nif.sriovrole) {
            case SRIOVRole.PF:
                if (nif.slaves) {
                    relationClasses.push(...nif.slaves.filter(slave => slave.pf === nif)
                        .map(vf => relationClassName(domIdBase, nif, vf)))
                }
                relationClasses.push(relationClassName(domIdBase, nif))
                break
            case SRIOVRole.VF:
                relationClasses.push(relationClassName(domIdBase, nif, nif.pf))
                break
            default:
                relationClasses.push(relationClassName(domIdBase, nif))
                break;
        }
    }

    const handleBadgeClick = (event: React.MouseEvent<HTMLElement, MouseEvent>) => {
        if (onClick) {
            onClick(event, nif)
        }
    }

    const HWIcon = nifSRIOVIcons[nif.sriovrole || SRIOVRole.None]

    // With lots of information prepared we can finally render the badge,
    // optionally wrapped into a tooltip with some detail information about
    // the network interface.
    return TooltipWrapper(
        <Badge
            className={clsx(className, relationClasses, aliasClass)}
            style={style}
            id={anchor || !button ? nifDomId : ''}
        >
            {capture && alignRight && <Capture className="nifcaptureicon alignright" target={nif} />}
            {nif.isPhysical &&
                <HWNif className="nifbagdeicon">
                    <HWIcon />
                </HWNif>
            }
            {nif.isPromiscuous &&
                <Promiscuous className="nifbagdeicon">
                    <HearingIcon />
                </Promiscuous>
            }
            {(button &&
                <Nif
                    variant="outlined"
                    className={clsx('reference', stretchBadgeClass, alignRightClass)}
                    onClick={handleBadgeClick}
                >{content}{endIcon}</Nif>
            ) ||
                <Nif
                    as="span"
                    className={clsx(nif.operstate.toLowerCase(), stretchBadgeClass, alignRightClass)}
                >{content}{endIcon}</Nif>
            }
            {capture && !alignRight && <Capture className="nifcaptureicon" target={nif} />}
        </Badge>
        , tooltip, !notooltip)
}
