// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import { useEffect } from 'react'

import { useSnackbar } from 'notistack'

import { errorMessage } from '../../utils/errors'


// Browsers report a ResizeObserver not being able to deliver all notifications
// within a single animation frame as an error, yet this is harmless.
const isHarmless = (message: string) => /^ResizeObserver loop/.test(message)

/**
 * The `UnhandledErrors` component posts errors that would otherwise go
 * unnoticed to a snackbar: these are exceptions thrown outside rendering, such
 * as in event handlers (clicks, websocket messages, ...), as well as unhandled
 * promise rejections. Errors thrown while rendering are instead caught by
 * error boundaries.
 *
 * This component requires a
 * [notistack](https://github.com/iamhosseindhv/notistack) `SnackbarProvider`
 * to be present in the application; it doesn't render anything itself.
 */
export const UnhandledErrors = () => {
    const { enqueueSnackbar } = useSnackbar()

    useEffect(() => {
        const notify = (error: unknown) => enqueueSnackbar(
            'unexpected error: ' + errorMessage(error),
            { variant: 'error', preventDuplicate: true })

        const handleError = (event: ErrorEvent) => {
            if (!isHarmless(event.message)) {
                notify(event.error ?? event.message)
            }
        }
        const handleRejection = (event: PromiseRejectionEvent) => notify(event.reason)

        window.addEventListener('error', handleError)
        window.addEventListener('unhandledrejection', handleRejection)
        return () => {
            window.removeEventListener('error', handleError)
            window.removeEventListener('unhandledrejection', handleRejection)
        }
    }, [enqueueSnackbar])

    return null
}
