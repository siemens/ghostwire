// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { NetworkInterface } from "models/gw";


export type Target = NetnsTarget | NamedTarget

/**
 * In its most basic form, all we got is a lousy network namespace identifier.
 * If we were lucky, then we got checkpoint data to verify we're still
 * referencing the same network namespace using a "canary" process with its
 * specific combination of PID and process start time.
 */
export interface NetnsTarget {
    /** 
     * Linux-kernel network namespace **identifier** (inode number). Please note
     * that this parameter actually is optional from the perspective of the
     * packetflix service, yet we're always supplying it from our UI side. It's
     * fine to set it to 0 anyway.
     */
    netns: number
    /** list of network interfaces to capture from, if not left unspecified. */
    'network-interfaces'?: string[]
    /**
     * PID of an optional "canary" process attached to the network namespace
     * to capture from. If specified, this allows the packetflix service to
     * verify that the target information isn't stale with respect to especially
     * the network namespace identifier. In case the information is detected to
     * be stale, then the packetflix service will fetch new discovery data from
     * its associated ghostwire service instace and then try to look up the
     * desired network namespace using other symbolic information, such as the
     * target name and type.
     */
    pid?: number
    /** 
     * unsigned 64bit integer number of the start time of a "canary" process
     * attached to the network namespace to capture from.
     */
    starttime?: number

    // TODO: capture-service, captureport
}

/**
 * 
 */
export interface NamedTarget extends NetnsTarget {
    /** name of target (that is, a "containee" in UI parlance). */
    name: string
    /** type of target (use containeeType). */
    type: string
    /** target namespace ~"prefix" (turtleNamespace in UI parlance). */
    prefix?: string
}
