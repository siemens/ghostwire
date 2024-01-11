// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { isContainer, Pod, PrimitiveContainee, Project, sortContaineesByName } from './containee'
import { ForwardedPort } from './forwardedports'
import { NetworkInterfaces } from './nif'
import { TransportPort } from './ports'
import { IpRoute } from './route'

export type NetworkNamespaces = { [key: number]: NetworkNamespace }

/**
 * Describes the properties of a single network namespace.
 */
export interface NetworkNamespace {
    /** namespace identifier, in form of an inode number (only). */
    netnsid: number
    /** is this the initial network namespace? */
    isInitial: boolean
    /** 
     * list of containers (and standalone processes, ...) attached to this
     * network namespace.
     */
    containers: PrimitiveContainee[]
    /** list of network interfaces in this network namespace. */
    nifs: NetworkInterfaces
    /** list of routes in this network namespace. */
    routes: IpRoute[]
    /** list of open transport-layer ports (~TSAPs). */
    transportPorts: TransportPort[]
    /** list of forwarded ports */
    forwardedPorts: ForwardedPort[]
    /** list of pods (=tightly grouped containers). */
    pods: Pod[]
    /** 
     * associated composer project if there is exactly one project associated
     * with this network namespace and all containers belong to that project,
     * but no other "surplus" containers.
     */
    project?: Project
}

/**
 * Type guard for `NetworkNamespace` objects. Returns true only if the given
 * object actually is a network namespace object, otherwise false.
 *
 * @param netns the network namespace object to be type-guarded.
 */
export const isNetworkNamespace = (netns: unknown): netns is NetworkNamespace => {
    return !!netns 
        && (netns as NetworkNamespace).netnsid !== undefined 
        && (netns as NetworkNamespace).nifs !== undefined
}

/**
 * Returns the lexicographically first containee for a given network
 * namespace. The only exception is the initial namespace, where its initial
 * process will always take containee precedence.
 *
 * @param netns network namespace object.
 */
export const firstContainee = (netns: NetworkNamespace) => {
    const containee = netns.containers.sort(sortContaineesByName)[0]
    return (isContainer(containee) && containee.pod) || containee
}

/**
 * Returns a stable (DOM element) identifier for a given network namespace. This
 * function is forgiving in view of incomplete or null network interface objects
 * in that it still returns an identifier under such conditions, albeit that
 * identifier might not be too useful.
 *
 * @param nif network namespace object.
 */
export const netnsId = (netns: NetworkNamespace) =>
    `netns-${netns && netns.netnsid}`

/**
 * Returns true, if the given network namespace only has a single and lonely
 * loopback "lo" network interface, but no other network interface.
 */
export const emptyNetns = (netns: NetworkNamespace) => {
    const nifs = Object.values(netns.nifs)
    return nifs.length === 1 && nifs[0].name === 'lo'
}
