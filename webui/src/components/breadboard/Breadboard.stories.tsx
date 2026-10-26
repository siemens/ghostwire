// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { infiniteLifetime } from 'components/address/IpAddressAndPrefix'
import { CardTray } from 'components/cardtray'
import { NetnsPlainCard } from 'components/netnsplaincard'
import {
    AddressFamily, type NetworkInterface, type NetworkNamespace, OperationalState,
    orderNetnsByContainees, type PrimitiveContainee, SRIOVRole,
} from 'models/gw'
import { discovery } from 'models/gw/mock'
import { containerBox, standaloneBox } from 'views/help/mockdata'

import { Breadboard } from './Breadboard'

const allNetnses = Object.values(discovery.networkNamespaces)

const newNetns = (netnsid: number, containee: PrimitiveContainee, isInitial = false) => {
    const netns: NetworkNamespace = {
        netnsid, isInitial, containers: [], nifs: {}, routes: [],
        transportPorts: [], forwardedPorts: [], pods: [],
    }
    netns.containers.push({ ...containee, netns })
    return netns
}

const newNif = (
    netns: NetworkNamespace, index: number, name: string, kind: string,
    ipv4?: string, details: Partial<NetworkInterface> = {},
) => {
    const nif: NetworkInterface = {
        netns, index, name, kind,
        operstate: OperationalState.Up,
        isPhysical: false,
        isLoopback: false,
        isPromiscuous: false,
        sriovrole: SRIOVRole.None,
        driverinfo: { driver: kind },
        labels: {},
        addresses: ipv4 ? [{
            address: ipv4, prefixlen: 24, family: AddressFamily.IPv4,
            preferredLifetime: infiniteLifetime, validLifetime: infiniteLifetime, scope: 0,
        }] : [],
        ...details,
    }
    netns.nifs[index] = nif
    return nif
}

// A host with hardware NICs going off into reality, one of them being down.
const externalNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' },
    })
    newNif(host, 3, 'eth1', '', undefined, {
        isPhysical: true, driverinfo: { driver: 'igc' }, operstate: OperationalState.Down,
    })
    return [host]
})()

// Dummy NICs lead nowhere, one of them being down.
const dummyNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    newNif(host, 2, 'dummy0', 'dummy', '10.10.0.1')
    const container = newNetns(4026532501, { ...containerBox, name: 'dummy_moby' })
    newNif(container, 1, 'lo', '', '127.0.0.1')
    newNif(container, 2, 'dummy1', 'dummy', '10.10.1.1', { operstate: OperationalState.Down })
    return [host, container]
})()

// A host with a hardware NIC serving as the master of IPVLANs in two containers.
const ipvlanNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, ipvlans: [],
    })
    const containers = [1, 2].map(no => {
        const netns = newNetns(4026532100 + no, { ...containerBox, name: `ipvlan_moby_${no}` })
        newNif(netns, 1, 'lo', '', '127.0.0.1')
        eth0.ipvlans!.push(newNif(netns, 2, 'eth0', 'ipvlan', `192.168.1.2${no}`, { ipvlan: eth0 }))
        return netns
    })
    return [host, ...containers]
})()

// A host with a hardware NIC serving as the master of MACVLANs in two containers.
const macvlanNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, macvlans: [],
    })
    const containers = [1, 2].map(no => {
        const netns = newNetns(4026532300 + no, { ...containerBox, name: `macvlan_moby_${no}` })
        newNif(netns, 1, 'lo', '', '127.0.0.1')
        eth0.macvlans!.push(newNif(netns, 2, 'eth0', 'macvlan', `192.168.1.3${no}`, { macvlan: eth0 }))
        return netns
    })
    return [host, ...containers]
})()

// A host with a hardware NIC serving as the underlay of a VXLAN in the host as
// well as of a VXLAN in a container.
const vxlanNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, overlays: [],
    })
    const vxlanDetails = (vid: number) => ({
        vid, arpProxy: false, remotePort: 4789, sourcePortLow: 0, sourcePortHigh: 0,
    })
    eth0.overlays!.push(newNif(host, 3, 'vxlan42', 'vxlan', '10.42.0.1', {
        underlay: eth0, vxlanDetails: vxlanDetails(42),
    }))
    const overlay = newNetns(4026532401, { ...containerBox, name: 'vxlan_moby' })
    newNif(overlay, 1, 'lo', '', '127.0.0.1')
    eth0.overlays!.push(newNif(overlay, 2, 'vxlan0', 'vxlan', '10.0.0.2', {
        underlay: eth0, vxlanDetails: vxlanDetails(4096),
    }))
    return [host, overlay]
})()

// A host with a netkit pair into a container, and a single-mode netkit
// (without any peer) in another container.
const netkitNetnses = (() => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const nk0 = newNif(host, 2, 'nk0', 'netkit', '10.0.0.1', {
        netkitDetails: { mode: 'L3', policy: 'forward', peerPolicy: 'forward', primary: true, single: false },
    })
    const paired = newNetns(4026532201, { ...containerBox, name: 'netkit_moby' })
    newNif(paired, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(paired, 2, 'eth0', 'netkit', '10.0.0.2', {
        netkitDetails: { mode: 'L3', policy: 'forward', peerPolicy: 'forward', primary: false, single: false },
        peer: nk0,
    })
    nk0.peer = eth0
    const single = newNetns(4026532202, { ...containerBox, name: 'netkit_single_moby' })
    newNif(single, 1, 'lo', '', '127.0.0.1')
    newNif(single, 2, 'nk1', 'netkit', '10.0.1.1', {
        netkitDetails: { mode: 'L3', policy: 'blackhole', primary: true, single: true },
    })
    return [host, paired, single]
})()

const wired = (netnses: NetworkNamespace[]) => (
    <Breadboard netns={netnses}>
        <CardTray animate>
            {netnses.toSorted(orderNetnsByContainees)
                .map(netns => <NetnsPlainCard netns={netns} key={netns.netnsid} />)}
        </CardTray>
    </Breadboard>
)

const meta: Meta<typeof Breadboard> = {
    title: 'Ghostwire/Breadboard',
    component: Breadboard,
    argTypes: {
        netns: { control: false },
    },
    tags: ['autodocs'],
}

export default meta

type Story = StoryObj<typeof Breadboard>

// Hover over network interface badges or the wires between them to enjoy some
// wire animation.
export const AllNetworkNamespaces: Story = {
    render: () => wired(allNetnses),
}

export const ExternalNic: Story = {
    render: () => wired(externalNetnses),
}

export const DummyNic: Story = {
    render: () => wired(dummyNetnses),
}

export const Ipvlan: Story = {
    render: () => wired(ipvlanNetnses),
}

export const Macvlan: Story = {
    render: () => wired(macvlanNetnses),
}

export const Vxlan: Story = {
    render: () => wired(vxlanNetnses),
}

export const Netkit: Story = {
    render: () => wired(netkitNetnses),
}
