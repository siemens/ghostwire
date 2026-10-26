// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { nifHw } from 'models/gw/mock'

import { NifCheckbox } from './NifCheckbox'

const meta: Meta<typeof NifCheckbox> = {
    title: 'Ghostwire/NifCheckbox',
    component: NifCheckbox,
    args: {
        nif: nifHw,
        onChange: fn(),
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifCheckbox>

export const Unchecked: Story = {}

export const Checked: Story = {
    args: {
        checked: true,
    },
}
