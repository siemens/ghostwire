// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import * as ip6addr from 'ip6addr'


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
 * Indicates when an address is an unspecified IP address.
 * 
 * @param addr some address, including non-IP
 * @returns true when address is an unspecified IPv4 or IPv6 address; false
 *   otherwise.
 */
export const isUnspecifiedIP = (addr: IpAddress) =>
    (addr.prefixlen === 32 && addr.address === '0.0.0.0')
    || (addr.prefixlen === 128 && addr.address === '::')

/**
 * Indicates when an address is a loopback IP address.
 * 
 * @param addr some address, including non-IP
 * @returns true when address is a loopback IPv4 address in the range
 *   127.0.0.0/8 or the ::1 IPv6 address; false otherwise.
 */
export const isLoopbackIP = (addr: IpAddress) =>
    (addr.prefixlen >= 8 && addr.address.startsWith('127.'))
    || (addr.prefixlen === 128 && addr.address === '::1')

/**
 * 
 */
export const isLLAv6 = (addr: IpAddress) =>
    (addr.family === AddressFamily.IPv6 && addr.prefixlen >= 16 && addr.address.startsWith('fe80:'))

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
    // We could have used ip6addr.compareCIDR, but then we would need to
    // interpolate the "addr/prefix" strings, so we can more easily do the
    // comparism for the prefix lengths separately and only when we really need
    // it.
    const order = ip6addr.compare(addr1.address, addr2.address)
    return order || (addr1.prefixlen - addr2.prefixlen)
}
