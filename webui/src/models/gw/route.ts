// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { AddressFamily } from './address'
import { NetworkInterface } from './nif';

export enum RouteType {
    Unicast = "unicast",
    Local = "local",
    Broadcast = "broadcast",
    Multicast = "multicast",
}

export interface IpRoute {
    destination: string
    prefixlen: number
    family: AddressFamily
    index: number
    nif: NetworkInterface
    nexthop?: string
    preference: string
    priority: number
    table: number
    type: RouteType
}

export const RouteTableLocal = 255

const routeString = (rt: IpRoute) =>
    `${rt.destination}/${rt.prefixlen.toString().padStart(3, '0')}+${rt.priority.toString().padStart(3, '0')}`

export const orderRoutes = (rt1: IpRoute, rt2: IpRoute) => {
    if (rt1.family !== rt2.family) {
        return rt1.family - rt2.family
    }
    return routeString(rt1).localeCompare(routeString(rt2))
}

export const routeKey = (rt: IpRoute) =>
    `${routeString(rt)}+${rt.nif && rt.nif.name}`
    