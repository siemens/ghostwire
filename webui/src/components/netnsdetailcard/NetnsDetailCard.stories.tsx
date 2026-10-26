// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { crowdedNetns, initialNetns, loOnlyContainerNetns } from 'models/gw/mock'

import { NetnsDetailCard } from './NetnsDetailCard'

const meta: Meta<typeof NetnsDetailCard> = {
    title: 'Ghostwire/NetnsDetailCard',
    component: NetnsDetailCard,
    args: {
        netns: initialNetns,
        filterLo: true,
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

export const WithLo: Story = {
    args: {
        filterLo: false,
    },
}

export const Filtered: Story = {
    args: {
        filterMAC: true,
        families: [AddressFamily.IPv4],
    },
}

/** "network:none" containers keep their lo even when filtering lo. */
export const LoOnlyContainer: Story = {
    args: {
        netns: loOnlyContainerNetns,
    },
}

/** The maximize and minimize buttons only work inside a view route. */
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
