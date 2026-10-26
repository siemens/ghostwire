// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { nifHw } from 'models/gw/mock'

import { NifAddressList } from './NifAddressList'

const meta: Meta<typeof NifAddressList> = {
    title: 'Ghostwire/NifAddressList',
    component: NifAddressList,
    args: {
        nif: nifHw,
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifAddressList>

export const Standard: Story = {}

export const NoMAC: Story = {
    args: {
        filterMAC: true,
    },
}

export const IPv6Only: Story = {
    args: {
        families: [AddressFamily.IPv6],
    },
}
