// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useEffect } from 'react'

import type { Meta, StoryObj } from '@storybook/react-vite'

import { Provider, useSetAtom } from 'jotai'

import { discoveryRefreshingAtom } from 'components/discovery'
import mockdata from 'models/gw/mock/mockdata.json'

import Metadata from './Metadata'

const discoveryWithMetadata = {
    ...mockdata,
    metadata: {
        ...mockdata.metadata,
        'osrel-name': 'Debian GNU/Linux',
        'osrel-version': '13 (trixie)',
        'kernel-version': '6.12.48',
        'cpus': [[0, 3], [6, 7]],
        'container-engines': {
            '1': { type: 'docker.com', version: '28.5.1', pid: 1234 },
            '2': { type: 'containerd.io', version: '2.1.4', pid: 1000 },
        },
        'industrial-edge': {
            'device-name': 'ied-mockbox',
            'semversion': '1.24.0-4',
            'developer-mode': 'true',
        },
    },
}

const StartRefresh = () => {
    const setRefreshing = useSetAtom(discoveryRefreshingAtom)
    useEffect(() => setRefreshing(true), [setRefreshing])
    return null
}

const meta: Meta<typeof Metadata> = {
    title: 'Ghostwire/Metadata',
    component: Metadata,
    // Metadata only gets its data from a discovery, so feed it a mock one.
    beforeEach: () => {
        const origFetch = globalThis.fetch
        globalThis.fetch = async () => Response.json(discoveryWithMetadata)
        return () => { globalThis.fetch = origFetch }
    },
    decorators: [
        (Story) => <Provider>
            <StartRefresh />
            <Story />
        </Provider>,
    ],
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Metadata>

/** Click the expand button to show the discovery metadata. */
export const Standard: Story = {}
