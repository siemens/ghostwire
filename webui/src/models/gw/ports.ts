// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { IpAddress } from "./address";
import { PrimitiveContainee } from "./containee";

export interface TransportPort {
    /** state, such as listening, connecting, connected, et cetera. */
    state: string
    /** 
     * simplified (aggregated) state with just listening or connected,
     * ignoring any details. 
     */
    macroState: string

    protocol: string

    localAddress: IpAddress
    localPort: number
    localServicename?: string
    remoteAddress: IpAddress
    remotePort: number
    remoteServicename?: string
    
    /**
     * a truely unique feature of the Ghostwire discovery engine: Linux is not
     * only "dual stack" (IPv4 and IPv6), but it also decided on handling IPv4
     * communication in a unified manner via IPv6 address family sockets. In
     * this model, IPv4 is still IPv4 on the wire, but the same IPv6 socket
     * can be used for talking IPv4 too, by using so-called "mapped IPv4
     * addresses".
     */
    v4mapped: boolean

    /** 
     * processes using this port; these are not the containees themselves, but
     * often child processes of the containee leader process(es). 
     */
    users: PortUser[]
}

export interface PortUser {
    /** command line of the process using the port. */
    cmdline: string[]
    /** the containee this process using the port belongs to. */
    containee: PrimitiveContainee
    /** PID of process using the port. */
    pid: number
}

/**
 * Sort order function for transport ports, ordering them by their (macro)
 * state: listening ports first, connected ports later.
 * 
 * @param portA first transport port object
 * @param portA second transport port object
 */
export const orderByState = (portA: TransportPort, portB: TransportPort) => {
    // "connected" comes before "listening"...
    return portA.macroState[0].localeCompare(portB.macroState[0])
}
