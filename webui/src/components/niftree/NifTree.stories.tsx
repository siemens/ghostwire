// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { NifInfoModalProvider } from 'components/nifinfomodal'
import { AddressFamily } from 'models/gw'
import { crowdedNetns, initialNetns, loOnlyContainerNetns } from 'models/gw/mock'

import { NifTree } from './NifTree'

const meta: Meta<typeof NifTree> = {
    title: 'Ghostwire/NifTree',
    component: NifTree,
    args: {
        netns: initialNetns,
        onNavigation: fn(),
        onContaineeNavigation: fn(),
    },
    argTypes: {
        netns: { control: false },
    },
    decorators: [
        (Story) => <NifInfoModalProvider><Story /></NifInfoModalProvider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifTree>

export const InitialNamespace: Story = {}

export const Filtered: Story = {
    args: {
        filterLo: true,
        filterMAC: true,
        families: [AddressFamily.IPv4],
    },
}

export const Container: Story = {
    args: {
        netns: crowdedNetns!,
    },
}

// "network:none" containers keep their lo even when filtering lo.
export const LoOnlyContainer: Story = {
    args: {
        netns: loOnlyContainerNetns,
        filterLo: true,
    },
}
