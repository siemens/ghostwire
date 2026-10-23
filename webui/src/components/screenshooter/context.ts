// (c) Siemens AG 2023
//
// SPDX-License-Identifier: MIT

import React, { useContext } from 'react'

// Changing the context to an HTML element to capture will automatically pop up
// the modal screenshooter dialog. It will automatically be reset to null after
// the modal dialog has been closed.
export const ScreenShooterModalContext = React.createContext<
    null | React.Dispatch<React.SetStateAction<HTMLElement|null>>>(null)

/**
 * Returns a setter that when setting true will pop up the screenshooter.
 */
export const useScreenShooterModal = () => {
    return useContext(ScreenShooterModalContext)
}
