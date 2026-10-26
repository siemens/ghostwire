// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { crowdedNetns, initialNetns } from 'models/gw/mock'

import { NamespaceContainees } from './NamespaceContainees'

const meta: Meta<typeof NamespaceContainees> = {
    title: 'Ghostwire/NamespaceContainees',
    component: NamespaceContainees,
    args: {
        netns: crowdedNetns!,
        onNavigation: fn(),
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NamespaceContainees>

export const Crowded: Story = {}

/** The list is cut short and an ellipsis indicates left out containees. */
export const Limited: Story = {
    args: {
        limit: 2,
    },
}

export const InitialNamespace: Story = {
    args: {
        netns: initialNetns,
    },
}
