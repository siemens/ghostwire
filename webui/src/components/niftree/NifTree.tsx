// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'
import { TransitionGroup } from 'react-transition-group'

import { Collapse, IconButton, Tooltip, styled } from '@mui/material'
import InfoIcon from '@mui/icons-material/Info'

import { AddressFamily, AddressFamilySet, PrimitiveContainee, NetworkInterface, NetworkNamespace, orderNifByName } from 'models/gw'
import { NifBadge } from 'components/nifbadge'
import { RelatedNif } from 'components/relatednif'
import { NifAddressList } from 'components/nifadresslist'
import { VxlanDetails } from 'components/vxlandetails'
import { NamespaceContainees } from 'components/namespacecontainees'
import { TunTapDetails } from 'components/tuntapdetails'
import { useNifInfoModal } from 'components/nifinfomodal'


// Indent of (1) nif properties, addresses, etc.; (2) sub-level nifs.
const nifPropsIndent = '2em'

// Hanging indentation used when rendering the related network interface and
// containee(s) of a particular network interface.
const nifHangingIndent = '4em'

const NifList = styled('div')(({ theme }) => ({
    '& .nif': {
        marginTop: theme.spacing(1),
    },
    '& .nif > .MuiSvgIcon-root': {
        verticalAlign: 'middle',
        marginRight: '0.3em',
    },
    '& .nif + .nif': {
        marginTop: theme.spacing(1),
    },
    '& .nifandrels': {
        // display network interfaces with their optional related nif,
        // bridge and boxed entities as blocks with hanging indentation in
        // case the route information needs line breaks. Since CSS still has
        // made "hanging" not official -- probably for reasons beyond
        // technical, considering how the styling would read out loud -- we
        // emulate hanging indentation using padding, except for the first
        // line.
        display: 'block',
        paddingLeft: nifHangingIndent,
        textIndent: `-${nifHangingIndent}`,
        lineHeight: '210%',
    },
    '& .nifandrels > *': {
        textIndent: 'initial',
    }
}))

const SubNifs = styled('div')(({ theme }) => ({
    display: 'block',
    marginTop: theme.spacing(1),
    marginLeft: nifPropsIndent,
}))

const Addresses = styled(NifAddressList)(({ theme }) => ({
    marginTop: theme.spacing(1),
    marginLeft: nifPropsIndent,
}))

const TunTapInfo = styled(TunTapDetails)(({ theme }) => ({
    marginTop: theme.spacing(1),
    marginLeft: nifPropsIndent,
}))

const VxlanInfo = styled(VxlanDetails)(() => ({
    marginLeft: nifPropsIndent,
}))

const InfoButton = styled(IconButton)(() => ({
    marginLeft: '0.25em',
}))


interface SubordinateNifsProps {
    /** network namespace object with interfaces. */
    nif: NetworkInterface
    /** if `true`, then hides MAC layer addresses. */
    filterMAC?: boolean
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** 
     * callback triggering when the user wants to navigate to this related
     * network interface.
     */
    onNavigation?: (nif: NetworkInterface) => void
    /**
     * callback triggering when the user wants to navigate to one of the
     * containees of the related network interface.
     */
    onContaineeNavigation?: (containee: PrimitiveContainee) => void
}

/**
 * Component `SubordinateNifs`, given a bridge or macvlan master network
 * interface, then renders a list of "subordinate" network interfaces. If the
 * specified network interface doesn't have subordinate interfaces, then this
 * component renders nothing.
 */
const SubordinateNifs = ({ nif, filterMAC, families, onNavigation, onContaineeNavigation }: SubordinateNifsProps) => {
    const setNifInfo = useNifInfoModal()

    const subnifs = (nif.slaves || [])
        .concat(nif.macvlans || [])
        .concat(nif.overlays || [])
        .sort(orderNifByName)

    const handleNavigation = (nif: NetworkInterface) => {
        if (onNavigation) {
            onNavigation(nif)
        }
    }

    return (subnifs.length > 0 &&
        <SubNifs>
            <TransitionGroup className="subniflist">
                {subnifs.map(nif => {
                    const othernif = nif.macvlan || nif.underlay || nif.pf
                    return <Collapse
                        className="nif"
                        key={`${nif.netns.netnsid}-${nif.name}`}
                        in
                    >
                        <div className="nifandrels">
                            {othernif && <>&nbsp;·····&nbsp;</>}
                            <NifBadge
                                nif={nif}
                                capture={!othernif}
                                button={othernif !== undefined}
                                onClick={() => handleNavigation(nif)}
                                families={families}
                            />
                            <Tooltip title="network interface information">
                                <InfoButton
                                    size="small"
                                    onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
                                        event.stopPropagation()
                                        if (setNifInfo) setNifInfo(nif)
                                    }}><InfoIcon /></InfoButton>
                            </Tooltip>
                            {nif.master &&
                                <RelatedNif
                                    nif={nif}
                                    onNavigation={onNavigation}
                                    onContaineeNavigation={onContaineeNavigation}
                                    families={families}
                                />}
                            {othernif && nif.netns !== othernif.netns &&
                                <NamespaceContainees
                                    netns={nif.netns}
                                    key={nif.netns.netnsid}
                                    onNavigation={onContaineeNavigation}
                                />}
                        </div>

                        <Addresses
                            nif={nif}
                            filterMAC={filterMAC}
                            families={families}
                        />
                    </Collapse>
                })}
            </TransitionGroup>
        </SubNifs>
    )
}

export interface NifTreeProps {
    /** network namespace object with interfaces. */
    netns: NetworkNamespace
    /** if `true`, then hides loopback interfaces. */
    filterLo?: boolean
    /** if `true`, then hides MAC layer addresses. */
    filterMAC?: boolean
    /** 
     * the IP address family/families to show (filter *through*, as opposed to
     * filtering *out*). If left undefined, then it defaults to showing both
     * IPv4 and IPv6.
     */
    families?: AddressFamilySet
    /** 
     * callback triggering when the user wants to navigate to this related
     * network interface.
     */
    onNavigation?: (nif: NetworkInterface) => void
    /**
     * callback triggering when the user wants to navigate to one of the
     * containees of the related network interface.
     */
    onContaineeNavigation?: (containee: PrimitiveContainee) => void
}

/**
 * Component `NifTree` renders the network interfaces belonging to a specific
 * network namespace as a tree-like hierarchy. Here, top-level network
 * interfaces are all interfaces which are not "enslaved"(\*) into a bridge.
 * The network interfaces are sorted (on all levels) according to their names.
 *
 * > As a technical(!) side note, Linux doesn't allow macvlan masters to be
 * > enslaved.
 *
 * Network interface addresses are rendered for all interfaces on the top
 * level, as well as "bridge port" network interfaces on the next sub-level.
 *
 * In particular...
 *
 * - **bridge** interfaces open a sub-hierarchy, consisting of the network
 *   interfaces acting as bridge ports. The "bridge port" network interfaces
 *   are not shown on the top level, but instead only beneath bridge
 *   interfaces; in consequence, their network addresses also shown in the
 *   sub-hierarchy.
 *
 * - **macvlan master** interfaces also open a sub-hierarchy, consisting of
 *   the macvlan interfaces attached to them. However, this sub-hierarchy
 *   consists only of references, so interface addresses are **not shown** for
 *   such references.
 *
 * (\*) Please note that "*enslaved*" is official Linux kernel terminology
 * with regard to network interfaces and their use as "(software) bridge
 * ports". This official terminology is also reflected in RTNETLINK API field
 * names, as well as network tools, such as the set "ip" (sub)commands.
 */
export const NifTree = ({ netns, filterLo, filterMAC, families: fam, onNavigation, onContaineeNavigation }: NifTreeProps) => {
    const setNifInfo = useNifInfoModal()

    const families = fam || [AddressFamily.IPv4, AddressFamily.IPv6]

    // List all network interfaces not acting as bridge ports.
    const toplevelnifs = Object.values(netns.nifs)
        .filter(nif => !nif.master)
        .sort(orderNifByName)

    return (
        <NifList>
            {toplevelnifs.filter(nif => !filterLo || nif.name !== 'lo')
                .map(nif => (
                    <TransitionGroup key={nif.name} component={null}>
                        <Collapse
                            className="nif"
                            key={nif.name}
                            in
                        >
                            {/* 
                              * show the network interface, if it's not enslaved;
                              * users are allowed to capture from these interfaces,
                              * as they're nifs inside this network namespace. 
                              */}
                            <div className="nifandrels">
                                <NifBadge
                                    nif={nif}
                                    families={families}
                                    capture
                                />
                                <Tooltip title="network interface information">
                                    <InfoButton
                                        size="small"
                                        onClick={(event: React.MouseEvent<HTMLButtonElement>) => {
                                            event.stopPropagation()
                                            if (setNifInfo) setNifInfo(nif)
                                        }}><InfoIcon /></InfoButton>
                                </Tooltip>
                                <RelatedNif
                                    nif={nif}
                                    onNavigation={onNavigation}
                                    onContaineeNavigation={onContaineeNavigation}
                                    families={families}
                                />
                            </div>

                            {/* optionally: TAP/TUN details */}
                            {nif.tuntapDetails && <TunTapInfo nif={nif} />}

                            {/* optionally: VXLAN details */}
                            {nif.vxlanDetails && <VxlanInfo nif={nif} />}

                            {/* configured MAC/IP addresses of the network interfaces */}
                            <Addresses
                                nif={nif}
                                filterMAC={filterMAC}
                                families={families}
                            />

                            {/* 
                              * show enslaved bridge network interfaces,
                              * MACVLANs, overlay VXLANs, et cetera only
                              * after the address list of the current 
                              * network interface.
                              */}
                            <SubordinateNifs
                                nif={nif}
                                filterMAC={filterMAC}
                                families={families}
                                onNavigation={onNavigation}
                                onContaineeNavigation={onContaineeNavigation}
                            />
                        </Collapse>
                    </TransitionGroup>
                ))}
        </NifList>
    )
}