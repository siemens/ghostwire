// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { atom } from 'jotai'
import { atomWithStorage } from 'jotai/utils'

import { AddressFamily } from 'models/gw'
import { NEVER } from 'components/cutoffselector'


const themeKey = 'ghostwire.theme'

const showLoopbackKey = 'ghostwire.showlo'
const showEmptyNetnsKey = 'ghostwire.showemptynetns'
const showMACKey = 'ghostwire.showmac'
const showIpv4Key = 'ghostwire.ipv4'
const showIpv6Key = 'ghostwire.ipv6'

const showSandboxesKey = 'ghostwire.showsandboxes'
const showNamespaceIds = 'ghostwire.shownsids'

const cutoffContaineesKey = 'ghostwire.cutoff.containees'
const cutoffNeighborhoodKey = 'ghostwire.cutoff.neighborhood'
const cutoffPortsKey = 'ghostwire.cutoff.ports'
const cutoffForwardedPortsKey = 'ghostwire.cutoff.forwardedports'
const cutoffRoutesKey = 'ghostwire.cutoff.routes'
const cutoffNifsKey = 'ghostwire.cutoff.nifs'

const snapshotDensityKey = 'ghostwire.snapshot.density'

const showIEAppIconsKey = 'ghostwire.showieappicons'

const showMultiBroadcastRoutesKey = 'ghostwire.showmultibroadcastroutes'

const filterPatternKey = 'ghostwire.filter.pattern'
const filterCaseSensitiveKey = 'ghostwire.filter.case'
const filterRegexpKey = 'ghostwire.filter.regexp'


export const THEME_USERPREF = 0
export const THEME_LIGHT = 1
export const THEME_DARK = -1
export const themeAtom = atomWithStorage(themeKey, THEME_USERPREF)

export const showLoopbackAtom = atomWithStorage(showLoopbackKey, false)
export const showEmptyNetnsAtom = atomWithStorage(showEmptyNetnsKey, false)
export const showMACAtom = atomWithStorage(showMACKey, false)

// The "IP families" setting actually consists of separate IPv4 and IPv6 boolean
// settings.
export const showIpv4Atom = atomWithStorage(showIpv4Key, true)
export const showIpv6Atom = atomWithStorage(showIpv6Key, false) // this default is PAINFUL!
export const showIpFamiliesAtom = atom((get) => {
    const families: AddressFamily[] = get(showIpv4Atom) ? [AddressFamily.IPv4] : []
    if (get(showIpv6Atom)) {
        families.push(AddressFamily.IPv6)
    }
    return families
})

export const containeesCutoffAtom = atomWithStorage(cutoffContaineesKey, 3)
export const neighborhoodCutoffAtom = atomWithStorage(cutoffNeighborhoodKey, 10)
export const portsCutoffAtom = atomWithStorage(cutoffPortsKey, 10)
export const forwardedPortsCutoffAtom = atomWithStorage(cutoffForwardedPortsKey, 10)
export const routesCutoffAtom = atomWithStorage(cutoffRoutesKey, 10)
export const nifsCutoffAtom = atomWithStorage(cutoffNifsKey, NEVER)

export const showSandboxesAtom = atomWithStorage(showSandboxesKey, false)

export const showNamespaceIdsAtom = atomWithStorage(showNamespaceIds, false)

export const snapshotDensityAtom = atomWithStorage(snapshotDensityKey, 1)

export const showIEAppIconsAtom = atomWithStorage(showIEAppIconsKey, false)

export const showMultiBroadcastRoutesAtom = atomWithStorage(showMultiBroadcastRoutesKey, false)

export const filterPatternAtom = atomWithStorage(filterPatternKey, '')
export const filterCaseSensitiveAtom = atomWithStorage(filterCaseSensitiveKey, false)
export const filterRegexpAtom = atomWithStorage(filterRegexpKey, false)
