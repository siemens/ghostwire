// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box } from '@mui/material'

import { DynVarsProvider } from 'components/dynvars'
import { orderNifByName } from 'models/gw'
import { discovery, nifHw, nifLo, nifVeth1, nifVeth2 } from 'models/gw/mock'

import { NifNavigator } from './NifNavigator'

// many "eth0" VETH interfaces, belonging to different containers.
const veths = Object.values(discovery.networkNamespaces)
    .flatMap(netns => Object.values(netns.nifs).filter(nif => nif.kind === 'veth'))
    .sort(orderNifByName)
    .slice(0, 3)

const meta: Meta<typeof NifNavigator> = {
    title: 'Ghostwire/NifNavigator',
    component: NifNavigator,
    args: {
        nif: nifVeth1,
        onNavigation: fn(),
        onChange: fn(),
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifNavigator>

// Click on the badge to pop up a menu with the related network interfaces.
export const Veth: Story = {}

export const VethPeer: Story = {
    args: {
        nif: nifVeth2!,
    },
}

export const MacvlanMaster: Story = {
    args: {
        nif: nifHw,
    },
}

// Nothing to navigate to.
export const Unrelated: Story = {
    args: {
        nif: nifLo,
    },
}

export const Capture: Story = {
    args: {
        capture: true,
        stretch: true,
        style: { width: '20em' },
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}

export const Checkable: Story = {
    args: {
        nifCheck: true,
        checked: true,
    },
}

export const SeveralVeths: Story = {
    render: (args) => <>
        {veths.map(nif => (
            <Box sx={{ py: 1 }} key={`${nif.netns.netnsid}-${nif.name}`}>
                <NifNavigator {...args} nif={nif} />
            </Box>
        ))}
    </>,
}
