// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { OperationalState } from 'models/gw'
import { initialNetns, nifHw } from 'models/gw/mock'

import { TargetCapture } from './TargetCapture'

const meta: Meta<typeof TargetCapture> = {
    title: 'Ghostwire/TargetCapture',
    component: TargetCapture,
    args: {
        target: initialNetns,
        demo: true,
    },
    argTypes: {
        target: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof TargetCapture>

export const NetworkNamespace: Story = {}

export const Small: Story = {
    args: {
        size: 'small',
    },
}

export const NetworkInterface: Story = {
    args: {
        target: nifHw,
    },
}

export const NonOperationalInterface: Story = {
    args: {
        target: { ...nifHw, operstate: OperationalState.Down },
    },
}

export const Disabled: Story = {
    args: {
        disabled: true,
    },
}
