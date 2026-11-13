// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { SRIOVRole } from 'models/gw'
import { nifHw } from 'models/gw/mock'

import { NifHWIcon } from './NifHWIcon'

const meta: Meta<typeof NifHWIcon> = {
    title: 'Ghostwire/NifHWIcon',
    component: NifHWIcon,
    args: {
        nif: nifHw,
        fontSize: 'large',
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifHWIcon>

export const PlainNIC: Story = {}

export const PhysicalFunction: Story = {
    args: {
        nif: { ...nifHw, sriovrole: SRIOVRole.PF },
    },
}

export const VirtualFunction: Story = {
    args: {
        nif: { ...nifHw, sriovrole: SRIOVRole.VF },
    },
}
