// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box, Typography } from '@mui/material'

import { DynVarsProvider } from 'components/dynvars'
import { useContextualId } from 'components/idcontext'
import { type NetworkInterface, nifId, OperationalState, TapTunMode } from 'models/gw'
import { initialNetns, nifHw, nifLo, nifMacvlan } from 'models/gw/mock'
import { scrollIdIntoView } from 'utils'

import { NifBadge } from './NifBadge'

const ScrollingButtonBadge = ({ nif }: { nif: NetworkInterface }) => {
    const domIdBase = useContextualId('')
    return <NifBadge
        nif={nif}
        button
        onClick={(_, nif) => scrollIdIntoView(domIdBase + nifId(nif))}
    />
}

const virtualNif: NetworkInterface = {
    ...nifHw,
    isPhysical: false,
    isPromiscuous: false,
    macvlans: undefined,
}

const opstateNifs = Object.values(OperationalState)
    .map(operstate => ({ ...virtualNif, operstate }))

const kindNifs: [string, NetworkInterface][] = [
    ['hardware interface', { ...nifHw, macvlans: undefined }],
    ['MACVLAN master', { ...nifHw, macvlans: [nifMacvlan] }],
    ['loopback', nifLo],
    ['bridge', { ...virtualNif, kind: 'bridge' }],
    ['VETH', { ...virtualNif, kind: 'veth' }],
    ['MACVLAN', { ...virtualNif, kind: 'macvlan' }],
    ['VXLAN', { ...virtualNif, kind: 'vxlan' }],
    ['TAP', { ...virtualNif, kind: 'tuntap', tuntapDetails: { mode: TapTunMode.TAP, processors: [] } }],
    ['TUN', { ...virtualNif, kind: 'tuntap', tuntapDetails: { mode: TapTunMode.TUN, processors: [] } }],
    ['dummy', { ...virtualNif, kind: 'dummy' }],
]

const meta: Meta<typeof NifBadge> = {
    title: 'Ghostwire/NifBadge',
    component: NifBadge,
    args: {
        nif: nifHw,
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifBadge>

export const Physical: Story = {}

export const Promiscuous: Story = {
    args: {
        nif: { ...nifHw, isPromiscuous: true },
    },
}

// "Unknown" renders the same as "up": the driver doesn't report any status.
export const OperationalStates: Story = {
    render: () => <>
        {opstateNifs.map(nif => (
            <Box sx={{ py: 1 }} key={nif.operstate}>
                <Typography variant="body2" color="textSecondary">{nif.operstate}:</Typography>
                <NifBadge nif={nif} />
            </Box>
        ))}
    </>,
}

export const Kinds: Story = {
    render: () => <>
        {kindNifs.map(([description, nif]) => (
            <Box sx={{ py: 1 }} key={description}>
                <Typography variant="body2" color="textSecondary">{description}:</Typography>
                <NifBadge nif={nif} />
            </Box>
        ))}
    </>,
}

export const ButtonAndAnchors: Story = {
    parameters: {
        docs: {
            description: {
                story: 'Anchor badges (`anchor`) act as navigation targets, while button '
                    + 'badges (`button`) only report clicks via `onClick`; it is up to the '
                    + 'parent to then navigate, such as scrolling the anchor badge of the same '
                    + 'network interface into view. The large gap is deliberate: scroll the anchor '
                    + 'badges out of view, then click the button badge at the bottom to smoothly '
                    + `scroll the "${nifHw.name}" anchor badge back into view.`,
            },
        },
    },
    render: () => <>
        {Object.values(initialNetns.nifs).map(nif => (
            <Box sx={{ py: 1 }} key={nif.name}>
                <NifBadge nif={nif} anchor />
            </Box>
        ))}
        <Typography variant="body2" color="textSecondary" sx={{ mt: 2, height: '30ex' }}>
            Deliberate gap: scroll the anchor badges above out of view, then click
            the button badge below to scroll the "{nifHw.name}" anchor badge back
            into view.
        </Typography>
        <ScrollingButtonBadge nif={nifHw} />
    </>,
}

export const Button: Story = {
    args: {
        button: true,
        onClick: fn(),
    },
}

export const Stretched: Story = {
    args: {
        stretch: true,
        style: { width: '20em' },
    },
}

export const NoTooltip: Story = {
    args: {
        notooltip: true,
    },
}

export const Checkbox: Story = {
    args: {
        nifCheck: true,
        checked: true,
        onChange: fn(),
    },
}

// Capture buttons need the "enableMonolith" dynamic variable.
export const Capture: Story = {
    args: {
        capture: true,
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}

export const CaptureAlignedRight: Story = {
    args: {
        capture: true,
        alignRight: true,
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}
