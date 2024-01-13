// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useState } from 'react'

import { useAtom } from 'jotai'
import { Button, Dialog, DialogActions, DialogContent, DialogTitle, MenuItem, Select, SelectChangeEvent, styled, useTheme } from '@mui/material'
import { useContext } from 'react'
import { snapshotDensityAtom } from 'views/settings'
import DensityIcon from 'icons/Density'
import { toPng } from 'html-to-image'
import { useSnackbar } from 'notistack'


// Changing the context to an HTML element to capture will automatically pop up
// the modal screenshooter dialog. It will automatically be reset to null after
// the modal dialog has been closed.
const ScreenShooterModalContext = React.createContext<
    null | React.Dispatch<React.SetStateAction<HTMLElement|null>>>(null)


const Density = styled(DensityIcon)(({ theme }) => ({
    marginRight: theme.spacing(1),
}))

const Contents = styled(DialogContent)(({ theme }) => ({
    paddingTop: `${theme.spacing(1)} !important`, // override overtight padding to preceeding title.
}))


// Generates an at least slightly useful unique filename, given a base filename.
const generateFilename = (basename: string) => {
    const now = new Date()
    const d = now.getFullYear()
        + (now.getMonth() + 1).toString().padStart(2, '0')
        + now.getDay().toString().padStart(2, '0')
    const tod = now.getHours().toString().padStart(2, '0')
        + now.getMinutes().toString().padStart(2, '0')
        + now.getSeconds().toString().padStart(2, '0')
    return `${basename}-${d}-${tod}.png`
}

export interface ScreenShooterProps {
    /** children to render. */
    children: React.ReactNode
    /** file basename to use. Defaults to "snapshot". */
    basename?: string
}

/**
 *  `ScreenShooter` provides a screen shot modal dialog. The dialog is triggered
 *  by using the setter returned by `useScreenShooterModal()` to set an HTML
 *  element. This HTML element is then captured and a PNG image downloaded.
 *
 * Users can select the image "density", that is, capture at either 1:1, double
 * density, et cetera. For "double density" the screenshot image size will be
 * twice the width and height of the original HTML element.
 *
 * When an outer notistack SnackbarProvider is expected to be present, then
 * suckzess as well as errors will be reported accordingly to it.
 */
export const ScreenShooter = ({ basename, children }: ScreenShooterProps) => {
    const theme = useTheme()

    const [htmlElement, setHtmlElement] = useState<HTMLElement|null>(null)
    const [density, setDensity] = useAtom(snapshotDensityAtom)
    const safeDensity = [1, 2, 4].find(setting => Math.max(Math.floor(density), 1) <= setting)

    const { enqueueSnackbar } = useSnackbar() || {}

    basename = basename || 'screenshot'

    const handleChange = (event: SelectChangeEvent<number>) => {
        setDensity(event.target.value as number)
    }

    const handleDownload = () => {
        if (!htmlElement) {
            return
        }
        toPng(
            htmlElement,
            {
                backgroundColor: theme.palette.background.default,
                pixelRatio: density,
            }
        ).then(dataurl => {
            const link = document.createElement('a')
            link.download = generateFilename(basename!)
            link.href = dataurl
            link.click()
            enqueueSnackbar && enqueueSnackbar(
                'successfully captured image', {
                variant: 'success',
                autoHideDuration: 2000,
            })
        }).catch(err => {
            enqueueSnackbar
                ? enqueueSnackbar(err.toString(), { variant: 'error' })
                : console.error(err.toString())
        })
        setHtmlElement(null)
    }

    const handleClose = () => {
        setHtmlElement(null)
    }

    return (
        <ScreenShooterModalContext.Provider value={setHtmlElement}>
            {children}
            <Dialog
                open={!!htmlElement}
                onClose={handleClose}
            >
                <DialogTitle>
                    Download PNG screenshot?
                </DialogTitle>
                <Contents>
                    <Select
                        size="small"
                        value={safeDensity}
                        onChange={handleChange}
                        startAdornment={<Density />}
                    >
                        <MenuItem value={1}>single density</MenuItem>
                        <MenuItem value={2}>double density</MenuItem>
                        <MenuItem value={4}>4× density</MenuItem>
                    </Select>
                </Contents>
                <DialogActions>
                    <Button onClick={handleClose} color="primary">
                        Cancel
                    </Button>
                    <Button onClick={handleDownload} color="primary" autoFocus>
                        Download
                    </Button>
                </DialogActions>
            </Dialog>
        </ScreenShooterModalContext.Provider>
    )

}

export default ScreenShooter

/**
 * Returns a setter that when setting true will pop up the screenshooter.
 */
export const useScreenShooterModal = () => {
    return useContext(ScreenShooterModalContext)
}
