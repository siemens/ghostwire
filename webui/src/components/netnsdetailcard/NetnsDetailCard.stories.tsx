// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { crowdedNetns, initialNetns } from 'models/gw/mock'

import { NetnsDetailCard } from './NetnsDetailCard'

const meta: Meta<typeof NetnsDetailCard> = {
    title: 'Ghostwire/NetnsDetailCard',
    component: NetnsDetailCard,
    args: {
        netns: initialNetns,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NetnsDetailCard>

export const InitialNamespace: Story = {}

export const Crowded: Story = {
    args: {
        netns: crowdedNetns!,
    },
}

export const Filtered: Story = {
    args: {
        filterLo: true,
        filterMAC: true,
        families: [AddressFamily.IPv4],
    },
}

// The maximize and minimize buttons only work inside a view route.
export const MaximizeMinimize: Story = {
    parameters: {
        routerProps: { initialEntries: ['/details'] },
    },
    args: {
        netns: crowdedNetns!,
        canMaximize: true,
        canMinimize: true,
    },
}
