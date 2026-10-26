// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { GHOSTWIRE_LABEL_ROOT, type NetworkInterface, TapTunMode } from 'models/gw'
import { nifBr, nifHw, nifLo, nifMacvlan } from 'models/gw/mock'

import { NifIcon } from './NifIcon'

const virtualNif: NetworkInterface = {
    ...nifHw,
    isPhysical: false,
    macvlans: undefined,
}

const meta: Meta<typeof NifIcon> = {
    title: 'Ghostwire/NifIcon',
    component: NifIcon,
    args: {
        nif: nifHw,
        fontSize: 'large',
    },
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof NifIcon>

export const Physical: Story = {}

export const ConsiderPhysical: Story = {
    args: {
        considerPhysical: true,
    },
}

export const GenericNic: Story = {
    args: {
        nif: virtualNif,
    },
}

export const Loopback: Story = {
    args: {
        nif: nifLo,
    },
}

export const Bridge: Story = {
    args: {
        nif: { ...virtualNif, kind: 'bridge' },
    },
}

export const InternalBridge: Story = {
    args: {
        nif: { ...nifBr, labels: { [GHOSTWIRE_LABEL_ROOT + 'bridge/internal']: '' } },
    },
}

export const Dummy: Story = {
    args: {
        nif: { ...virtualNif, kind: 'dummy' },
    },
}

export const Veth: Story = {
    args: {
        nif: { ...virtualNif, kind: 'veth' },
    },
}

export const Macvlan: Story = {
    args: {
        nif: { ...virtualNif, kind: 'macvlan' },
    },
}

export const MacvlanMaster: Story = {
    args: {
        nif: { ...virtualNif, macvlans: [nifMacvlan] },
    },
}

export const Ipvlan: Story = {
    args: {
        nif: { ...virtualNif, kind: 'ipvlan' },
    },
}

export const IpvlanMaster: Story = {
    args: {
        nif: { ...virtualNif, ipvlans: [{ ...virtualNif, kind: 'ipvlan' }] },
    },
}

export const Netkit: Story = {
    args: {
        nif: { ...virtualNif, kind: 'netkit' },
    },
}

export const SingleNetkit: Story = {
    args: {
        nif: {
            ...virtualNif, kind: 'netkit',
            netkitDetails: { mode: 'L3', policy: 'forward', primary: true, single: true },
        },
    },
}

export const Tap: Story = {
    args: {
        nif: { ...virtualNif, kind: 'tuntap', tuntapDetails: { mode: TapTunMode.TAP, processors: [] } },
    },
}

export const Tun: Story = {
    args: {
        nif: { ...virtualNif, kind: 'tuntap', tuntapDetails: { mode: TapTunMode.TUN, processors: [] } },
    },
}

export const Vxlan: Story = {
    args: {
        nif: { ...virtualNif, kind: 'vxlan' },
    },
}
