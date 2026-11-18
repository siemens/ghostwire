// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box } from '@mui/material'

import { DynVarsProvider } from '../dynvars'
import { type Containee, ContainerFlavors, ContainerState } from '../../models/gw'
import { boundBox, containerBox, pod, standaloneBox } from '../../views/help/mockdata'

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

/**
 * These are examples of `ContaineeBadge`s for bind-mounted namespaces,
 * stand-alone processes attached to a namespace, and containers – also with
 * different flavors. Please notice how the tooltips reflect the type of
 * containee as well as the type of container.
 */
export const Types: Story = {
    render: () => badges(types),
}

/**
 * Container `ContaineeBadge`s in various states. Please notice how the
 * tooltips reflect the container status.
 */
export const DockerStates: Story = {
    render: () => badges(dockerStates),
}

/**
 * Additionally, a `ContaineeBadge` might act as a button referencing/linking to
 * the namespace UI element the containee is attached to. When acting as links,
 * `ContaineeBadge`s appear as flat UI buttons (including ripple effect, et
 * cetera), yet keep their status indication.
 *
 * Use the `button` property (as well as an `onClick` callback handler) to
 * enable this aspect on a `ContaineeBadge`.
 */
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

/**
 * A containee badge optionally can show an additional capture "fin" button,
 * rendered next to the containee badge. The capture button always appears to
 * the right of the containee badge; it does not obey text direction, as the
 * badge text direction is ltr due to using ASCII latin glyphs (even if
 * intermediate APIs or transport protocols might use Unicode).
 */
export const Capture: Story = {
    args: {
        capture: true,
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}

/**
 * Without any tooltip, using `notooltip`, so no tooltip appears when hovering
 * with the mouse (or a pen) over the containee badge.
 */
export const NoTooltip: Story = {
    args: {
        notooltip: true,
    },
}
