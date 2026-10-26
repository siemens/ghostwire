// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { initialNetns } from 'models/gw/mock'

import { NifAddressList } from './NifAddressList'

// nifHw has only a MAC address, so pick a NIC with MAC, IPv4 and IPv6 addresses.
const wlan = Object.values(initialNetns.nifs)
    .find(nif => nif.isPhysical && [AddressFamily.MAC, AddressFamily.IPv4, AddressFamily.IPv6]
        .every(family => nif.addresses.some(addr => addr.family === family)))!

// leave out the ULAs to reduce clutter.
const nif = { ...wlan, addresses: wlan.addresses.filter(addr => !addr.address.startsWith('fdc6:')) }

const meta: Meta<typeof NifAddressList> = {
    title: 'Ghostwire/NifAddressList',
    component: NifAddressList,
    args: {
        nif: nif,
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifAddressList>

/** The MAC address, followed by the IPv4 and then the IPv6 addresses. */
export const Standard: Story = {}

/** Only the IPv4 and IPv6 addresses, without the MAC address. */
export const NoMAC: Story = {
    args: {
        filterMAC: true,
    },
}

/** Only the IPv6 addresses, without the MAC and IPv4 addresses. */
export const IPv6Only: Story = {
    name: 'IPv6 Only',
    args: {
        filterMAC: true,
        families: [AddressFamily.IPv6],
    },
}
