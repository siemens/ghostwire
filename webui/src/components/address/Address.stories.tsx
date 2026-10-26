// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { nifHw, nifLo } from 'models/gw/mock'

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

export const NoFamilyIcon: Story = {
    args: {
        address: loIPv4,
        nofamilyicon: true,
    },
}

export const NoTooltip: Story = {
    args: {
        address: loIPv4,
        notooltip: true,
    },
}
