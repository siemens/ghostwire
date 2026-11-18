// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { nifHw, nifMacvlan, nifVeth1, nifVeth2 } from '../../models/gw/mock'

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

/**
 * For network interfaces with related interfaces, (only) these related network
 * interfaces are rendered. If the related network interface is in a different
 * network namespace, then additional information about the related network
 * namespace is rendered in form of the entities (containers, stand-alone
 * processes) "inside" the related namespace.
 */
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

/**
 * **Nothing gets rendered** for network interfaces without any related
 * interfaces, such as hardware interfaces without attached macvlan interfaces,
 * the loopback interface, et cetera.
 */
export const Unrelated: Story = {
    args: {
        nif: { ...nifHw, macvlans: undefined },
    },
}
