// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'
import { fn } from 'storybook/test'

import { Box, Typography } from '@mui/material'

import { DynVarsProvider } from '../dynvars'
import { useContextualId } from '../idcontext'
import { type NetworkInterface, nifId, OperationalState, TapTunMode } from '../../models/gw'
import { initialNetns, nifHw, nifLo, nifMacvlan } from '../../models/gw/mock'
import { Tile, Tiles } from '../../stories'
import { scrollIdIntoView } from '../../utils'

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
    ['bridge', { ...virtualNif, kind: 'bridge', operstate: OperationalState.Up }],
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

/**
 * "Physical" network interfaces are designated by a "NIC" icon preceding the
 * badge.
 */
export const Physical: Story = {}

/**
 * A network interface put into promiscuous mode displays a lughole
 * (eavesdropper) icon.
 */
export const Promiscuous: Story = {
    args: {
        nif: { ...nifHw, isPromiscuous: true },
    },
}

/**
 * The operational status of a network interface is rendered inside the badge
 * using an icon. As an additional visual clue, the status icon is colored
 * according to state. Please note that the "unknown" status is rendered the
 * same as the "up" state: "unknown" means that the corresponding network
 * interface driver doesn't report any status, yet the interface is not
 * considered to be down.
 */
export const OperationalStates: Story = {
    render: () => (
        <Tiles>
            {opstateNifs.map(nif => (
                <Tile key={nif.operstate}>
                    <Typography variant="body2" color="textSecondary">{nif.operstate}:</Typography>
                    <NifBadge nif={nif} />
                </Tile>
            ))}
        </Tiles>
    ),
}

/** The type of network interface is indicated by an icon inside the badge. */
export const Kinds: Story = {
    render: () => (
        <Tiles>
            {kindNifs.map(([description, nif]) => (
                <Tile key={description}>
                    <Typography variant="body2" color="textSecondary">{description}:</Typography>
                    <NifBadge nif={nif} />
                </Tile>
            ))}
        </Tiles>
    ),
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

/**
 * A network interface badge optionally can stretch in order to take up the
 * available horizontal space; this requires the `stretch` property, but also
 * suitable element width setting, as otherwise the `NifBadge` component won't
 * stretch.
 */
export const Stretched: Story = {
    args: {
        stretch: true,
        style: { width: '20em' },
    },
}

/**
 * In some contexts where an outer component already provides a tooltip, you
 * might not want to have an additional tooltip on the network interface badge,
 * as this quickly gets confusing. Use `notooltip` to switch off the tooltip on
 * a network interface badge.
 */
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

/**
 * Additional network interface capture buttons need to be requested through
 * the `capture` property and then are enabled only through the dynamic
 * variable `enableMonolith`, passed in by the server serving the application
 * (see also `DynVarsProvider`, which this story uses to enable capture
 * buttons).
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
 * Depending on the alignment of the badge, the capture button will be shown
 * either to the right or left of the badge.
 */
export const CaptureAlignedRight: Story = {
    args: {
        capture: true,
        alignRight: true,
    },
    decorators: [
        (Story) => <DynVarsProvider value={{ enableMonolith: true }}><Story /></DynVarsProvider>,
    ],
}

/**
 * A plain network interface badge side-by-side to a button badge. The button
 * variant has rounded edges and reacts to hovering the mouse over it or
 * clicking it, not least with a ripple animation.
 */
export const PlainVersusButton: Story = {
    render: () => (
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1 }}>
            <NifBadge nif={nifHw} /> ⟵ versus ⟶ <NifBadge nif={nifHw} button />
        </Box>
    ),
}
