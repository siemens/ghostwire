import data from './mockdata.json'
import { type Container, isContainer } from '../containee'
import { fromjson, type JSONObject } from '../model'
import type { NetworkNamespace } from '../netns'

export const discovery = fromjson(data as unknown as JSONObject)

export const initialNetns = Object.values(discovery.networkNamespaces)
    .find(netns => netns.isInitial)!

export const nifLo = Object.values(initialNetns.nifs)
    .find(nif => nif.name === 'lo')!

export const nifHw = Object.values(initialNetns.nifs)
    .find(nif => nif.isPhysical)!

export const nifMacvlanMaster = Object.values(initialNetns.nifs)
    .find(nif => !!nif.macvlans)!

export const nifMacvlan = nifMacvlanMaster.macvlans![0]

export const nifBr = Object.values(initialNetns.nifs)
    .find(nif => nif.kind === 'bridge' && nif.alias === 'mock_default')!

export const nifVeth1 = nifBr.slaves![0]

export const nifVeth2 = nifVeth1.peer

export const mockAddresses = [
    ...nifLo.addresses,
    ...nifHw.addresses,
]

export const crowdedNetns = Object.values(discovery.networkNamespaces)
    .find(netns => netns.containers.find(containee => containee.name.startsWith('mock_sharednet')))

// Synthesized "network:none" container, as the mock discovery data lacks one.
export const loOnlyContainerNetns = (() => {
    const protoContainer = Object.values(discovery.networkNamespaces)
        .flatMap(netns => netns.containers)
        .find(containee => isContainer(containee) && !containee.pod && !containee.project) as Container
    const netns: NetworkNamespace = {
        netnsid: 4026599999,
        isInitial: false,
        containers: [],
        nifs: {},
        routes: [],
        transportPorts: [],
        forwardedPorts: [],
        pods: [],
    }
    netns.nifs[nifLo.index] = { ...nifLo, netns }
    netns.containers.push({
        ...protoContainer,
        name: 'mock_nonetwork',
        id: 'mock_nonetwork',
        netns,
    })
    return netns
})()
