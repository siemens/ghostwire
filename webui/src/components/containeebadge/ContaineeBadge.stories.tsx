// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box } from '@mui/material'

import { DynVarsProvider } from 'components/dynvars'
import { type Containee, ContainerFlavors, ContainerState } from 'models/gw'
import { boundBox, containerBox, pod, standaloneBox } from 'views/help/mockdata'

import { ContaineeBadge } from './ContaineeBadge'

const types: Containee[] = [
    boundBox,
    standaloneBox,
    containerBox,
    { ...containerBox, name: 'containerd_moby', flavor: ContainerFlavors.CONTAINERD, engineType: 'containerd' },
    { ...containerBox, name: 'podman_moby', flavor: ContainerFlavors.PODMAN, engineType: 'podman' },
    { ...containerBox, name: 'kind-control-plane', flavor: ContainerFlavors.KIND },
    { ...containerBox, name: 'managed_plugin', flavor: ContainerFlavors.DOCKERPLUGIN },
    { ...containerBox, name: 'edge-iot-core', flavor: ContainerFlavors.IERUNTIME },
    { ...containerBox, name: 'ieapp', flavor: ContainerFlavors.IEAPP },
    pod,
]

const dockerStates: Containee[] = Object.entries(ContainerState)
    .map(([name, state]) => ({ ...containerBox, name: `${name}_Moby`, state }))

const badges = (containees: Containee[]) => <>
    {containees.map(containee => (
        <Box sx={{ py: 1 }} key={containee.name}>
            <ContaineeBadge containee={containee} />
        </Box>
    ))}
</>

const meta: Meta<typeof ContaineeBadge> = {
    title: 'Ghostwire/ContaineeBadge',
    component: ContaineeBadge,
    args: {
        containee: containerBox,
    },
    argTypes: {
        containee: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ContaineeBadge>

export const Standard: Story = {}

// The tooltips reflect the type of containee.
export const Types: Story = {
    render: () => badges(types),
}

// The tooltips reflect the container state.
export const DockerStates: Story = {
    render: () => badges(dockerStates),
}

export const Button: Story = {
    args: {
        button: true,
        onClick: fn(),
    },
}

export const Angled: Story = {
    args: {
        angled: true,
    },
}

export const Capture: Story = {
    args: {
        capture: true,
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}

export const NoTooltip: Story = {
    args: {
        notooltip: true,
    },
}
