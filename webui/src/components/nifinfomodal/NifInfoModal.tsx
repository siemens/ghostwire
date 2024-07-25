// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { ReactNode, useContext, useState } from 'react'
import { styled } from '@mui/material'
import { NetworkInterface } from 'models/gw'
import { Alert, Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton, Snackbar, Tooltip } from '@mui/material'
import ClearIcon from '@mui/icons-material/Clear'
import { ContentCopy } from '@mui/icons-material'
import CloseIcon from '@mui/icons-material/Close'

import { NifIcon } from 'components/nificon'

const NifDialogTitle = styled(DialogTitle)(({ theme }) => ({
    '& .close': {
        position: 'relative',
        right: theme.spacing(-1),
        top: theme.spacing(-0.5),
    },
    '& .nificon.MuiSvgIcon-root': {
        position: 'relative',
        top: theme.spacing(0.5),
    },
}))

const CloseButton = styled(IconButton)(({ theme }) => ({
    position: 'absolute',
    right: theme.spacing(1),
    top: theme.spacing(1),
    color: theme.palette.grey[500],
}))

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

const Contents = styled(DialogContent)(({ theme }) => ({
    marginLeft: theme.spacing(2),
    marginRight: theme.spacing(2),
    paddingLeft: 0,
    paddingRight: 0,
    fontFamily: theme.typography.fontFamily,
    fontSize: theme.typography.body1.fontSize,
}))


const NifInfoModalContext = React.createContext<
    undefined | React.Dispatch<React.SetStateAction<NetworkInterface | undefined>>>(undefined)

export interface NifInfoModalProviderProps {
    children: React.ReactNode
}

export const NifInfoModalProvider = ({ children }: NifInfoModalProviderProps) => {
    const [nif, setNif] = useState<NetworkInterface>()
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

    const handleClose = () => {
        setNif(undefined)
    }

    const handleCopyToClipboard = () => {
        let s = ''
        s = add(s, 'interface name', nif?.name)
        s = add(s, 'type/kind', nif?.kind || '(virtual) hardware')
        s = add(s, 'driver', nif?.driverinfo.driver)
        s = add(s, 'firmware version', (nif && nif.driverinfo.fwversion || '') !== 'N/A' ? nif?.driverinfo.fwversion : '')
        s = add(s, 'ext ROM version', nif?.driverinfo.eromversion)
        s = add(s, 'bus', nif?.driverinfo.businfo)
        navigator.clipboard.writeText(s)
        setSnackbarOpen(true)
    }

    return (
        <NifInfoModalContext.Provider value={setNif}>
            {children}
            {nif && <Dialog
                scroll="paper"
                open={!!nif}
                onClose={handleClose}
            >
                <NifDialogTitle>
                    <NifIcon
                        nif={nif}
                        considerPhysical
                        fontSize="medium"
                        className="nificon"
                    />&nbsp;Network Interface Information
                    <CloseButton
                        className="close"
                        aria-label="close"
                        onClick={handleClose}
                        size="large">
                        <CloseIcon />
                    </CloseButton>
                </NifDialogTitle>
                <Contents dividers>
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
                </Contents>
                <DialogActions>
                    <Button autoFocus onClick={handleClose}>Close</Button>
                </DialogActions>
                <Snackbar
                    open={snackbarOpen}
                    onClose={() => setSnackbarOpen(false)}
                    autoHideDuration={2000}
                    security="info"
                >
                    <Alert security="info">copied to clipboard</Alert>
                </Snackbar>
            </Dialog>}
        </NifInfoModalContext.Provider>
    )
}

const add = (s: string, key: string, value: string | undefined) => {
    if (!value) {
        return s
    }
    return s + key + ': ' + value + '\n'
}

export default NifInfoModalProvider

/**
 * Returns a setter to specify the Nif to show information about in a modal
 * dialog.
 */
export const useNifInfoModal = () => {
    return useContext(NifInfoModalContext)
}
