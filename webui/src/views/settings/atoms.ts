// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { atom, type PrimitiveAtom, type SetStateAction } from 'jotai'
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


/**
 * Settings changes are persisted to local storage only in stores where this is
 * true; the help viewer turns it off for its own deterministic settings.
 */
export const persistSettingsAtom = atom(true)

const settingAtom = <T>(key: string, initialValue: T): PrimitiveAtom<T> => {
    // getOnInit, as otherwise unmounted settings read in other atoms' write
    // methods would always yield the initial value instead of the stored one.
    const persisted = atomWithStorage(key, initialValue, undefined, { getOnInit: true })
    const transient = atom<{ value: T } | null>(null)
    const setting: PrimitiveAtom<T> = atom(
        (get) => {
            const t = get(transient)
            return t ? t.value : get(persisted)
        },
        (get, set, update: SetStateAction<T>) => {
            const value = typeof update === 'function'
                ? (update as (prev: T) => T)(get(setting))
                : update
            if (get(persistSettingsAtom)) {
                set(persisted, value)
            } else {
                set(transient, { value })
            }
        })
    return setting
}

export const THEME_USERPREF = 0
export const THEME_LIGHT = 1
export const THEME_DARK = -1
export const themeAtom = settingAtom(themeKey, THEME_USERPREF)

export const showLoopbackAtom = settingAtom(showLoopbackKey, false)
export const showEmptyNetnsAtom = settingAtom(showEmptyNetnsKey, false)
export const showMACAtom = settingAtom(showMACKey, false)

// The "IP families" setting actually consists of separate IPv4 and IPv6 boolean
// settings.
export const showIpv4Atom = settingAtom(showIpv4Key, true)
export const showIpv6Atom = settingAtom(showIpv6Key, false) // this default is PAINFUL!
export const showIpFamiliesAtom = atom((get) => {
    const families: AddressFamily[] = get(showIpv4Atom) ? [AddressFamily.IPv4] : []
    if (get(showIpv6Atom)) {
        families.push(AddressFamily.IPv6)
    }
    return families
})

export const containeesCutoffAtom = settingAtom(cutoffContaineesKey, 3)
export const neighborhoodCutoffAtom = settingAtom(cutoffNeighborhoodKey, 10)
export const portsCutoffAtom = settingAtom(cutoffPortsKey, 10)
export const forwardedPortsCutoffAtom = settingAtom(cutoffForwardedPortsKey, 10)
export const routesCutoffAtom = settingAtom(cutoffRoutesKey, 10)
export const nifsCutoffAtom = settingAtom(cutoffNifsKey, NEVER)

export const showSandboxesAtom = settingAtom(showSandboxesKey, false)

export const showNamespaceIdsAtom = settingAtom(showNamespaceIds, false)

export const snapshotDensityAtom = settingAtom(snapshotDensityKey, 1)

export const showIEAppIconsAtom = settingAtom(showIEAppIconsKey, false)

export const showMultiBroadcastRoutesAtom = settingAtom(showMultiBroadcastRoutesKey, false)

export const filterPatternAtom = settingAtom(filterPatternKey, '')
export const filterCaseSensitiveAtom = settingAtom(filterCaseSensitiveKey, false)
export const filterRegexpAtom = settingAtom(filterRegexpKey, false)
