import data from './mockdata.json'
import { fromjson } from '../model'

export const discovery = fromjson(data)

export const initialNetns = Object.values(discovery.networkNamespaces)
    .find(netns => netns.isInitial)

export const nifLo = Object.values(initialNetns.nifs)
    .find(nif => nif.name === 'lo')

export const nifHw = Object.values(initialNetns.nifs)
    .find(nif => nif.isPhysical)

export const nifMacvlanMaster = Object.values(initialNetns.nifs)
    .find(nif => !!nif.macvlans)

export const nifMacvlan = nifMacvlanMaster.macvlans[0]

export const nifBr = Object.values(initialNetns.nifs)
    .find(nif => nif.kind === 'bridge' && nif.alias === 'mock_default')

export const nifVeth1 = nifBr.slaves[0]

export const nifVeth2 = nifVeth1.peer

export const mockAddresses = [
    ...nifLo.addresses,
    ...nifHw.addresses,
]

export const crowdedNetns = Object.values(discovery.networkNamespaces)
    .find(netns => netns.containers.find(containee => containee.name.startsWith('mock_sharednet')))
