// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { AddressFamily, loFiNetns } from 'models/gw'
import { crowdedNetns, initialNetns } from 'models/gw/mock'

import { NetnsPlainCard } from './NetnsPlainCard'

const meta: Meta<typeof NetnsPlainCard> = {
    title: 'Ghostwire/NetnsPlainCard',
    component: NetnsPlainCard,
    args: {
        netns: loFiNetns(initialNetns, true),
        onNavigation: fn(),
        onNetnsZoom: fn(),
        onContaineeZoom: fn(),
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NetnsPlainCard>

export const InitialNamespace: Story = {}

export const Crowded: Story = {
    args: {
        netns: loFiNetns(crowdedNetns!, true),
    },
}

export const WithLo: Story = {
    args: {
        netns: initialNetns,
    },
}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}
