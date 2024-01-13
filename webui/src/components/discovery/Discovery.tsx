// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import { useEffect } from 'react'

import { useAtom } from 'jotai'

import { useSnackbar } from 'notistack'

import useInterval from 'hooks/interval'

import { 
    discoveryRefreshingAtom, discoveryRefreshIntervalAtom, discoveryErrorAtom, 
    refreshIntervalKey,
} from './disco'

/**
 * The `Discovery` component is responsible for retrieving and updating
 * discovery information based on the refresh interval set (in milliseconds; can
 * also be `null` as opposed to `0`). On creation of this component it will
 * start an update automatically, unless the refresh interval is `null`.
 *
 * This component also handles posting discovery and update errors to a
 * snackbar. For this, it requires a
 * [notistack](https://github.com/iamhosseindhv/notistack) `SnackbarProvider` to
 * be present in the application.
 *
 * Please note that this component doesn't accept any children. Instead, just
 * place it somewhere where it can post notifications.
 *
 * This component is licensed under the [Apache License, Version
 * 2.0](http://www.apache.org/licenses/LICENSE-2.0).
 */
export const Discovery = () => {
    // In order to report discovery REST API failures... 
    const { enqueueSnackbar } = useSnackbar()

    // Discovery status and control...
    const [discoveryError] = useAtom(discoveryErrorAtom)
    const [interval] = useAtom(discoveryRefreshIntervalAtom)
    const [, setDiscoveryRefreshing] = useAtom(discoveryRefreshingAtom)

    // Get new discovery data after some time; please note that useInterval
    // interprets a null cycle as switching off the timer.
    useInterval(() => setDiscoveryRefreshing(true), interval)

    // Initially fetch discovery data, unless the cycle is null.
    useEffect(() => {
        if (interval !== null) {
            setDiscoveryRefreshing(true)
        }
        localStorage.setItem(refreshIntervalKey, JSON.stringify(interval))
    }, [interval])

    useEffect(
        () => {
            discoveryError && enqueueSnackbar(discoveryError, { variant: 'error' })
        },
        [discoveryError])

    // Do not render anything.
    return null
}
