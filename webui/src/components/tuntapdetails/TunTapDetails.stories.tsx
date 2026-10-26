// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { TapTunMode } from 'models/gw'
import { nifHw } from 'models/gw/mock'
import { containerBox, standaloneBox } from 'views/help/mockdata'

import { TunTapDetails } from './TunTapDetails'

const meta: Meta<typeof TunTapDetails> = {
    title: 'Ghostwire/TunTapDetails',
    component: TunTapDetails,
    argTypes: {
        nif: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof TunTapDetails>

export const Tap: Story = {
    args: {
        nif: {
            ...nifHw,
            name: 'tap0',
            kind: 'tuntap',
            isPhysical: false,
            tuntapDetails: {
                mode: TapTunMode.TAP,
                processors: [
                    { cmdline: ['/usr/bin/qemu-system-x86_64', '-netdev', 'tap,id=net0'], containee: standaloneBox, pid: 4242 },
                    { cmdline: ['/usr/sbin/openvpn', '--dev', 'tap0'], containee: containerBox, pid: 666 },
                ],
            },
        },
    },
}

/** Renders nothing, as there are no processes serving the TUN. */
export const Unserved: Story = {
    args: {
        nif: {
            ...nifHw,
            name: 'tun0',
            kind: 'tuntap',
            isPhysical: false,
            tuntapDetails: { mode: TapTunMode.TUN, processors: [] },
        },
    },
}
