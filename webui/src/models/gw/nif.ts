// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { IpAddress } from './address'
import { firstContainee, NetworkNamespace } from './netns'
import { PortUser } from './ports'

/**
 * The network interfaces of a network namespace are keyed by their so-called
 * "index" numbers, from 1 onwards.
 */
export type NetworkInterfaces = { [key: number]: NetworkInterface }

/**
 * Information about a particular network interface in some network namespace.
 */
export interface NetworkInterface {
    netns: NetworkNamespace /** network namespace we belong to. */
    name: string /** name of network interface. */
    alias?: string /** some user-space assigned name for this interface. */
    index: number /** nif index within network namespace. */
    /**
     * kind of (virtual) network interface, or "" if this is a loopback
     * interface or a hardware interface. Check isPhysical to differentiate
     * loopback interfaces from hardware interfaces. For some historical
     * reason, Linux doesn't deem loopback interfaces to be virtual interface
     * and doesn't give it its own kind.
     */
    kind: string
    operstate: OperationalState /** operational state of network interface. */
    /** true if not a virtual interface, but has (maybe virtual) hardware attached. */
    isPhysical: boolean
    isPromiscuous: boolean /** true if in promiscuous mode. */
    sriovrole: SRIOVRole /** SR-IOV role PF or VF, if applicable. */

    /** nif labels, if any. */
    labels: { [key: string]: string }

    macvlans?: NetworkInterface[] /* attached macvlan interface(s) */
    slaves?: NetworkInterface[]
    overlays?: NetworkInterface[]

    pf?: NetworkInterface /** if a VF where we found its PF */
    master?: NetworkInterface /** if enslaved, the bridge interface */
    macvlan?: NetworkInterface /** if macvlan this is our "master" interface */
    peer?: NetworkInterface /** veth peer network interface */
    underlay?: NetworkInterface /** vxlan underlay network interface */

    addresses: IpAddress[] /** IP addresses assigned to this network interface */

    tuntapDetails?: TunTapDetails
    vxlanDetails?: VxlanDetails
    vlanDetails?: VlanDetails
}

/**
 * Type guard for `NetworkInterface` objects. Returns true only if the given
 * object is a network interface object, otherwise false.
 *
 * @param nif network interface object to be type-guarded.
 */
export const isNetworkInterface = (nif: unknown): nif is NetworkInterface => {
    return (nif as NetworkInterface)?.netns !== undefined
        && (nif as NetworkInterface)?.index !== undefined
}

export enum TapTunMode {
    TAP = 'tap',
    TUN = 'tun',
}

export type TapTunProcessor = PortUser

export interface TunTapDetails {
    mode: TapTunMode /** "tap" or "tun" */
    processors: TapTunProcessor[] /** processes serving this TAP/TUN */
}

export interface VxlanDetails {
    vid: number
    arpProxy: boolean
    //remote: TODO:
    remotePort: number
    //source: TODO:
    sourcePortLow: number
    sourcePortHigh: number
}

export interface VlanDetails {
    vid: number
    vlanProtocol: number,
}

/**
 * The SR-IOV roles apply only to PCI network interfaces with SR-IOV
 * functionality enabled. To ease handling, None represents both non-SR-IOV
 * network interfaces as well as network interfaces on which SR-IOV is disabled.
 */
export enum SRIOVRole {
    None = 0,
    PF,
    VF,
}

/**
 * The operational states of network interfaces as reported via RTNETLINK and
 * passed on by Ghostwire. Please note that some operational states are not used
 * by Linux kernels, such as Dormant, Testing, and NotPresent.
 */
export enum OperationalState {
    Unknown = 'UNKNOWN', /** unknown is considered to be "running", like "up". */
    NotPresent = 'NOTPRESENT',
    Down = 'DOWN',
    Up = 'UP', /** considered to be "running", like "unknown". */
    LowerLayerDown = 'LOWERLAYERDOWN',
    Dormant = 'DORMANT',
    Testing = 'TESTING'
}

/**
 * Returns true, if the specified network interface is operational in some
 * sense: that is, beyond being up, it might be in an unknown state which is
 * also considered to be operational.
 *
 * @param nif network interface object
 */
export const isOperational = (nif: NetworkInterface) => (
    ![
        OperationalState.Down,
        OperationalState.NotPresent,
        OperationalState.Testing
    ].includes(nif.operstate)
)

/**
 * Sort order function for sorting network interfaces by name, with the
 * exception of "lo" loopback interfaces always getting sorted first.
 *
 * @param nifA first network interface object.
 * @param nifB second network interface object.
 *
 * @returns -1 if nifA comes before nifB, 1 if nifA comes after nifB, and 0 if
 * nifA and nifB have equal names.
 */
export const orderNifByName = (nifA: NetworkInterface, nifB: NetworkInterface) => (
    (nifA.name === 'lo' && -1) ||
    (nifB.name === 'lo' && 1) ||
    nifA.name.localeCompare(nifB.name)
)

/**
 * Sort order function for sorting network interfaces not only by their names,
 * but then also by the containee names housing the interfaces. In case of
 * multiple containees for a network interface only the lexicographically
 * first will be used (with the exception of the initial namespace containee
 * initwhatever being always first).
 *
 * Please note that this sort order function does not handle the 'lo' loopback
 * interfaces differently from any other network interface name, opposed to
 * what `orderNifByName` does.
 *
 * @param nifA first network interface object.
 * @param nifB second network interface object.
 */
export const orderNifByNameAndContainee = (nifA: NetworkInterface, nifB: NetworkInterface) => {
    const d = nifA.name.localeCompare(nifB.name)
    if (d !== 0) {
        return d
    }
    const containeeA = firstContainee(nifA.netns)
    const containeeB = firstContainee(nifB.netns)
    return containeeA.name.localeCompare(containeeB.name)
}

/**
 * Returns a stable (DOM element) identifier for a given network interface. This
 * function is forgiving in view of incomplete or null network interface objects
 * in that it still returns an identifier under such conditions, albeit that
 * identifier might not be too useful.
 *
 * @param nif network interface object.
 */
export const nifId = (nif: NetworkInterface) =>
    `nif-${nif && nif.netns && nif.netns.netnsid}-${nif && nif.name}`
