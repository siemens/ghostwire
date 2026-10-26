// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Provider } from 'jotai'
import { SnackbarProvider } from 'notistack'

import { Stack, Typography } from '@mui/material'

import Refresher from 'components/refresher'
import mockdata from 'models/gw/mock/mockdata.json'

import { Discovery } from './Discovery'
import { useDiscovery } from './disco'

const Status = () => {
    const discovery = useDiscovery()
    return <Stack direction="row" spacing={2} sx={{ alignItems: 'center' }}>
        <Refresher />
        <Typography>
            {Object.keys(discovery.networkNamespaces).length} network namespaces discovered
        </Typography>
    </Stack>
}

// Discovery fetches from the global fetch, so swap it for the story's lifetime.
const stubFetch = (response: () => Response) => () => {
    const origFetch = globalThis.fetch
    globalThis.fetch = async () => response()
    return () => { globalThis.fetch = origFetch }
}

const meta: Meta<typeof Discovery> = {
    title: 'Ghostwire/Discovery',
    component: Discovery,
    decorators: [
        (Story) => <Provider>
            <SnackbarProvider>
                <Story />
                <Status />
            </SnackbarProvider>
        </Provider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Discovery>

export const Successful: Story = {
    // uses the stored refresh interval; if off, click the refresh button.
    beforeEach: stubFetch(() => Response.json(mockdata)),
}

// Not in the docs page, as its fetch stub would clash with the other story's.
export const Failing: Story = {
    tags: ['!autodocs'],
    beforeEach: stubFetch(() => new Response(null, { status: 503, statusText: 'Service Unavailable' })),
}
