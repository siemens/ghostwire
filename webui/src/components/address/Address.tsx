// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { forwardRef } from 'react'
import clsx from 'clsx'

import { AddressFamily, IpAddress } from 'models/gw'
import { IpAddressAndPrefix, IpAddressLifetime } from './IpAddressAndPrefix'
import { TooltipWrapper } from 'utils/tooltipwrapper'
import MacAddress from 'icons/addresses/MacAddress'
import Ipv4Address from 'icons/addresses/Ipv4Address'
import Ipv6Address from 'icons/addresses/Ipv6Address'
import { rgba } from 'utils/rgba'
import { styled } from '@mui/material'


const AddressContainer = styled('span')(({ theme }) => ({
    '& .MuiSvgIcon-root': {
        // tell the layout algo to hammer the icon to the text ceiling and
        // then position it relatively slightly more upwards so it appears
        // to be roughly centered to the text height...
        verticalAlign: 'text-top',
        position: 'relative',
        top: '-0.15ex',
        color: rgba(theme.palette.text.primary, 0.1),
        marginRight: '0.1em',
    },
}))

const Lifetimes = styled('span')(({ theme }) => ({
    color: theme.palette.text.secondary,
}))

const Lifetime = styled('span')(({ theme }) => ({
    whiteSpace: 'nowrap',
}))

const Addr = styled('span')(({ theme }) => ({
    fontFamily: 'Roboto Mono',
}))

const Prefix = styled('span')(({ theme }) => ({
    color: theme.palette.address.prefix,
    textDecoration: 'underline dotted'
}))

const IID = styled('span')(({ theme }) => ({
    color: theme.palette.address.iid,
    '& .route': { color: theme.palette.text.disabled }
}))

const PrefixLen = styled('span')(({ theme }) => ({
    color: theme.palette.address.prefix,
    marginLeft: '0.2em',
}))

const MAC = styled('span')(({ theme }) => ({
    color: theme.palette.text.secondary,
}))


// Map address families to their address family icons.
const addressFamilyIcons = {
    [AddressFamily.MAC]: MacAddress,
    [AddressFamily.IPv4]: Ipv4Address,
    [AddressFamily.IPv6]: Ipv6Address,
}

export interface AddressProps {
    /** 
     * a network address object, either for the IPv4, IPv6, or MAC address
     * family. 
     */
    address: IpAddress
    /** 
     * indicates that the address is a route destination; this also implies
     * notooltip.
     */
    route?: boolean
    /** 
     * do not visually emphasize prefix and interface identifier ("IID")
     * address parts. 
     */
    plain?: boolean
    /** 
     * shows address family icon; normally implied, except for `plain={true}`.
     */
    familyicon?: boolean
    /** CSS class name(s). */
    className?: string
    /** 
     * disable tooltip for address (there's never any tooltip for lifetimes).
     */
    notooltip?: boolean
    /** don't show an address family icon. */
    nofamilyicon?: boolean
}

/**
 * Renders a network address (of address family IPv4, IPv6, and MAC) in a
 * textual representation. In particular, for IP addresses the `Address`
 * component visually separates the prefix and host ("interface identifier")
 * parts, as well as a trailing prefix length indication. Also, for IP
 * addresses the preferred and valid lifetimes are shown (if defined).
 *
 * For IPv6 addresses, `Address` automatically performs "zero compression" of
 * zero-value groups using the standardized "::" notation ([RFC
 * 5952](https://tools.ietf.org/html/rfc5952)). Please note that it prefers
 * compressing the prefix part, where possible, over the IID (=interface
 * identifier, ~host) part.
 *
 * The lifetimes do not only apply to IPv6 addresses but also to IPv4
 * addresses. They are set for IPv4 addresses, for instance, by DHCPv4
 * clients.
 *
 * Address lifetimes are formatted as "days hh:mm:ss", with at most seconds
 * resolution.
 */
export const Address = forwardRef<HTMLSpanElement, AddressProps>((props, ref) => {

    const {
        address, route, plain, familyicon, className, notooltip, nofamilyicon,
        ...andnowforsomethingcompletelydifferentprops
    } = props

    // As for the seemingly strange component definition, please see:
    //
    // https://www.selbekk.io/blog/2020/05/forwarding-refs-in-typescript/. In
    // short: as we need to use this component in contexts where a "ref"
    // reference to the ("root") DOM element of the Address' component is
    // required, such as inside a <Tooltip> component. For this to work
    // correctly, we need to implement React's forwardRef API, as the "ref"
    // property isn't an ordinary property as all(most) the others.
    //
    // Additionally, we need to spread additional properties passed into us by
    // a Material-UI <Tooltip> component, see also:
    // https://github.com/mui-org/material-ui/issues/20653#issuecomment-616715804.

    const AddressIcon = ((!nofamilyicon || familyicon) && !route && addressFamilyIcons[address.family]) || undefined
    const isMAC = address.family === AddressFamily.MAC
    const [prefix, iid,] = isMAC ? [] : IpAddressAndPrefix(address.address, address.prefixlen)
    const tooltip = <>{((isMAC && 'MAC address') ||
        (address.family === AddressFamily.IPv6 && 'IPv6 address') ||
        'IPv4 address')}</>

    const addressOnly =
        <Addr>
            {(plain && <>
                {familyicon && <AddressIcon />}{address.address}
            </>) || (!isMAC && <>
                {AddressIcon && <AddressIcon />}
                <Prefix>{prefix}</Prefix>
                <IID className={clsx(route && 'route')}>{iid}</IID>
                <PrefixLen>/{address.prefixlen}</PrefixLen>
            </>) || (address.address !== '00:00:00:00:00:00' &&
                <>
                    <AddressIcon />
                    <MAC>{address.address}</MAC>
                </>
                )}
        </Addr>

    return (
        <AddressContainer
            ref={ref}
            className={className}
            {...andnowforsomethingcompletelydifferentprops}
        >
            {TooltipWrapper(addressOnly, tooltip, !notooltip)}
            {!route && !isMAC && address.preferredLifetime && address.validLifetime &&
                <Lifetimes>
                    {' · '}
                    <Lifetime>preferred {IpAddressLifetime(address.preferredLifetime)}</Lifetime>
                    {' · '}
                    <Lifetime>valid {IpAddressLifetime(address.validLifetime)}</Lifetime>
                </Lifetimes>
            }
        </AddressContainer>
    )
})
