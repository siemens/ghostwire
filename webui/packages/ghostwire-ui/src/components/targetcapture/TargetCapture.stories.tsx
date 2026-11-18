// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Paper } from '@mui/material'

import { OperationalState } from '../../models/gw'
import { initialNetns, nifHw } from '../../models/gw/mock'

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

/** Small-sized and medium-sized target capture buttons. */
export const Sizes: Story = {
    render: (args) => <>
        <Paper variant="outlined" sx={{ p: 1, mb: 1 }}>
            <TargetCapture {...args} size="small" />
        </Paper>
        <Paper variant="outlined" sx={{ p: 1 }}>
            <TargetCapture {...args} size="medium" />
        </Paper>
    </>,
}
