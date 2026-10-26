// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useEffect } from 'react'

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Provider, useSetAtom } from 'jotai'

import { Box, Typography } from '@mui/material'

import { discoveryRefreshingAtom } from 'components/discovery'

import { Ghost } from './Ghost'

const StartRefresh = () => {
    const setRefreshing = useSetAtom(discoveryRefreshingAtom)
    useEffect(() => setRefreshing(true), [setRefreshing])
    return null
}

const meta: Meta<typeof Ghost> = {
    title: 'Ghostwire/Ghost',
    component: Ghost,
    args: {
        children: <Typography>Nothing to see here, no refresh going on.</Typography>,
    },
    argTypes: {
        children: { control: false },
    },
    decorators: [
        (Story) => <Box sx={{ height: '30ex' }}><Provider><Story /></Provider></Box>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Ghost>

export const Idle: Story = {}

// Keeps the discovery refresh going forever by never answering the fetch.
export const Refreshing: Story = {
    beforeEach: () => {
        const origFetch = globalThis.fetch
        globalThis.fetch = () => new Promise<Response>(() => { })
        return () => { globalThis.fetch = origFetch }
    },
    render: (args) => <>
        <StartRefresh />
        <Ghost {...args} />
    </>,
}
