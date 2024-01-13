// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { styled, Tooltip } from '@mui/material'

import { AddressFamily, IpAddress, IpRoute, NetworkInterface, RouteTableLocal } from 'models/gw'
import { Address } from 'components/address'
import { NifBadge } from 'components/nifbadge'
import RouteVia from 'icons/routes/RouteVia'
import DefaultRoute from 'icons/routes/DefaultRoute'
import SomeRoute from 'icons/routes/Route'
import DirectSubnetRoute from 'icons/routes/DirectSubnetRoute'
import HostRoute from 'icons/routes/HostRoute'
import NoRoute from 'icons/routes/NoRoute'
import { rgba } from 'utils/rgba'
import { Campaign as MultiBroadcastRoute } from '@mui/icons-material'
import { useAtom } from 'jotai'
import { showMultiBroadcastRoutesAtom } from 'views/settings'


const RouteInfo = styled('div')(({ theme }) => ({
    '& .icon': {
        verticalAlign: 'middle',
        color: `${rgba(theme.palette.text.primary, 0.1)}`,
        marginRight: '0.1em',
    },
}))

const DestAddress = styled(Address)(() => ({
    marginLeft: '0.5em',
    marginRight: '0.5em',
}))

const NextHop = styled(Address)(() => ({
    marginLeft: '0.5em',
}))

const EgressNif = styled(NifBadge)(() => ({
    marginLeft: '0.5em',
}))

const Metric = styled('span')(() => ({
    marginLeft: '0.5em',
}))


export interface RouteProps {
    /** route object describing the details of a single network route. */
    route: IpRoute
    /** 
     * callback triggering when the user chooses to navigate to a particular
     * outgoing network interface of this route by clicking on the outgoing
     * nif badge.
     */
    onNavigation?: (nif: NetworkInterface) => void
    /** CSS class name(s) */
    className?: string
}

/**
 * The `Route` component renders a **single** route, execept for routes in the
 * so-called "local" routing table with ID 255 which won't be rendered at all.
 * The outgoing network interface (which can be unspecified!) is rendered as a
 * reference/link.
 */
export const Route = ({ route, onNavigation, className }: RouteProps) => {

    const [showMultiBroadcastRoutes] = useAtom(showMultiBroadcastRoutesAtom)

    // Determine the route type icon based on especially the prefix length of
    // the route, that is, how many destination(s) are covered. Also determine
    // the tooltip title which reflects the type of route in text instead of a
    // graphical depiction, effectively complementing it.
    let tooltip = ''
    let RouteIcon
    if (route.type === 'multicast' || route.type === 'broadcast') {
        tooltip = route.type + ' route'
        RouteIcon = MultiBroadcastRoute
    } else if (!route.nif) {
        tooltip = 'nirvana route'
        RouteIcon = NoRoute
    } else if (route.prefixlen === 0) {
        tooltip = 'default route'
        RouteIcon = DefaultRoute
    } else if (route.prefixlen === (route.family === AddressFamily.IPv6 ? 128 : 32)) {
        tooltip = 'host route'
        RouteIcon = HostRoute
    } else if (!route.nexthop) {
        tooltip = 'direct subnet route'
        RouteIcon = DirectSubnetRoute
    } else {
        tooltip = 'route'
        RouteIcon = SomeRoute
    }

    // Handle clicks on the outgoing network interface, if there is any; we
    // don't need to check as the callback cannot fire without an outgoing
    // network interface button badge component.
    const handleNifClick = () => {
        if (onNavigation) {
            onNavigation(route.nif)
        }
    }

    // Render the route (table) entry, unless it comes from Linux' local
    // table, which we don't show as it is noisy and clobbers the display.
    return (((showMultiBroadcastRoutes || route.table !== RouteTableLocal) &&
        <RouteInfo className={className}>
            <Tooltip title={tooltip}>
                <span>
                    {/* type of route icon and the route destination */}
                    {RouteIcon && <RouteIcon className="icon" />}
                    <DestAddress
                        route
                        notooltip
                        address={{
                            address: route.destination,
                            family: route.family,
                            prefixlen: route.prefixlen,
                        } as IpAddress} />
                </span>
            </Tooltip>

            {/* the next hop to take, if any. */}
            <RouteVia className="icon" />
            {route.nexthop &&
                <Tooltip title="next hop">
                    <NextHop
                        plain
                        notooltip
                        address={{
                            address: route.nexthop,
                            family: route.family,
                        } as IpAddress}
                    />
                </Tooltip>
            }

            {/* outgoing network interface, if any. */}
            {route.nif &&
                <EgressNif
                    nif={route.nif}
                    button
                    notooltip
                    onClick={handleNifClick}
                />
            }
            <Metric>metric {route.priority}</Metric>
        </RouteInfo>
    ) || <></>)
}
