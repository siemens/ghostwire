// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { parse } from 'ipaddr.js'


export enum AddressFamily {
    MAC = 0, /** pseudo family for "hw" network addresses */
    IPv4 = 2,
    IPv6 = 10
}

export type AddressFamilySet = AddressFamily[]

/**
 * Returns the AddressFamily enumeration value corresponding with either the
 * "ipv4" or "ipv6" string value.
 *
 * @param f either "ipv4" or "ipv6"
 */
export const addressFamilyByName = (f: string) => {
    if (f === 'ipv4') {
        return AddressFamily.IPv4
    } else if (f === 'ipv6') {
        return AddressFamily.IPv6
    }
    return AddressFamily.MAC // erm, well, anyway.
}

/**
 * Representation of a data-link layer or network layer address.
 */
export interface IpAddress {
    address: string
    prefixlen: number
    family: AddressFamily
    preferredLifetime: number
    validLifetime: number
    scope: number
}

/**
 * Sort order function for sorting IpAddress objects with the following
 * priority:
 *
 * - address family, with IPv4 coming before IPv6 (ouch!),
 * - address (within same IP family) -- the ordering bases on the correct IP
 *   address octet string representation, not on a textual comparism,
 * - prefix length (for same address and family)
 *
 * And yes, prefix length applies also to IPv4 (it already has for decades) and
 * it makes things so much easier compared to that totally borked "net mask"
 * concept, where you had to explicitly prevent users from using "255.255.0.31"
 * or such nonsense.
 *
 * @param addr1 first address object
 * @param addr2 second address object
 */
export const orderAddresses = (addr1: IpAddress, addr2: IpAddress) => {
    const no1Addr = !addr1
    const no2Addr = !addr2
    if (no1Addr || no2Addr) {
        return no1Addr === no2Addr ? 0 : no1Addr ? -1 : 1
    }
    if (addr1.family !== addr2.family) {
        return addr1.family - addr2.family
    }
    // If we're comparing MAC addresses, we can simply compare them as strings;
    // they have fixed length and format, so unless the discovery service feeds
    // us trash, this will work beautifully.
    if (addr1.family === AddressFamily.MAC) {
        return addr1.address.localeCompare(addr2.address)
    }
    const ip6addr1 = parse(addr1.address).toByteArray()
    const ip6addr2 = parse(addr2.address).toByteArray()
    let order: number = 0
    for (let idx = 0; idx < 15; idx++) {
        order = ip6addr1[idx] - ip6addr2[idx]
        if (order) {
            break
        }
    }
    return order || (addr1.prefixlen - addr2.prefixlen)
}
