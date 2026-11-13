// (c) Siemens AG 2024
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'
import { styled } from '@mui/material'
import { type NetworkInterface } from 'models/gw'
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, IconButton } from '@mui/material'
import CloseIcon from '@mui/icons-material/Close'

import { NifIcon } from 'components/nificon'
import { NifInfoModalContext } from './context'
import { NifInfo } from './NifInfo'

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

const Contents = styled(DialogContent)(({ theme }) => ({
    marginLeft: theme.spacing(2),
    marginRight: theme.spacing(2),
    paddingLeft: 0,
    paddingRight: 0,
    fontFamily: theme.typography.fontFamily,
    fontSize: theme.typography.body1.fontSize,
}))


export interface NifInfoModalProviderProps {
    children: React.ReactNode
}

export const NifInfoModalProvider = ({ children }: NifInfoModalProviderProps) => {
    const [nif, setNif] = useState<NetworkInterface>()

    const handleClose = () => {
        setNif(undefined)
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
                    <NifInfo nif={nif} />
                </Contents>
                <DialogActions>
                    <Button autoFocus onClick={handleClose}>Close</Button>
                </DialogActions>
            </Dialog>}
        </NifInfoModalContext.Provider>
    )
}

export default NifInfoModalProvider
