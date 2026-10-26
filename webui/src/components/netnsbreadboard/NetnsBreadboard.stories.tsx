// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily, orderNetnsByContainees } from 'models/gw'
import { discovery, loOnlyContainerNetns } from 'models/gw/mock'

import { NetnsBreadboard } from './NetnsBreadboard'

// Only some network namespaces, so that some wires lead to network interfaces
// that aren't rendered at all.
const somenetnses = Object.values(discovery.networkNamespaces)
    .sort(orderNetnsByContainees)
    .filter(netns => Object.values(netns.nifs).length === 1
        || ['init', 'ghostwire'].some(prefix => netns.containers[0].name.startsWith(prefix)))
    .concat(loOnlyContainerNetns)

const meta: Meta<typeof NetnsBreadboard> = {
    title: 'Ghostwire/NetnsBreadboard',
    component: NetnsBreadboard,
    args: {
        netns: {
            ...discovery.networkNamespaces,
            [loOnlyContainerNetns.netnsid]: loOnlyContainerNetns,
        },
        // mirror the app's default of hiding lo.
        filterLo: true,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NetnsBreadboard>

export const Everything: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const WithLo: Story = {
    args: {
        filterLo: false,
    },
}

/**
 * Hides lo interfaces as well as non-container network namespaces with only lo;
 * the "network:none" container stays visible including its lo.
 */
export const Filtered: Story = {
    args: {
        netns: somenetnses,
        filterEmpty: true,
    },
}
