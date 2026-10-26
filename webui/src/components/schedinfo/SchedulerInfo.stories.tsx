// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { containerBoxProcess } from 'views/help/mockdata'

import SchedulerInfo from './SchedulerInfo'

const meta: Meta<typeof SchedulerInfo> = {
    title: 'Ghostwire/SchedulerInfo',
    component: SchedulerInfo,
    args: {
        process: containerBoxProcess,
    },
    argTypes: {
        process: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof SchedulerInfo>

export const Normal: Story = {
    args: {
        showNormal: true,
    },
}

export const Nice: Story = {
    args: {
        process: { ...containerBoxProcess, policy: 0, nice: 10 },
    },
}

export const NotNice: Story = {
    args: {
        process: { ...containerBoxProcess, policy: 0, nice: -5 },
    },
}

export const FIFO: Story = {
    args: {
        process: { ...containerBoxProcess, policy: 1, priority: 50 },
    },
}

export const RoundRobin: Story = {
    args: {
        process: { ...containerBoxProcess, policy: 2, priority: 0 },
    },
}

export const Idle: Story = {
    args: {
        process: { ...containerBoxProcess, policy: 5 },
    },
}
