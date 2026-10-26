// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box } from '@mui/material'

import { nifBr, nifHw, nifLo, nifVeth1 } from 'models/gw/mock'

import { NifInfo } from './NifInfo'

const meta: Meta<typeof NifInfo> = {
    title: 'Ghostwire/NifInfo',
    component: NifInfo,
    args: {
        nif: nifHw,
    },
    argTypes: {
        nif: { control: false },
    },
    decorators: [
        (Story) => <Box sx={{ width: '30em' }}><Story /></Box>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifInfo>

export const HardwareNic: Story = {}

export const Veth: Story = {
    args: {
        nif: nifVeth1,
    },
}

// A firmware version of "N/A" isn't shown.
export const Bridge: Story = {
    args: {
        nif: nifBr,
    },
}

// Missing information is rendered as crossed out.
export const Loopback: Story = {
    args: {
        nif: nifLo,
    },
}
