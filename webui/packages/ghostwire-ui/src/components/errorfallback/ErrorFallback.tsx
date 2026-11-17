// (c) Siemens AG 2026
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'

import { type FallbackProps } from 'react-error-boundary'

import {
    Alert,
    AlertTitle,
    Box,
    Button,
    Dialog,
    DialogActions,
    DialogContent,
    DialogContentText,
    DialogTitle,
    type SxProps,
    type Theme,
} from '@mui/material'

import { errorMessage, reloadPage } from '../../utils/errors'


export interface ErrorNoticeProps {
    /** short description of what failed. */
    title: React.ReactNode
    /** the thrown value, if any; its message is shown below the title. */
    error?: unknown
    /** if set, shows a "Retry" button calling this handler. */
    onRetry?: () => void
    /** if true, shows a "Reload page" button. */
    reload?: boolean
    /** CSS class name(s). */
    className?: string
    sx?: SxProps<Theme>
}

/**
 * `ErrorNotice` renders an inline error notice, for instance as the fallback
 * of an error boundary. It optionally offers to retry (rendering) and/or to
 * reload the whole page.
 */
export const ErrorNotice = ({ title, error, onRetry, reload, className, sx }: ErrorNoticeProps) => (
    <Alert
        severity="error"
        className={className}
        sx={sx}
        action={(onRetry || reload) && <>
            {onRetry && <Button color="inherit" size="small" onClick={onRetry}>Retry</Button>}
            {reload && <Button color="inherit" size="small" onClick={reloadPage}>Reload page</Button>}
        </>}
    >
        <AlertTitle>{title}</AlertTitle>
        {error !== undefined && errorMessage(error)}
    </Alert>
)

/**
 * `AppErrorFallback` is the last-resort error boundary fallback for the whole
 * application, offering only to reload the page.
 */
export const AppErrorFallback = ({ error }: FallbackProps) => (
    <Box sx={{ p: 2 }}>
        <ErrorNotice title="The user interface ran into a problem" error={error} reload />
    </Box>
)

/**
 * `LoadErrorFallback` is an error boundary fallback for lazily loaded
 * content: it shows an inline error notice and additionally pops up a dialog
 * offering to retry or to reload the page. A retry might succeed after a
 * temporary loss of connection, but after a service update only reloading the
 * page fetches the now current web UI resources.
 */
export const LoadErrorFallback = ({ error, resetErrorBoundary }: FallbackProps) => {
    const [open, setOpen] = useState(true)

    return <>
        <ErrorNotice
            sx={{ m: 2 }}
            title="This content couldn't be loaded"
            error={error}
            onRetry={resetErrorBoundary}
            reload
        />
        <Dialog open={open} onClose={() => setOpen(false)}>
            <DialogTitle>Content couldn't be loaded</DialogTitle>
            <DialogContent>
                <DialogContentText>
                    Either the connection to the service has been interrupted or
                    the service has been updated in the meantime. Reloading the
                    page fetches the most recent version of the user interface.
                </DialogContentText>
            </DialogContent>
            <DialogActions>
                <Button onClick={() => setOpen(false)}>Close</Button>
                <Button onClick={() => resetErrorBoundary()}>Retry</Button>
                <Button variant="contained" onClick={reloadPage}>Reload page</Button>
            </DialogActions>
        </Dialog>
    </>
}
