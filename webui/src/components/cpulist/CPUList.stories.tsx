// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Box } from '@mui/material'

import { CPUList } from './CPUList'

const meta: Meta<typeof CPUList> = {
    title: 'Ghostwire/CPUList',
    component: CPUList,
    args: {
        cpus: [[0, 3], [5, 5], [8, 11], [16, 23], [42, 42]],
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof CPUList>

export const Standard: Story = {}

export const WithIcon: Story = {
    args: {
        showIcon: true,
        tooltip: 'CPUs online',
    },
}

export const NoWrap: Story = {
    args: {
        noWrap: true,
    },
    decorators: [
        (Story) => <Box sx={{ width: '4em' }}><Story /></Box>,
    ],
}
