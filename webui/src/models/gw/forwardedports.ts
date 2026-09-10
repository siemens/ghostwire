// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT
import type { IpAddress } from "./address"
import type { NetworkNamespace } from "./netns"
import type { PortUser } from "./ports"

export interface ForwardedPort {
    protocol: string
    
    address: IpAddress
    port: number
    servicename?: string
    forwardedAddress: IpAddress
    forwardedPort: number
    forwardedServicename?: string
    netns: NetworkNamespace

    /** 
     * processes having socket(s) willing to serve the forwarded port; these are
     * not the containees themselves, but often child processes of the containee
     * leader process(es). 
     */
    users: PortUser[]
}