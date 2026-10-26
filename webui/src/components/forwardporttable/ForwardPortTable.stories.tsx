// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { AddressFamily, type IpAddress, type NetworkNamespace } from 'models/gw'
import { crowdedNetns, initialNetns } from 'models/gw/mock'

import { ForwardPortTable } from './ForwardPortTable'

const ip = (address: string, family: AddressFamily): IpAddress => ({
    address,
    family,
    prefixlen: family === AddressFamily.IPv4 ? 32 : 128,
    preferredLifetime: 0,
    validLifetime: 0,
    scope: 0,
})

const forwardingNetns: NetworkNamespace = { ...initialNetns, forwardedPorts: [] }

const target = crowdedNetns!.containers[0]

forwardingNetns.forwardedPorts = [
    {
        protocol: 'tcp',
        address: ip('0.0.0.0', AddressFamily.IPv4),
        port: 8080,
        servicename: 'http-alt',
        forwardedAddress: ip('172.17.0.2', AddressFamily.IPv4),
        forwardedPort: 80,
        forwardedServicename: 'http',
        netns: forwardingNetns,
        users: [{ cmdline: ['/usr/bin/nginx', '-g', 'daemon off;'], containee: target, pid: 4242 }],
    },
    {
        protocol: 'tcp',
        address: ip('::', AddressFamily.IPv6),
        port: 8443,
        forwardedAddress: ip('fd00::2', AddressFamily.IPv6),
        forwardedPort: 443,
        forwardedServicename: 'https',
        netns: forwardingNetns,
        users: [],
    },
    {
        protocol: 'udp',
        address: ip('127.0.0.1', AddressFamily.IPv4),
        port: 5353,
        forwardedAddress: ip('172.17.0.3', AddressFamily.IPv4),
        forwardedPort: 53,
        forwardedServicename: 'domain',
        netns: forwardingNetns,
        users: [{ cmdline: ['/usr/sbin/dnsmasq', '-k'], containee: target, pid: 666 }],
    },
]

const meta: Meta<typeof ForwardPortTable> = {
    title: 'Ghostwire/ForwardPortTable',
    component: ForwardPortTable,
    args: {
        netns: forwardingNetns,
    },
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof ForwardPortTable>

export const Standard: Story = {}

export const IPv4Only: Story = {
    name: 'IPv4 Only',
    args: {
        families: [AddressFamily.IPv4],
    },
}

export const Empty: Story = {
    args: {
        netns: initialNetns,
    },
}

export const HideEmpty: Story = {
    args: {
        netns: initialNetns,
        hideEmpty: true,
    },
}
