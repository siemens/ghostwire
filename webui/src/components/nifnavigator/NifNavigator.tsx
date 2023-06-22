// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { Menu, MenuItem, styled } from '@mui/material'

import { AddressFamilySet, firstContainee, NetworkInterface, nifId, orderNifByNameAndContainee } from 'models/gw'
import { NifBadge } from 'components/nifbadge'
import { ContaineeBadge } from 'components/containeebadge'


const AliasName = styled('span')(({ theme }) => ({
    fontStyle: 'italic',
}))

/**
 * Returns a description of a related network interface, consisting of the
 * interface name as well as its containee (badge). In case of multiple
 * containees, only the lexicographically first will be used (the only
 * exception to the rule being the initial network namespace containee
 * initwhatever).
 * 
 * @param nif network interface object
 */
const relatedItem = (nif: NetworkInterface) => {
    const containee = firstContainee(nif.netns)
    var nifname = nif.name
    var alias = ""
    // Do we want to show an alias or a container engine network name instead of
    // some incomprehensible interface name?
    if (nif.master && nif.master.kind === 'bridge') {
        nifname = ""
        alias = nif.master.alias ? nif.master.alias : nif.master.name
    } else {
        alias = nif.alias
    }

    return <>
        {nifname}{alias
            ? (nifname
                ? <> (~<AliasName>{alias}</AliasName>)</>
                : <> ~<AliasName>{alias}</AliasName></>)
            : ''
        }
        {containee && (<>&nbsp;<ContaineeBadge containee={containee} /></>)}
    </>
}

export interface NifNavigatorProps {
    /** network interface object describing a network interface in detail. */
    nif: NetworkInterface
    /** optionally show a network interface capture button? */
    capture?: boolean
    /** put an ID to this badge in any case (for placing the wiring). */
    anchor?: boolean
    /** 
     * if true, stretch the network interface badge to fill available
     * horizontal space. 
     */
    stretch?: boolean
    /** right align button contents when stretched. */
    alignRight?: boolean
    /** 
     * callback triggering when the user chooses to navigate to a particular
     * related network interface.
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
    /** CSS properties, for instance, to allow grid placement. */
    style?: React.CSSProperties
}

/**
 * A `NifNavigator` is a network interface (button) badge that can be clicked
 * and then shows a pop-up menu with options to navigate to related network
 * interfaces (VETH peers, MACVLAN master, MACVLANs, et cetera).
 *
 * The pop-up menu list of related network interfaces is sorted first by
 * related network interface name, and only then by the names of the
 * containees of the network namespaces where the interfaces belong to. Only
 * the lexicographically first containee name is shown and used for sorting.
 * In case of the initial network namespace the name of the initial containee
 * (initsomething) is used.
 *
 * This component basically wraps and supplies a `NifBadge` component with a
 * pop-up menu, filled with information about the related network interfaces.
 */
export const NifNavigator = ({
    nif,
    capture,
    anchor,
    stretch,
    alignRight,
    onNavigation,
    families,
    className,
    style
}: NifNavigatorProps) => {

    const [anchorEl, setAnchorEl] = React.useState<null | HTMLElement>(null)

    // Gather the related network interfaces, as we want to later populate the
    // pop-up menu with their names and containees.
    let relatedNifs = []
    if (nif.pf) {
        relatedNifs.push(nif.pf)
    }
    if (nif.macvlan) {
        relatedNifs.push(nif.macvlan)
    }
    if (nif.macvlans) {
        relatedNifs.push(...nif.macvlans)
    }
    if (nif.peer) {
        relatedNifs.push(nif.peer)
    }
    if (nif.overlays) {
        relatedNifs.push(...nif.overlays)
    }
    if (nif.underlay) {
        relatedNifs.push(nif.underlay)
    }
    if (nif.slaves) {
        relatedNifs.push(...nif.slaves)
    }

    const handleBadgeClick = (event: React.MouseEvent<HTMLElement>) => {
        setAnchorEl(event.currentTarget)
    }

    const handleMenuClose = () => {
        setAnchorEl(null)
    }

    // Bubble up a navigation event where the user clicks on a network
    // interface button badge and selects a related network interface.
    const handleNavigationClick = (nif: NetworkInterface) => {
        setAnchorEl(null)
        if (onNavigation) {
            onNavigation(nif)
        }
    }

    return (<>
        <NifBadge
            nif={nif}
            capture={capture}
            anchor={anchor}
            families={families}
            button={relatedNifs.length > 0}
            stretch={stretch}
            alignRight={alignRight}
            onClick={handleBadgeClick}
            className={className}
            style={style}
        />
        {/* Only render the related network interface popup menu if actually there some */}
        {relatedNifs.length > 0 &&
            <Menu
                anchorEl={anchorEl}
                keepMounted
                open={anchorEl !== null}
                onClose={handleMenuClose}
            >
                {
                    relatedNifs
                        .sort(orderNifByNameAndContainee)
                        .map(relnif =>
                            <MenuItem
                                key={nifId(relnif)}
                                onClick={() => handleNavigationClick(relnif)}
                            >
                                {relatedItem(relnif)}
                            </MenuItem>
                        )
                }
            </Menu>
        }
    </>)
}
