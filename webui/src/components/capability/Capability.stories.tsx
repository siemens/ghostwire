// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Capability } from './Capability'

const meta: Meta<typeof Capability> = {
    title: 'Ghostwire/Capability',
    component: Capability,
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Capability>

export const DockerDefault: Story = {
    args: {
        name: 'CAP_DAC_OVERRIDE',
    },
}

export const NonDefault: Story = {
    args: {
        name: 'CAP_SYS_ADMIN',
    },
}

export const Plain: Story = {
    args: {
        name: 'CAP_SYS_ADMIN',
        plain: true,
    },
}
