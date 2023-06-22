// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React from 'react'

import { NetworkInterface } from 'models/gw'
import { AddressFamily, AddressFamilySet, orderAddresses } from 'models/gw/address'
import { Address } from 'components/address'

export interface NifAddressListProps {
    /** network interface object with address configuration. */
    nif: NetworkInterface
    /** if `true`, then hides MAC layer addresses. */
    filterMAC?: boolean
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
 * `NifAddressList` renders a sorted list of the network addresses configured
 * for a specific network interface. The list is first sorted by address
 * family MAC, IPv4, IPv6, and second by network addresses within the same
 * family.
 */
export const NifAddressList = ({ nif, filterMAC, families, className }: NifAddressListProps) => {
    const addrFamilies = [
        ...(filterMAC ? [] : [AddressFamily.MAC]),
        ...(families || [AddressFamily.IPv4, AddressFamily.IPv6])
    ]
    return (
        <div className={className}>
            {addrFamilies.map(family => nif.addresses
                .filter(addr => addr.family === family)
                .sort(orderAddresses)
                .map(addr => <div key={`${addr.address}/${addr.prefixlen}`}><Address address={addr} /></div>))
                .flat()
            }
        </div>
    )
}
