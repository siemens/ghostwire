// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import { type ReactNode, useState } from 'react'

import { Alert, IconButton, Snackbar, styled, Tooltip } from '@mui/material'
import ClearIcon from '@mui/icons-material/Clear'
import { ContentCopy } from '@mui/icons-material'

import { type NetworkInterface } from 'models/gw'

// nif details are a grid (table) of key-value pairs; even if some values are
// arrays consisting of multiple values themselves, but we won't notice at this
// level.
const Details = styled('div')(({ theme }) => ({
    display: 'grid',
    gridTemplateColumns: 'minmax(auto, max-content) minmax(50%, auto)', // https://stackoverflow.com/a/62163763
    columnGap: theme.spacing(2),
    rowGap: theme.spacing(0),
}))

const Property = styled('div')(({ theme }) => ({
    gridColumn: '1 / 2',
    color: theme.palette.text.secondary,
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    whiteSpace: 'nowrap',
}))

const Value = styled('div')(() => ({
    gridColumn: '2 / 3',
    minHeight: '24px', // ensures consistent height when no icon in value.
    alignSelf: 'baseline',
    overflowWrap: 'break-word',

    '& > .MuiSvgIcon-root': { verticalAlign: 'middle' },
}))

export interface NifInfoProps {
    /** network interface to show information about. */
    nif: NetworkInterface
    /** optional CSS class name(s). */
    className?: string
}

/**
 * Renders (driver) information about a network interface as a property grid,
 * together with a button for copying this information to the clipboard.
 */
export const NifInfo = ({ nif, className }: NifInfoProps) => {
    const [snackbarOpen, setSnackbarOpen] = useState(false)

    // Render a single property with value row in the property grid.
    let row = 0
    const prop = (name: string, value: ReactNode) => {
        if (!value) {
            value = <ClearIcon fontSize="inherit" color="disabled" />
        }
        row++
        return [
            <Property key={`k${row}`} style={{ gridRow: `${row}/${row + 1}` }}>
                {name}
            </Property>,
            <Value key={`v${row}`} style={{ gridRow: `${row}/${row + 1}` }}>
                {value}
            </Value>
        ]
    }

    const handleCopyToClipboard = () => {
        let s = ''
        s = add(s, 'interface name', nif.name)
        s = add(s, 'type/kind', nif.kind || '(virtual) hardware')
        s = add(s, 'driver', nif.driverinfo.driver)
        s = add(s, 'firmware version', (nif.driverinfo.fwversion || '') !== 'N/A' ? nif.driverinfo.fwversion : '')
        s = add(s, 'ext ROM version', nif.driverinfo.eromversion)
        s = add(s, 'bus', nif.driverinfo.businfo)
        navigator.clipboard.writeText(s)
        setSnackbarOpen(true)
    }

    return (
        <div className={className}>
            <Details>
                {prop('interface name', nif.name)}
                {prop('type/kind', nif.kind ? `virtual ${nif.kind}` : '(virtualized) hardware')}
                {prop('driver', nif.driverinfo.driver)}
                {prop('firmware version', nif.driverinfo.fwversion !== 'N/A' && nif.driverinfo.fwversion)}
                {prop('ext ROM version', nif.driverinfo.eromversion)}
                {prop('bus', nif.driverinfo.businfo)}
            </Details>
            <Tooltip title="copy information to clipboard">
                <IconButton
                    style={{ float: 'right' }}
                    onClick={handleCopyToClipboard}
                ><ContentCopy /></IconButton>
            </Tooltip>
            <Snackbar
                open={snackbarOpen}
                onClose={() => setSnackbarOpen(false)}
                autoHideDuration={2000}
            >
                <Alert severity="info">copied to clipboard</Alert>
            </Snackbar>
        </div>
    )
}

const add = (s: string, key: string, value: string | undefined) => {
    if (!value) {
        return s
    }
    return s + key + ': ' + value + '\n'
}

export default NifInfo
