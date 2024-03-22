// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { AddressFamilySet, PrimitiveContainee, NetworkInterface, SRIOVRole } from 'models/gw'
import { NifBadge } from 'components/nifbadge'
import { NamespaceContainees } from 'components/namespacecontainees'
import { styled } from '@mui/material'

const Via = styled(NifBadge)(({ theme }) => ({
    marginLeft: theme.spacing(1),
}))

export interface RelatedNifProps {
    /** 
     * network interface object (A) for which to render the details of the
     * related network interface (R). This parameter is **not** the related
     * interface (R) itself; the related interface (R) will be correctly
     * determined for suitable (A)s.
     */
    nif: NetworkInterface
    /**
     * optional network interface object to which we should not relate back;
     * this breaks funny locking cycles especially on VxLAN interfaces enslaved
     * to a bridge and at the same time enslaved to their master underlay
     * interface: don't relate back to the master when we're not shown
     * subordinate to a bridge.
     */
    unrelatedNif?: NetworkInterface
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
    /** CSS class name(s). */
    className?: string
}

/**
 * The `RelatedNif` component renders the network interface badge and optionally
 * "boxed entites" (containers, stand-alone processes) of the network namespace
 * the related network interface is in. The boxed entities are only rendered if
 * the related interface is in a different network namespace from the specified
 * network interface.
 *
 * If a related network interface acts as a bridge port, then the responsible
 * bridge is shown too. This is especially convenient with multiple Docker
 * compose projects with their own per-project bridges, or even using multiple
 * bridges.
 *
 * Network interfaces with related interfaces are:
 *
 * - **`veth`**: the corresponding peer network interface; this is always a
 *   one-to-one relationship.
 *
 * - **`macvlan`**: the "master" network interface; this is a one-to-many
 *   relationship, where each macvlan network interface has exactly one master
 *   network interface, but the master network interface might have multiple
 *   macvlan interfaces attached to it. Please also note that the master
 *   interface of a macvlan interface is a hardware/"physical" network
 *   interface. When attaching a macvlan interface to another macvlan interface,
 *   the Linux kernel automatically will attach the macvlan interface to the
 *   master hardware interface instead, so there never will be cascades of
 *   macvlans.
 * 
 * - **`vxlan`**: the "underlay" network interface; this is a one-to-many
 *   relationship, where each vxlan network interface has exactly one underlay
 *   network interface, but the underlay network interface might have multiple
 *   "overlays", that is, vxlans.
 *
 * All other kinds of network interfaces don't render any related interfaces.
 */
export const RelatedNif = ({
    nif,
    unrelatedNif,
    families,
    onNavigation,
    onContaineeNavigation,
    className,
}: RelatedNifProps) => {
    let othernif: NetworkInterface | undefined
    switch (nif.sriovrole) {
        case SRIOVRole.VF:
            othernif = nif.pf
            break
        default:
            switch (nif.kind) {
                case 'macvlan':
                    othernif = nif.macvlan
                    break
                case 'veth':
                    othernif = nif.peer
                    break
                case 'vxlan':
                    othernif = nif.underlay
                    break
                default:
                    return <></>
            }
    }

    // trigger the navigation callback with the information about the related
    // network interface the user wants to navigate to.
    const handleOtherNifClick = () => {
        if (onNavigation) {
            onNavigation(othernif!)
        }
    }

    // trigger the navigation callback with information about the master
    // (bridge) network interface when the user wants to navigate to it.
    const handleMasterClick = () => {
        if (onNavigation && othernif!.master) {
            onNavigation(othernif!.master)
        }
    }

    if (!othernif || othernif === unrelatedNif) {
        return <></>
    }
    return (<span className={className || ''}>
        &nbsp;·····&nbsp;
        <NifBadge
            nif={othernif}
            families={families}
            button
            onClick={handleOtherNifClick}
        />

        {/* if this an "enslaved" bridge interface, then show its bridge. */}
        {othernif.master &&
            <Via
                nif={othernif.master}
                families={families}
                button
                onClick={handleMasterClick}
            />}

        {/* 
          * render the containees in the network namespace of the related
          * interface, or at least some of them) 
          */}
        {nif.netns !== othernif.netns &&
            <NamespaceContainees
                netns={othernif.netns}
                key={othernif.netns.netnsid}
                onNavigation={onContaineeNavigation}
            />}
    </span>)
}
