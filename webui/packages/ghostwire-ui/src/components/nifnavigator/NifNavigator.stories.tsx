// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useState } from 'react'

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box, Paper, Typography } from '@mui/material'

import { DynVarsProvider } from '../dynvars'
import { firstContainee, orderNifByName } from '../../models/gw'
import { discovery, nifHw, nifLo, nifVeth1, nifVeth2 } from '../../models/gw/mock'

import { NifNavigator, type NifNavigatorProps } from './NifNavigator'

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

/** Click on the badge to pop up a menu with the related network interfaces. */
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

/** Nothing to navigate to. */
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

/**
 * You'll most probably see several `eth0` network interfaces, which actually
 * are VETH interfaces ... rest assured, they belong to different containers,
 * this story just doesn't show the containers. When you click on such a
 * network interface badge, you'll see the related VETH peer interface and the
 * container it belongs to (usually `init (1)`).
 */
export const SeveralVeths: Story = {
    render: (args) => <>
        {veths.map(nif => (
            <Box sx={{ py: 1 }} key={`${nif.netns.netnsid}-${nif.name}`}>
                <NifNavigator {...args} nif={nif} />
            </Box>
        ))}
    </>,
}

// Keeps track of the most recent navigation in order to show it.
const NavigatedToDemo = (args: NifNavigatorProps) => {
    const [navigatedTo, setNavigatedTo] = useState('')
    return <>
        <Typography>NAVIGATED TO: {navigatedTo}</Typography>
        {[nifHw, ...veths].map(nif => (
            <Paper variant="outlined" sx={{ p: 1, my: 1 }} key={`${nif.netns.netnsid}-${nif.name}`}>
                <NifNavigator
                    {...args}
                    nif={nif}
                    onNavigation={(navnif) => {
                        args.onNavigation?.(navnif)
                        setNavigatedTo(`${navnif.name} [containee: ${firstContainee(navnif.netns).name}]`)
                    }}
                />
            </Paper>
        ))}
    </>
}

/**
 * This story renders some `NifNavigator` components, including a MACVLAN
 * master interface as well as a few VETH peer interfaces. When clicking on a
 * (navigatable) network interface badge, a pop-up menu appears listing the
 * related network interfaces.
 *
 * When selecting one of the related network interfaces, the "NAVIGATED TO:"
 * information will update accordingly to show that the callback triggered
 * correctly and also got correct interface navigation information.
 */
export const NavigatedTo: Story = {
    render: (args) => <NavigatedToDemo {...args} />,
}
