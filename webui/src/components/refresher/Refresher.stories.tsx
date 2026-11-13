// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Provider, useAtomValue } from 'jotai'

import { Typography } from '@mui/material'

import { discoveryRefreshIntervalAtom } from 'components/discovery'

import Refresher from './Refresher'

const Interval = () => {
    const refreshInterval = useAtomValue(discoveryRefreshIntervalAtom)
    return <Typography>interval: {refreshInterval === null ? 'off' : `${refreshInterval}ms`}</Typography>
}

const meta: Meta<typeof Refresher> = {
    title: 'Ghostwire/Refresher',
    component: Refresher,
    decorators: [
        (Story) => <Provider>
            <Interval />
            <Story />
        </Provider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Refresher>

export const DefaultIntervals: Story = {}

export const CustomIntervals: Story = {
    args: {
        intervals: [
            { interval: null },
            { interval: 666, label: 'bad timing' },
            { interval: 1000 },
        ],
    },
}
