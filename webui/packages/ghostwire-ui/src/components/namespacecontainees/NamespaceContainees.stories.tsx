// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box, Paper } from '@mui/material'

import { useContextualId } from '../idcontext'
import { NetnsDetailCard } from '../netnsdetailcard'
import { netnsId } from '../../models/gw'
import { crowdedNetns, initialNetns } from '../../models/gw/mock'
import { scrollIdIntoView } from '../../utils'

import { NamespaceContainees, type NamespaceContaineesProps } from './NamespaceContainees'

const meta: Meta<typeof NamespaceContainees> = {
    title: 'Ghostwire/NamespaceContainees',
    component: NamespaceContainees,
    args: {
        netns: crowdedNetns!,
        onNavigation: fn(),
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NamespaceContainees>

/**
 * In this story there are several containers, processes, et cetera attached to
 * the same network namespace, hitting the display limit ingrained into the
 * `NamespaceContainees` component. The list is thus cut short and an ellipsis
 * shown to indicate that some inhabitants of the network namespace have been
 * left out.
 */
export const Crowded: Story = {}

/** The list is cut short and an ellipsis indicates left out containees. */
export const Limited: Story = {
    args: {
        limit: 2,
    },
}

export const InitialNamespace: Story = {
    args: {
        netns: initialNetns,
    },
}

// Scrolls the network namespace (card) into view when navigating from one of
// its containees.
const NavigationDemo = (args: NamespaceContaineesProps) => {
    const domIdBase = useContextualId('')
    return <>
        <Paper variant="outlined" sx={{ p: 1 }}>
            <NamespaceContainees
                {...args}
                onNavigation={(containee) => {
                    args.onNavigation?.(containee)
                    scrollIdIntoView(domIdBase + netnsId(args.netns))
                }}
            />
        </Paper>
        <Box sx={{ mt: '30ex' }}>
            <NetnsDetailCard netns={args.netns} />
        </Box>
    </>
}

/**
 * Users can navigate from the list to the corresponding network namespace
 * (card). Clicking or tapping on a reference badge will scroll the network
 * namespace into view, if necessary.
 */
export const Navigation: Story = {
    render: (args) => <NavigationDemo {...args} />,
}
