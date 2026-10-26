// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { nifHw, nifMacvlan, nifVeth1, nifVeth2 } from 'models/gw/mock'

import { RelatedNif } from './RelatedNif'

const meta: Meta<typeof RelatedNif> = {
    title: 'Ghostwire/RelatedNif',
    component: RelatedNif,
    args: {
        nif: nifVeth1,
        onNavigation: fn(),
        onContaineeNavigation: fn(),
    },
    argTypes: {
        nif: { control: false },
        unrelatedNif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof RelatedNif>

// The related nif is in a different netns, so its containees are shown too.
export const Veth: Story = {}

export const VethPeer: Story = {
    args: {
        nif: nifVeth2!,
    },
}

export const Macvlan: Story = {
    args: {
        nif: nifMacvlan,
    },
}

// Renders nothing, as there are no related network interfaces.
export const Unrelated: Story = {
    args: {
        nif: { ...nifHw, macvlans: undefined },
    },
}
