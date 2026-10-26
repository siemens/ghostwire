// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily } from 'models/gw'
import { crowdedNetns } from 'models/gw/mock'

import { DetailedContainees } from './DetailedContainees'

const meta: Meta<typeof DetailedContainees> = {
    title: 'Ghostwire/DetailedContainees',
    component: DetailedContainees,
    args: {
        netns: crowdedNetns!,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof DetailedContainees>

export const Standard: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const ShowSandbox: Story = {
    args: {
        showSandbox: true,
    },
}
