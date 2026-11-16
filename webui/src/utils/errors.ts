// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { getErrorMessage } from 'react-error-boundary'

/**
 * Returns a human-readable message for an arbitrary thrown value, which might
 * not even be an `Error`.
 */
export const errorMessage = (error: unknown) =>
    getErrorMessage(error) || String(error)

/** Reloads the whole page, fetching the current version of the web UI. */
export const reloadPage = () => window.location.reload()
