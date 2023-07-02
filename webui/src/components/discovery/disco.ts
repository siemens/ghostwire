// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { atom, useAtom, Setter } from 'jotai'

import { Discovery as DiscoveryResult, fromjson } from 'models/gw'
import { showIEAppIconsAtom } from 'views/settings'

/** Internal discovery result state; can be used only via useDiscovery(). */
const discoveryResultAtom = atom({
    networkNamespaces: {}
} as DiscoveryResult)

/**
 * Internal discovery error state; internally used to display a snackbar
 * message via the Discovery component. 
 */
export const discoveryErrorAtom = atom("")

export const refreshIntervalKey = "gw.refresh.interval"

const initialRefreshInterval = (() => {
    try {
        const interval = JSON.parse(localStorage.getItem(refreshIntervalKey));
        if (interval === null || (Number.isInteger(interval) && interval > 500)) {
            return interval
        }
    } catch (e) { }
    return 5000;
})()

/** 
 * The discovery refresh interval state; `null` means the refresh is disabled.
 * The refresh interval is automatically synced to the local storage.
 */
export const discoveryRefreshIntervalAtom = atom(
    initialRefreshInterval,
    (_get, set, interval: number) => {
        set(discoveryRefreshIntervalAtom, interval)
        localStorage.setItem(refreshIntervalKey, JSON.stringify(interval))
    }
)

/** 
 * Discovery refresh status; setting the status to "true" triggers an ad-hoc
 * refresh, unless there is already a refresh ongoing. It is not possible to
 * reset an ongoing refresh.
 */
export const discoveryRefreshingAtom = atom(
    false,
    (get, set, arg) => {
        const refreshing = get(discoveryRefreshingAtom)
        if (arg as boolean && !refreshing) {
            const discoverIEAppIcons = get(showIEAppIconsAtom)
            set(discoveryRefreshingAtom, true)
            fetchDiscoveryData(set, discoverIEAppIcons ? 'ieappicons' : '')
        }
    }
)

/** 
 * Use the namespace discovery result in a react component; on purpose, there no
 * way to set it (it wouldn't make sense).
 */
export const useDiscovery = () => {
    const [discovery] = useAtom(discoveryResultAtom)
    return discovery
}

// Fetch the namespace+process discovery data from the server, postprocess
// the JSON result, and finally update the discovery data state with the new
// information about all namespaces, adding information about the previous
// discovery state.
const fetchDiscoveryData = (set: Setter, query: string) => {
    fetch('json' + (query ? '?' + query : '')) // relies on base href!
        .then(httpresult => {
            // Nota bene: we keep the refreshing state still "on" while we're
            // processing the result and will only switch it to "off" after
            // we've converted the result into our internal representation and
            // just right before we set the result.

            // fetch() doesn't throw an error for non-2xx reponse status codes, so
            // we throw up here instead, so we can catch below later in the promise
            // chain.
            if (!httpresult.ok) {
                throw Error(httpresult.status + " " + httpresult.statusText)
            }
            try {
                return httpresult.json()
            } catch (e) {
                throw Error('malformed discovery API response')
            }
        })
        .then(jsondata => fromjson(jsondata))
        .then(discovery => {
            set(discoveryRefreshingAtom, false)
            set(discoveryResultAtom, discovery)
        })
        .catch((error) => {
            // Don't forget to reset the refreshing indication and then set the
            // error result, so someone else can pick it up and send a toast to the
            // snackbar. Before 10pm. And only less than six toasts. Just for
            // testing eyesight.
            set(discoveryRefreshingAtom, false)
            // Ugly hack to ensure that another error with exactly the same
            // error message as the previous one will be shown and not optimized
            // away, as set will otherwise do: reset the message and set it to
            // the "new" value. This will cause only a single message to be
            // shown.
            set(discoveryErrorAtom, '')
            set(discoveryErrorAtom,
                'refreshing failed: ' + error.toString().replace(/^[E|e]rror: /, ''))
            console.log(error)
        })
}
