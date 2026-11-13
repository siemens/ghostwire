// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { discovery } from 'models/gw/mock'

import { VxlanDetails } from './VxlanDetails'

const vxlan = Object.values(discovery.networkNamespaces)
    .flatMap(netns => Object.values(netns.nifs))
    .find(nif => nif.kind === 'vxlan')!

const meta: Meta<typeof VxlanDetails> = {
    title: 'Ghostwire/VxlanDetails',
    component: VxlanDetails,
    args: {
        nif: vxlan,
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof VxlanDetails>

export const Standard: Story = {}

export const ArpProxy: Story = {
    args: {
        nif: {
            ...vxlan,
            vxlanDetails: { vid: 42, arpProxy: true, remotePort: 4789, sourcePortLow: 0, sourcePortHigh: 0 },
        },
    },
}
