// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { type Container, isContainer } from 'models/gw'
import { discovery } from 'models/gw/mock'
import { type Service } from 'utils/neighborhood'

import { Neighborhood } from './Neighborhood'

const composeService = 'com.docker.compose.service'

const containers = Object.values(discovery.networkNamespaces)
    .flatMap(netns => netns.containers)
    .filter((cntr): cntr is Container => isContainer(cntr) && !!cntr.labels[composeService])

const services: Service[] = Object.values(containers.reduce((services, cntr) => {
    const name = cntr.labels[composeService]
    services[name] ??= { name, containers: [], networks: ['mock_default'] }
    services[name].containers.push(cntr)
    return services
}, {} as { [name: string]: Service }))

// a stand-alone container without any service is its own name-less service.
services.push({
    name: '',
    containers: [{ ...containers[0], name: 'lonely_moby', labels: {} }],
    networks: ['mock_default'],
})

const meta: Meta<typeof Neighborhood> = {
    title: 'Ghostwire/Neighborhood',
    component: Neighborhood,
    args: {
        services: services,
        seenby: [containers[0]],
    },
    argTypes: {
        services: { control: false },
        seenby: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Neighborhood>

// "Check" needs a Ghostwire service's mobydig websocket, so it fails here.
export const Standard: Story = {}
