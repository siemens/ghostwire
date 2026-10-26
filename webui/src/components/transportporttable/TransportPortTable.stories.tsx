// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { discovery, initialNetns } from 'models/gw/mock'

import { TransportPortTable } from './TransportPortTable'

const portsNetns = Object.values(discovery.networkNamespaces)
    .reduce((most, netns) => netns.transportPorts.length > most.transportPorts.length ? netns : most)

const meta: Meta<typeof TransportPortTable> = {
    title: 'Ghostwire/TransportPortTable',
    component: TransportPortTable,
    args: {
        netns: portsNetns,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof TransportPortTable>

export const Standard: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const InitialNamespace: Story = {
    args: {
        netns: initialNetns,
    },
}
