// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box, Paper, Typography } from '@mui/material'

import { AddressFamily, orderAddresses } from '../../models/gw'
import { mockAddresses, nifHw, nifLo } from '../../models/gw/mock'

import { Address } from './Address'

const loIPv4 = nifLo.addresses.find(addr => addr.family === AddressFamily.IPv4)!
const loIPv6 = nifLo.addresses.find(addr => addr.family === AddressFamily.IPv6)!
const loMAC = nifLo.addresses.find(addr => addr.family === AddressFamily.MAC)!
const hwMAC = nifHw.addresses.find(addr => addr.family === AddressFamily.MAC)!

const meta: Meta<typeof Address> = {
    title: 'Ghostwire/Address',
    component: Address,
    argTypes: {
        address: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Address>

export const IPv4Loopback: Story = {
    name: 'IPv4 Loopback',
    args: {
        address: loIPv4,
    },
}

export const IPv6Loopback: Story = {
    name: 'IPv6 Loopback',
    args: {
        address: loIPv6,
    },
}

export const MAC: Story = {
    args: {
        address: hwMAC,
    },
}

/** Loopback interfaces have an all-zero MAC address, which isn't rendered. */
export const AllZeroMAC: Story = {
    args: {
        address: loMAC,
    },
}

/**
 * In some contexts the address family icon will clutter and distract, so it
 * can be switched off using the `nofamilyicon` property. Hopefully, IKEA
 * customer relationship management won't object...
 */
export const NoFamilyIcon: Story = {
    args: {
        address: loIPv4,
        nofamilyicon: true,
    },
}

/**
 * The tooltip can optionally be disabled; this is used when displaying network
 * addresses as part of routes, where inner tooltips would be really vexing and
 * annoying.
 */
export const NoTooltip: Story = {
    args: {
        address: loIPv4,
        notooltip: true,
    },
}

/**
 * Various addresses, IPv4, IPv6, and even some MAC-layer addresses; please
 * note that this includes an all-zero MAC address, which isn't rendered at all
 * (such all-zero MACs pop up in the context of loopback interfaces).
 */
export const AllAddresses: Story = {
    render: () => <>
        {[...mockAddresses].sort(orderAddresses).map((addr, idx) => (
            <Box key={idx} sx={{ py: 1 }}>
                <Typography variant="body1" color="textSecondary">
                    {addr.address}/{addr.prefixlen}
                </Typography>
                <Paper variant="outlined" sx={{ p: 1 }}>
                    <Address address={addr} />
                </Paper>
            </Box>
        ))}
    </>,
}
