// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import type { Meta, StoryObj } from '@storybook/react-vite'

import { infiniteLifetime } from 'components/address/IpAddressAndPrefix'
import { CardTray } from 'components/cardtray'
import { NetnsPlainCard } from 'components/netnsplaincard'
import {
    AddressFamily, loFiNetns, type NetworkInterface, type NetworkNamespace, OperationalState,
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

const operstate = (down: boolean) => down ? OperationalState.Down : OperationalState.Up

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

// A host with a bridge having a veth port that is paired with a container's
// veth; optionally the bridge is down.
const bridgeNetnses = (bridgeDown = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const br = newNif(host, 3, 'mybridge', 'bridge', '172.17.0.1', {
        slaves: [], operstate: operstate(bridgeDown),
    })
    const veth = newNif(host, 5, 'veth4a1b2c3', 'veth', undefined, { master: br })
    br.slaves!.push(veth)
    const container = newNetns(4026532501, { ...containerBox, name: 'bridge_moby' })
    newNif(container, 1, 'lo', '', '127.0.0.1')
    veth.peer = newNif(container, 4, 'eth0', 'veth', '172.17.0.2', { peer: veth })
    return [host, container]
}

// A host with a hardware NIC serving as the master of IPVLANs in two
// containers; optionally the master or the first IPVLAN is down.
const ipvlanNetnses = (masterDown = false, slaveDown = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, ipvlans: [],
        operstate: operstate(masterDown),
    })
    const containers = [1, 2].map(no => {
        const netns = newNetns(4026532100 + no, { ...containerBox, name: `ipvlan_moby_${no}` })
        newNif(netns, 1, 'lo', '', '127.0.0.1')
        eth0.ipvlans!.push(newNif(netns, 2, 'eth0', 'ipvlan', `192.168.1.2${no}`, {
            ipvlan: eth0, operstate: operstate(slaveDown && no === 1),
        }))
        return netns
    })
    return [host, ...containers]
}

// A host with a hardware NIC serving as the master of MACVLANs in two
// containers; optionally the master or the first MACVLAN is down.
const macvlanNetnses = (masterDown = false, slaveDown = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, macvlans: [],
        operstate: operstate(masterDown),
    })
    const containers = [1, 2].map(no => {
        const netns = newNetns(4026532300 + no, { ...containerBox, name: `macvlan_moby_${no}` })
        newNif(netns, 1, 'lo', '', '127.0.0.1')
        eth0.macvlans!.push(newNif(netns, 2, 'eth0', 'macvlan', `192.168.1.3${no}`, {
            macvlan: eth0, operstate: operstate(slaveDown && no === 1),
        }))
        return netns
    })
    return [host, ...containers]
}

// A host with a hardware NIC serving as the master interface of two VLANs, one
// of them in a container; optionally the master or the host's VLAN is down.
const vlanNetnses = (masterDown = false, vlanDown = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, slaves: [],
        operstate: operstate(masterDown),
    })
    const vlanDetails = (vid: number) => ({ vid, vlanProtocol: 0x8100 })
    eth0.slaves!.push(newNif(host, 3, 'eth0.10', 'vlan', '10.10.0.1', {
        master: eth0, vlanDetails: vlanDetails(10), operstate: operstate(vlanDown),
    }))
    const container = newNetns(4026532501, { ...containerBox, name: 'vlan_moby' })
    newNif(container, 1, 'lo', '', '127.0.0.1')
    eth0.slaves!.push(newNif(container, 2, 'eth0', 'vlan', '10.20.0.2', {
        master: eth0, vlanDetails: vlanDetails(20),
    }))
    return [host, container]
}

// A host with a hardware NIC serving as the underlay of a VXLAN in the host as
// well as of a VXLAN in a container; optionally the underlay or the host's
// VXLAN is down.
const vxlanNetnses = (underlayDown = false, overlayDown = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(host, 2, 'eth0', '', '192.168.1.10', {
        isPhysical: true, driverinfo: { driver: 'e1000e' }, overlays: [],
        operstate: operstate(underlayDown),
    })
    const vxlanDetails = (vid: number) => ({
        vid, arpProxy: false, remotePort: 4789, sourcePortLow: 0, sourcePortHigh: 0,
    })
    eth0.overlays!.push(newNif(host, 3, 'vxlan42', 'vxlan', '10.42.0.1', {
        underlay: eth0, vxlanDetails: vxlanDetails(42), operstate: operstate(overlayDown),
    }))
    const overlay = newNetns(4026532401, { ...containerBox, name: 'vxlan_moby' })
    newNif(overlay, 1, 'lo', '', '127.0.0.1')
    eth0.overlays!.push(newNif(overlay, 2, 'vxlan0', 'vxlan', '10.0.0.2', {
        underlay: eth0, vxlanDetails: vxlanDetails(4096),
    }))
    return [host, overlay]
}

// A host with a netkit pair into a container, and a single-mode netkit
// (without any peer) in another container; optionally the container's netkit
// peer and the single-mode netkit are down.
const netkitNetnses = (down = false) => {
    const host = newNetns(4026531840, standaloneBox, true)
    newNif(host, 1, 'lo', '', '127.0.0.1')
    const nk0 = newNif(host, 2, 'nk0', 'netkit', '10.0.0.1', {
        netkitDetails: { mode: 'L3', policy: 'forward', peerPolicy: 'forward', primary: true, single: false },
    })
    const paired = newNetns(4026532201, { ...containerBox, name: 'netkit_moby' })
    newNif(paired, 1, 'lo', '', '127.0.0.1')
    const eth0 = newNif(paired, 2, 'eth0', 'netkit', '10.0.0.2', {
        netkitDetails: { mode: 'L3', policy: 'forward', peerPolicy: 'forward', primary: false, single: false },
        peer: nk0, operstate: operstate(down),
    })
    nk0.peer = eth0
    const single = newNetns(4026532202, { ...containerBox, name: 'netkit_single_moby' })
    newNif(single, 1, 'lo', '', '127.0.0.1')
    newNif(single, 2, 'nk1', 'netkit', '10.0.1.1', {
        netkitDetails: { mode: 'L3', policy: 'blackhole', primary: true, single: true },
        operstate: operstate(down),
    })
    return [host, paired, single]
}

const wired = (netnses: NetworkNamespace[]) => (
    <Breadboard netns={netnses}>
        <CardTray animate>
            {netnses.toSorted(orderNetnsByContainees)
                .map(netns => <NetnsPlainCard netns={loFiNetns(netns, true)} key={netns.netnsid} />)}
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

/**
 * Hover over network interface badges or the wires between them to enjoy some
 * wire animation.
 */
export const AllNetworkNamespaces: Story = {
    render: () => wired(allNetnses),
}

/**
 * The host's hardware NICs `eth0` and `eth1`, each with an external wire going
 * off into reality. `eth0` is up, thus its wire is rendered as up; `eth1` is
 * down, thus its wire is rendered as down.
 */
export const ExternalNic: Story = {
    render: () => wired(externalNetnses),
}

/**
 * Dummy interfaces lead nowhere: their wires end in a dead-end marker instead
 * of going off into reality. The host's `dummy0` is up, thus its wire is
 * rendered as up; `dummy1` in the `dummy_moby` container is down, thus its wire
 * is rendered as down.
 */
export const DummyNic: Story = {
    render: () => wired(dummyNetnses),
}

/**
 * The host's bridge `mybridge` with its port `veth4a1b2c3`, wired to its veth
 * peer `eth0` in the `bridge_moby` container. All interfaces are up, thus the
 * wire is rendered as up.
 */
export const BridgeWithPorts: Story = {
    render: () => wired(bridgeNetnses()),
}

/**
 * Same as "Bridge With Ports", but the bridge `mybridge` is down. Its port
 * `veth4a1b2c3` and the veth peer `eth0` are still up, so their wire is
 * rendered as up.
 */
export const BridgeDown: Story = {
    render: () => wired(bridgeNetnses(true)),
}

/**
 * Two IPVLANs wired to their master interface, the host's hardware NIC `eth0`:
 * IPVLAN `eth0` in the `ipvlan_moby_1` container and IPVLAN `eth0` in the
 * `ipvlan_moby_2` container. All interfaces are up, thus all wires are
 * rendered as up.
 */
export const Ipvlan: Story = {
    render: () => wired(ipvlanNetnses()),
}

/**
 * Same as "Ipvlan", but the IPVLAN in the `ipvlan_moby_1` container is down,
 * thus only its wire to the master interface `eth0` is rendered as down. The
 * IPVLAN in the `ipvlan_moby_2` container is up, so its wire is rendered as up.
 */
export const IpvlanDown: Story = {
    render: () => wired(ipvlanNetnses(false, true)),
}

/**
 * Same as "Ipvlan", but the master interface `eth0` is down while both IPVLANs
 * are up. Nevertheless, all wires of `eth0` are rendered as down, including
 * its external wire.
 */
export const IpvlanMasterDown: Story = {
    render: () => wired(ipvlanNetnses(true)),
}

/**
 * Two MACVLANs wired to their master interface, the host's hardware NIC `eth0`:
 * MACVLAN `eth0` in the `macvlan_moby_1` container and MACVLAN `eth0` in the
 * `macvlan_moby_2` container. All interfaces are up, thus all wires are
 * rendered as up.
 */
export const Macvlan: Story = {
    render: () => wired(macvlanNetnses()),
}

/**
 * Same as "Macvlan", but the MACVLAN in the `macvlan_moby_1` container is down,
 * thus only its wire to the master interface `eth0` is rendered as down. The
 * MACVLAN in the `macvlan_moby_2` container is up, so its wire is rendered as
 * up.
 */
export const MacvlanDown: Story = {
    render: () => wired(macvlanNetnses(false, true)),
}

/**
 * Same as "Macvlan", but the master interface `eth0` is down while both
 * MACVLANs are up. Nevertheless, all wires of `eth0` are rendered as down,
 * including its external wire.
 */
export const MacvlanMasterDown: Story = {
    render: () => wired(macvlanNetnses(true)),
}

/**
 * Two VLANs wired to their master interface, the host's hardware NIC `eth0`:
 * VLAN 10 `eth0.10` in the host itself and VLAN 20 `eth0` in the `vlan_moby`
 * container. All interfaces are up, thus all wires are rendered as up.
 */
export const Vlan: Story = {
    render: () => wired(vlanNetnses()),
}

/**
 * Same as "Vlan", but the host's VLAN `eth0.10` is down, thus only its wire to
 * the master interface `eth0` is rendered as down. The container's VLAN `eth0`
 * is up, so its wire is rendered as up.
 */
export const VlanDown: Story = {
    render: () => wired(vlanNetnses(false, true)),
}

/**
 * Same as "Vlan", but the master interface `eth0` is down while both VLANs are
 * up. Nevertheless, all wires of `eth0` are rendered as down, including its
 * external wire.
 */
export const VlanMasterDown: Story = {
    render: () => wired(vlanNetnses(true)),
}

/**
 * Two VXLANs wired to their underlay, the host's hardware NIC `eth0`: VXLAN
 * `vxlan42` with VNI 42 in the host itself and VXLAN `vxlan0` with VNI 4096 in
 * the `vxlan_moby` container. All interfaces are up, thus all wires are
 * rendered as up.
 */
export const Vxlan: Story = {
    render: () => wired(vxlanNetnses()),
}

/**
 * Same as "Vxlan", but the host's VXLAN `vxlan42` is down, thus only its wire
 * to the underlay `eth0` is rendered as down. The container's VXLAN `vxlan0`
 * is up, so its wire is rendered as up.
 */
export const VxlanDown: Story = {
    render: () => wired(vxlanNetnses(false, true)),
}

/**
 * Same as "Vxlan", but the underlay `eth0` is down while both VXLANs are up.
 * Nevertheless, all wires of `eth0` are rendered as down, including its
 * external wire.
 */
export const VxlanUnderlayDown: Story = {
    render: () => wired(vxlanNetnses(true)),
}

/**
 * The host's netkit `nk0` wired to its netkit peer `eth0` in the `netkit_moby`
 * container. The single-mode netkit `nk1` in the `netkit_single_moby`
 * container has no peer, thus no wire. All interfaces are up, thus the wire is
 * rendered as up.
 */
export const Netkit: Story = {
    render: () => wired(netkitNetnses()),
}

/**
 * Same as "Netkit", but the netkit peer `eth0` in the `netkit_moby` container
 * as well as the single-mode netkit `nk1` are down. Thus, the wire between
 * `nk0` and `eth0` is rendered as down.
 */
export const NetkitDown: Story = {
    render: () => wired(netkitNetnses(true)),
}
