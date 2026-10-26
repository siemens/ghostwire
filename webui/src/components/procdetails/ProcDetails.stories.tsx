// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { ProcessDetails } from './ProcDetails'

const meta: Meta<typeof ProcessDetails> = {
    title: 'Ghostwire/ProcessDetails',
    component: ProcessDetails,
    args: {
        cmdline: ['/usr/bin/fool', '--world-domination', '--now'],
        pid: 12345,
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ProcessDetails>

export const Standard: Story = {}
